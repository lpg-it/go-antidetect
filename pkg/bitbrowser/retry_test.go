package bitbrowser

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 1 {
		t.Errorf("MaxAttempts = %d, want 1", config.MaxAttempts)
	}
	if config.BaseDelay != 1*time.Second {
		t.Errorf("BaseDelay = %v, want 1s", config.BaseDelay)
	}
	if config.MaxDelay != 30*time.Second {
		t.Errorf("MaxDelay = %v, want 30s", config.MaxDelay)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("Multiplier = %v, want 2.0", config.Multiplier)
	}
	if config.Jitter != 0.1 {
		t.Errorf("Jitter = %v, want 0.1", config.Jitter)
	}
}

func TestRetryer_NoRetry(t *testing.T) {
	config := &RetryConfig{
		MaxAttempts: 1,
	}
	r := newRetryer(config)

	attempts := 0
	err := r.do(context.Background(), func() error {
		attempts++
		return errors.New("always fail")
	})

	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
	if err == nil {
		t.Error("expected error, got nil")
	}
	// With MaxAttempts=1, should return the original error, not RetryError
	if errors.Is(err, ErrRetryExhausted) {
		t.Error("single attempt should not return RetryError")
	}
}

func TestRetryer_SuccessOnFirstAttempt(t *testing.T) {
	config := &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
	}
	r := newRetryer(config)

	attempts := 0
	err := r.do(context.Background(), func() error {
		attempts++
		return nil
	})

	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRetryer_SuccessAfterRetry(t *testing.T) {
	config := &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		Multiplier:  1.0, // No backoff to speed up test
		Jitter:      0,
	}
	r := newRetryer(config)

	attempts := 0
	err := r.do(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return &NetworkError{Op: "test", URL: "http://test", Err: errors.New("transient")}
		}
		return nil
	})

	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRetryer_ExhaustsAllAttempts(t *testing.T) {
	config := &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		Multiplier:  1.0,
		Jitter:      0,
	}
	r := newRetryer(config)

	attempts := 0
	err := r.do(context.Background(), func() error {
		attempts++
		return &NetworkError{Op: "test", URL: "http://test", Err: errors.New("always fail")}
	})

	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
	if !errors.Is(err, ErrRetryExhausted) {
		t.Errorf("expected ErrRetryExhausted, got %T", err)
	}

	var retryErr *RetryError
	if !errors.As(err, &retryErr) {
		t.Error("expected *RetryError")
	} else if retryErr.Attempts != 3 {
		t.Errorf("retryErr.Attempts = %d, want 3", retryErr.Attempts)
	}
}

func TestRetryer_NonRetryableErrorStopsImmediately(t *testing.T) {
	config := &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
	}
	r := newRetryer(config)

	attempts := 0
	validationErr := &ValidationError{Field: "id", Message: "required"}

	err := r.do(context.Background(), func() error {
		attempts++
		return validationErr
	})

	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (should stop on non-retryable error)", attempts)
	}
	if err != validationErr {
		t.Errorf("expected original validation error to be returned")
	}
}

func TestRetryer_ContextCancellation(t *testing.T) {
	config := &RetryConfig{
		MaxAttempts: 100, // Large number to ensure we don't exhaust naturally
		BaseDelay:   100 * time.Millisecond,
		Multiplier:  1.0,
		Jitter:      0,
		RetryIf:     func(err error) bool { return true }, // Always retry
	}
	r := newRetryer(config)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	var attempts int32

	err := r.do(ctx, func() error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("always fail")
	})

	// Should have stopped before 100 attempts due to context timeout
	finalAttempts := atomic.LoadInt32(&attempts)
	if finalAttempts >= 100 {
		t.Errorf("should have stopped before all attempts due to context, got %d", finalAttempts)
	}

	// Should have made at least 2 attempts (immediate + after first delay)
	if finalAttempts < 2 {
		t.Errorf("should have made at least 2 attempts, got %d", finalAttempts)
	}

	// Should return RetryError or context error
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestRetryer_ContextAlreadyCancelled(t *testing.T) {
	config := &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
	}
	r := newRetryer(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before starting

	attempts := 0
	err := r.do(ctx, func() error {
		attempts++
		return nil
	})

	if attempts != 0 {
		t.Errorf("attempts = %d, want 0 (should check context before first attempt)", attempts)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %T", err)
	}
}

func TestRetryer_CustomRetryIf(t *testing.T) {
	customErr := errors.New("custom retryable error")
	config := &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		Multiplier:  1.0,
		Jitter:      0,
		RetryIf: func(err error) bool {
			return err == customErr
		},
	}
	r := newRetryer(config)

	t.Run("retries custom error", func(t *testing.T) {
		attempts := 0
		_ = r.do(context.Background(), func() error {
			attempts++
			return customErr
		})

		if attempts != 3 {
			t.Errorf("attempts = %d, want 3", attempts)
		}
	})

	t.Run("does not retry non-matching error", func(t *testing.T) {
		attempts := 0
		_ = r.do(context.Background(), func() error {
			attempts++
			return errors.New("different error")
		})

		if attempts != 1 {
			t.Errorf("attempts = %d, want 1", attempts)
		}
	})
}

func TestRetryer_CalculateDelay(t *testing.T) {
	config := &RetryConfig{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   1 * time.Second,
		Multiplier: 2.0,
		Jitter:     0, // No jitter for predictable tests
	}
	r := newRetryer(config)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 100 * time.Millisecond},  // 100ms * 2^0
		{2, 200 * time.Millisecond},  // 100ms * 2^1
		{3, 400 * time.Millisecond},  // 100ms * 2^2
		{4, 800 * time.Millisecond},  // 100ms * 2^3
		{5, 1000 * time.Millisecond}, // Capped at MaxDelay
		{6, 1000 * time.Millisecond}, // Still capped
	}

	for _, tt := range tests {
		delay := r.calculateDelay(tt.attempt)
		if delay != tt.expected {
			t.Errorf("calculateDelay(%d) = %v, want %v", tt.attempt, delay, tt.expected)
		}
	}
}

func TestRetryer_CalculateDelayWithJitter(t *testing.T) {
	config := &RetryConfig{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   1 * time.Second,
		Multiplier: 2.0,
		Jitter:     0.5, // 50% jitter
	}
	r := newRetryer(config)

	// With 50% jitter, delay should be in range [50ms, 150ms] for first attempt
	minExpected := 50 * time.Millisecond
	maxExpected := 150 * time.Millisecond

	// Run multiple times to verify jitter is applied
	for i := 0; i < 10; i++ {
		delay := r.calculateDelay(1)
		if delay < minExpected || delay > maxExpected {
			t.Errorf("calculateDelay(1) = %v, want between %v and %v", delay, minExpected, maxExpected)
		}
	}
}

func TestRetryer_NilConfig(t *testing.T) {
	r := newRetryer(nil)

	// Should use default config
	if r.config.MaxAttempts != 1 {
		t.Errorf("MaxAttempts = %d, want 1", r.config.MaxAttempts)
	}
}

func TestRetryer_ZeroMaxAttempts(t *testing.T) {
	config := &RetryConfig{
		MaxAttempts: 0,
		BaseDelay:   10 * time.Millisecond,
	}
	r := newRetryer(config)

	attempts := 0
	err := r.do(context.Background(), func() error {
		attempts++
		return nil
	})

	// Should treat 0 as 1
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRetryer_ActualBackoffTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timing test in short mode")
	}

	config := &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   50 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		Multiplier:  2.0,
		Jitter:      0,
		RetryIf:     func(err error) bool { return true }, // Always retry
	}
	r := newRetryer(config)

	start := time.Now()
	attempts := 0

	_ = r.do(context.Background(), func() error {
		attempts++
		return errors.New("always fail")
	})

	elapsed := time.Since(start)

	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}

	// Expected: attempt 1 (immediate) + 50ms delay + attempt 2 + 100ms delay + attempt 3
	// Total delay: ~150ms (allowing some margin)
	expectedMin := 140 * time.Millisecond
	expectedMax := 250 * time.Millisecond

	if elapsed < expectedMin || elapsed > expectedMax {
		t.Errorf("elapsed = %v, want between %v and %v", elapsed, expectedMin, expectedMax)
	}
}
