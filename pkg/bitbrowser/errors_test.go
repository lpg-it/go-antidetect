package bitbrowser

import (
	"errors"
	"net/http"
	"testing"
)

func TestNetworkError(t *testing.T) {
	t.Run("Error message with operation", func(t *testing.T) {
		err := &NetworkError{
			Op:  "connect",
			URL: "http://localhost:54345",
			Err: errors.New("connection refused"),
		}

		expected := "bitbrowser: network error during connect to http://localhost:54345: connection refused"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Error message without operation", func(t *testing.T) {
		err := &NetworkError{
			URL: "http://localhost:54345",
			Err: errors.New("connection refused"),
		}

		expected := "bitbrowser: network error to http://localhost:54345: connection refused"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("connection refused")
		err := &NetworkError{
			Op:  "connect",
			URL: "http://localhost:54345",
			Err: underlying,
		}

		if err.Unwrap() != underlying {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), underlying)
		}
	})

	t.Run("Is returns true for ErrNetwork", func(t *testing.T) {
		err := &NetworkError{
			Op:  "connect",
			URL: "http://localhost:54345",
			Err: errors.New("connection refused"),
		}

		if !errors.Is(err, ErrNetwork) {
			t.Error("errors.Is(err, ErrNetwork) should be true")
		}
	})

	t.Run("errors.As works correctly", func(t *testing.T) {
		err := &NetworkError{
			Op:  "connect",
			URL: "http://localhost:54345",
			Err: errors.New("connection refused"),
		}

		var netErr *NetworkError
		if !errors.As(err, &netErr) {
			t.Error("errors.As should work with *NetworkError")
		}

		if netErr.Op != "connect" {
			t.Errorf("netErr.Op = %q, want %q", netErr.Op, "connect")
		}
	})
}

func TestAPIError(t *testing.T) {
	t.Run("Error message with status code", func(t *testing.T) {
		err := &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "internal server error",
			Endpoint:   "/browser/open",
		}

		expected := "bitbrowser: API error on /browser/open (status 500): internal server error"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Error message without status code", func(t *testing.T) {
		err := &APIError{
			StatusCode: 0,
			Message:    "profile not found",
			Endpoint:   "/browser/detail",
		}

		expected := "bitbrowser: API error on /browser/detail: profile not found"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Is returns true for ErrAPI", func(t *testing.T) {
		err := &APIError{
			StatusCode: http.StatusNotFound,
			Message:    "not found",
			Endpoint:   "/browser/detail",
		}

		if !errors.Is(err, ErrAPI) {
			t.Error("errors.Is(err, ErrAPI) should be true")
		}
	})

	t.Run("errors.As works correctly", func(t *testing.T) {
		err := &APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid request",
			Endpoint:   "/browser/update",
		}

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Error("errors.As should work with *APIError")
		}

		if apiErr.StatusCode != http.StatusBadRequest {
			t.Errorf("apiErr.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusBadRequest)
		}
	})
}

func TestValidationError(t *testing.T) {
	t.Run("Error message with field", func(t *testing.T) {
		err := &ValidationError{
			Field:   "id",
			Message: "is required",
		}

		expected := `bitbrowser: validation error on field "id": is required`
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Error message without field", func(t *testing.T) {
		err := &ValidationError{
			Message: "invalid input",
		}

		expected := "bitbrowser: validation error: invalid input"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Is returns true for ErrValidation", func(t *testing.T) {
		err := &ValidationError{
			Field:   "id",
			Message: "is required",
		}

		if !errors.Is(err, ErrValidation) {
			t.Error("errors.Is(err, ErrValidation) should be true")
		}
	})

	t.Run("Unwrap returns nil", func(t *testing.T) {
		err := &ValidationError{
			Field:   "id",
			Message: "is required",
		}

		if err.Unwrap() != nil {
			t.Error("Unwrap() should return nil")
		}
	})
}

func TestTimeoutError(t *testing.T) {
	t.Run("Error message with duration", func(t *testing.T) {
		err := &TimeoutError{
			Op:       "http_request",
			Duration: "30s",
			Err:      errors.New("context deadline exceeded"),
		}

		expected := "bitbrowser: timeout during http_request after 30s"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Error message without duration", func(t *testing.T) {
		err := &TimeoutError{
			Op:  "http_request",
			Err: errors.New("context deadline exceeded"),
		}

		expected := "bitbrowser: timeout during http_request"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Is returns true for ErrTimeout", func(t *testing.T) {
		err := &TimeoutError{
			Op:  "http_request",
			Err: errors.New("context deadline exceeded"),
		}

		if !errors.Is(err, ErrTimeout) {
			t.Error("errors.Is(err, ErrTimeout) should be true")
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("context deadline exceeded")
		err := &TimeoutError{
			Op:  "http_request",
			Err: underlying,
		}

		if err.Unwrap() != underlying {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), underlying)
		}
	})
}

func TestRetryError(t *testing.T) {
	t.Run("Error message", func(t *testing.T) {
		err := &RetryError{
			Attempts: 3,
			LastErr:  errors.New("connection refused"),
		}

		expected := "bitbrowser: retry exhausted after 3 attempts: connection refused"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("Is returns true for ErrRetryExhausted", func(t *testing.T) {
		err := &RetryError{
			Attempts: 3,
			LastErr:  errors.New("connection refused"),
		}

		if !errors.Is(err, ErrRetryExhausted) {
			t.Error("errors.Is(err, ErrRetryExhausted) should be true")
		}
	})

	t.Run("Unwrap returns last error", func(t *testing.T) {
		lastErr := errors.New("connection refused")
		err := &RetryError{
			Attempts: 3,
			LastErr:  lastErr,
		}

		if err.Unwrap() != lastErr {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), lastErr)
		}
	})

	t.Run("Can unwrap to underlying error type", func(t *testing.T) {
		netErr := &NetworkError{
			Op:  "connect",
			URL: "http://localhost:54345",
			Err: errors.New("connection refused"),
		}
		err := &RetryError{
			Attempts: 3,
			LastErr:  netErr,
		}

		// Should be able to check for ErrNetwork through the chain
		if !errors.Is(err, ErrNetwork) {
			t.Error("errors.Is(err, ErrNetwork) should be true through chain")
		}

		// Should be able to extract NetworkError
		var extractedNetErr *NetworkError
		if !errors.As(err, &extractedNetErr) {
			t.Error("errors.As should extract *NetworkError through chain")
		}
	})
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error is not retryable",
			err:      nil,
			expected: false,
		},
		{
			name: "network error is retryable",
			err: &NetworkError{
				Op:  "connect",
				URL: "http://localhost:54345",
				Err: errors.New("connection refused"),
			},
			expected: true,
		},
		{
			name: "timeout error is retryable",
			err: &TimeoutError{
				Op:  "http_request",
				Err: errors.New("deadline exceeded"),
			},
			expected: true,
		},
		{
			name: "validation error is not retryable",
			err: &ValidationError{
				Field:   "id",
				Message: "is required",
			},
			expected: false,
		},
		{
			name: "retry exhausted error is not retryable",
			err: &RetryError{
				Attempts: 3,
				LastErr:  errors.New("connection refused"),
			},
			expected: false,
		},
		{
			name: "API 500 error is retryable",
			err: &APIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "internal server error",
				Endpoint:   "/browser/open",
			},
			expected: true,
		},
		{
			name: "API 502 error is retryable",
			err: &APIError{
				StatusCode: http.StatusBadGateway,
				Message:    "bad gateway",
				Endpoint:   "/browser/open",
			},
			expected: true,
		},
		{
			name: "API 503 error is retryable",
			err: &APIError{
				StatusCode: http.StatusServiceUnavailable,
				Message:    "service unavailable",
				Endpoint:   "/browser/open",
			},
			expected: true,
		},
		{
			name: "API 504 error is retryable",
			err: &APIError{
				StatusCode: http.StatusGatewayTimeout,
				Message:    "gateway timeout",
				Endpoint:   "/browser/open",
			},
			expected: true,
		},
		{
			name: "API 429 error is retryable",
			err: &APIError{
				StatusCode: http.StatusTooManyRequests,
				Message:    "too many requests",
				Endpoint:   "/browser/open",
			},
			expected: true,
		},
		{
			name: "API 400 error is not retryable",
			err: &APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "bad request",
				Endpoint:   "/browser/open",
			},
			expected: false,
		},
		{
			name: "API 404 error is not retryable",
			err: &APIError{
				StatusCode: http.StatusNotFound,
				Message:    "not found",
				Endpoint:   "/browser/detail",
			},
			expected: false,
		},
		{
			name: "API error without status code is not retryable",
			err: &APIError{
				StatusCode: 0,
				Message:    "profile not found",
				Endpoint:   "/browser/detail",
			},
			expected: false,
		},
		{
			name:     "generic error is not retryable",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryable(tt.err)
			if got != tt.expected {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewNetworkError", func(t *testing.T) {
		err := NewNetworkError("connect", "http://localhost", errors.New("refused"))
		if err.Op != "connect" {
			t.Errorf("Op = %q, want %q", err.Op, "connect")
		}
		if err.URL != "http://localhost" {
			t.Errorf("URL = %q, want %q", err.URL, "http://localhost")
		}
	})

	t.Run("NewAPIError", func(t *testing.T) {
		err := NewAPIError("/browser/open", 500, "server error")
		if err.Endpoint != "/browser/open" {
			t.Errorf("Endpoint = %q, want %q", err.Endpoint, "/browser/open")
		}
		if err.StatusCode != 500 {
			t.Errorf("StatusCode = %d, want %d", err.StatusCode, 500)
		}
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("id", "is required")
		if err.Field != "id" {
			t.Errorf("Field = %q, want %q", err.Field, "id")
		}
		if err.Message != "is required" {
			t.Errorf("Message = %q, want %q", err.Message, "is required")
		}
	})

	t.Run("NewTimeoutError", func(t *testing.T) {
		err := NewTimeoutError("request", "30s", errors.New("deadline"))
		if err.Op != "request" {
			t.Errorf("Op = %q, want %q", err.Op, "request")
		}
		if err.Duration != "30s" {
			t.Errorf("Duration = %q, want %q", err.Duration, "30s")
		}
	})

	t.Run("NewRetryError", func(t *testing.T) {
		err := NewRetryError(3, errors.New("failed"))
		if err.Attempts != 3 {
			t.Errorf("Attempts = %d, want %d", err.Attempts, 3)
		}
	})
}
