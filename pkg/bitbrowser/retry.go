package bitbrowser

import (
	"context"
	"math"
	"math/rand/v2"
	"time"
)

// RetryConfig configures the retry behavior.
type RetryConfig struct {
	// MaxAttempts is the maximum number of attempts (including the initial attempt).
	// Setting to 1 means no retries. Default is 1.
	MaxAttempts int

	// BaseDelay is the initial delay before the first retry.
	// Default is 1 second.
	BaseDelay time.Duration

	// MaxDelay is the maximum delay between retries.
	// Default is 30 seconds.
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases after each retry.
	// Default is 2.0 (exponential backoff).
	Multiplier float64

	// Jitter adds randomness to the delay to prevent thundering herd.
	// Value between 0 and 1, where 0 means no jitter and 1 means up to 100% jitter.
	// Default is 0.1 (10% jitter).
	Jitter float64

	// RetryIf is an optional function to determine if an error is retryable.
	// If nil, the default IsRetryable function is used.
	RetryIf func(error) bool
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults.
// By default, MaxAttempts is 1 (no retries) for backward compatibility.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts: 1,
		BaseDelay:   1 * time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
		Jitter:      0.1,
		RetryIf:     nil,
	}
}

// retryer handles retry logic for operations.
type retryer struct {
	config *RetryConfig
}

// newRetryer creates a new retryer with the given configuration.
func newRetryer(config *RetryConfig) *retryer {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &retryer{config: config}
}

// do executes the given function with retry logic.
// It respects context cancellation and returns early if the context is done.
func (r *retryer) do(ctx context.Context, fn func() error) error {
	if r.config.MaxAttempts <= 0 {
		r.config.MaxAttempts = 1
	}

	retryIf := r.config.RetryIf
	if retryIf == nil {
		retryIf = IsRetryable
	}

	var lastErr error
	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		// Check context before each attempt
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return NewRetryError(attempt-1, lastErr)
			}
			return err
		}

		// Execute the function
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Check if we should retry
		if attempt >= r.config.MaxAttempts {
			break
		}

		if !retryIf(lastErr) {
			return lastErr
		}

		// Calculate delay with exponential backoff
		delay := r.calculateDelay(attempt)

		// Wait with context awareness
		select {
		case <-ctx.Done():
			return NewRetryError(attempt, lastErr)
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	// All attempts exhausted
	if r.config.MaxAttempts > 1 {
		return NewRetryError(r.config.MaxAttempts, lastErr)
	}
	return lastErr
}

// calculateDelay computes the delay for the given attempt number.
// attempt is 1-indexed (first attempt is 1).
func (r *retryer) calculateDelay(attempt int) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}

	// Exponential backoff: baseDelay * multiplier^(attempt-1)
	delay := float64(r.config.BaseDelay) * math.Pow(r.config.Multiplier, float64(attempt-1))

	// Apply maximum delay cap (only if MaxDelay is set)
	if r.config.MaxDelay > 0 && delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	// Apply jitter: random value between [delay * (1 - jitter), delay * (1 + jitter)]
	if r.config.Jitter > 0 {
		jitterRange := delay * r.config.Jitter
		delay = delay - jitterRange + (rand.Float64() * 2 * jitterRange)
	}

	return time.Duration(delay)
}
