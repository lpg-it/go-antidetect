package bitbrowser

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinel errors for error type checking using errors.Is().
var (
	// ErrNetwork indicates a network-level error (connection, DNS, etc.).
	ErrNetwork = errors.New("network error")

	// ErrAPI indicates an API-level error (non-2xx response, business logic error).
	ErrAPI = errors.New("API error")

	// ErrValidation indicates a validation error (invalid input).
	ErrValidation = errors.New("validation error")

	// ErrTimeout indicates a timeout error.
	ErrTimeout = errors.New("timeout error")

	// ErrRetryExhausted indicates all retry attempts have been exhausted.
	ErrRetryExhausted = errors.New("retry exhausted")
)

// NetworkError represents a network-level error.
type NetworkError struct {
	Op      string // Operation that failed (e.g., "connect", "read", "write")
	URL     string // URL that was being accessed
	Err     error  // Underlying error
}

func (e *NetworkError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("bitbrowser: network error during %s to %s: %v", e.Op, e.URL, e.Err)
	}
	return fmt.Sprintf("bitbrowser: network error to %s: %v", e.URL, e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

func (e *NetworkError) Is(target error) bool {
	return target == ErrNetwork
}

// APIError represents an API-level error from BitBrowser.
type APIError struct {
	StatusCode int    // HTTP status code (0 if not applicable)
	Message    string // Error message from API
	Endpoint   string // API endpoint that was called
	Err        error  // Underlying error (if any)
}

func (e *APIError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("bitbrowser: API error on %s (status %d): %s", e.Endpoint, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("bitbrowser: API error on %s: %s", e.Endpoint, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func (e *APIError) Is(target error) bool {
	return target == ErrAPI
}

// ValidationError represents an input validation error.
type ValidationError struct {
	Field   string // Field that failed validation
	Message string // Validation error message
	Value   any    // The invalid value (optional)
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("bitbrowser: validation error on field %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("bitbrowser: validation error: %s", e.Message)
}

func (e *ValidationError) Unwrap() error {
	return nil
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrValidation
}

// TimeoutError represents a timeout error.
type TimeoutError struct {
	Op       string // Operation that timed out
	Duration string // Timeout duration (as string for display)
	Err      error  // Underlying error
}

func (e *TimeoutError) Error() string {
	if e.Duration != "" {
		return fmt.Sprintf("bitbrowser: timeout during %s after %s", e.Op, e.Duration)
	}
	return fmt.Sprintf("bitbrowser: timeout during %s", e.Op)
}

func (e *TimeoutError) Unwrap() error {
	return e.Err
}

func (e *TimeoutError) Is(target error) bool {
	return target == ErrTimeout
}

// RetryError represents an error after all retry attempts have been exhausted.
type RetryError struct {
	Attempts int   // Number of attempts made
	LastErr  error // The last error that occurred
}

func (e *RetryError) Error() string {
	return fmt.Sprintf("bitbrowser: retry exhausted after %d attempts: %v", e.Attempts, e.LastErr)
}

func (e *RetryError) Unwrap() error {
	return e.LastErr
}

func (e *RetryError) Is(target error) bool {
	return target == ErrRetryExhausted
}

// IsRetryable determines if an error is retryable.
// Network errors and certain HTTP status codes are considered retryable.
// API business logic errors (e.g., "profile not found") are not retryable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Network errors are retryable
	if errors.Is(err, ErrNetwork) {
		return true
	}

	// Timeout errors are retryable
	if errors.Is(err, ErrTimeout) {
		return true
	}

	// Check for specific API errors that might be retryable
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// Server errors (5xx) are retryable
		if apiErr.StatusCode >= http.StatusInternalServerError {
			return true
		}
		// Too Many Requests is retryable
		if apiErr.StatusCode == http.StatusTooManyRequests {
			return true
		}
		// Service Unavailable is retryable
		if apiErr.StatusCode == http.StatusServiceUnavailable {
			return true
		}
		// Gateway errors are retryable
		if apiErr.StatusCode == http.StatusBadGateway || apiErr.StatusCode == http.StatusGatewayTimeout {
			return true
		}
		// Other API errors (4xx, business logic) are not retryable
		return false
	}

	// Validation errors are not retryable
	if errors.Is(err, ErrValidation) {
		return false
	}

	// Retry exhausted errors are not retryable
	if errors.Is(err, ErrRetryExhausted) {
		return false
	}

	// By default, unknown errors are not retryable
	return false
}

// NewNetworkError creates a new NetworkError.
func NewNetworkError(op, url string, err error) *NetworkError {
	return &NetworkError{
		Op:  op,
		URL: url,
		Err: err,
	}
}

// NewAPIError creates a new APIError.
func NewAPIError(endpoint string, statusCode int, message string) *APIError {
	return &APIError{
		Endpoint:   endpoint,
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewTimeoutError creates a new TimeoutError.
func NewTimeoutError(op, duration string, err error) *TimeoutError {
	return &TimeoutError{
		Op:       op,
		Duration: duration,
		Err:      err,
	}
}

// NewRetryError creates a new RetryError.
func NewRetryError(attempts int, lastErr error) *RetryError {
	return &RetryError{
		Attempts: attempts,
		LastErr:  lastErr,
	}
}
