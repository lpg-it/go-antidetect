package bitbrowser

import (
	"context"
	"log/slog"
	"time"
)

// WithLogger sets the logger for the client.
// If nil, logging is disabled.
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithRetryConfig sets the retry configuration for the client.
// If nil, no retries will be performed (MaxAttempts=1).
func WithRetryConfig(config *RetryConfig) ClientOption {
	return func(c *Client) {
		if config != nil {
			c.retryConfig = config
		}
	}
}

// WithRetry is a convenience option to enable retries with default settings.
// maxAttempts specifies the maximum number of attempts (including the initial attempt).
// For example, WithRetry(3) means 1 initial attempt + 2 retries.
func WithRetry(maxAttempts int) ClientOption {
	return func(c *Client) {
		config := DefaultRetryConfig()
		config.MaxAttempts = maxAttempts
		c.retryConfig = config
	}
}

// logRequest logs an outgoing request.
func (c *Client) logRequest(ctx context.Context, method, path string, body any) {
	if c.logger == nil {
		return
	}

	c.logger.DebugContext(ctx, "bitbrowser: sending request",
		slog.String("method", method),
		slog.String("path", path),
	)
}

// logResponse logs a response from the API.
func (c *Client) logResponse(ctx context.Context, path string, statusCode int, duration time.Duration, success bool) {
	if c.logger == nil {
		return
	}

	level := slog.LevelDebug
	if !success {
		level = slog.LevelWarn
	}

	c.logger.Log(ctx, level, "bitbrowser: received response",
		slog.String("path", path),
		slog.Int("status_code", statusCode),
		slog.Duration("duration", duration),
		slog.Bool("success", success),
	)
}

// logError logs an error.
func (c *Client) logError(ctx context.Context, path string, err error, attempt int) {
	if c.logger == nil {
		return
	}

	attrs := []any{
		slog.String("path", path),
		slog.String("error", err.Error()),
	}

	if attempt > 0 {
		attrs = append(attrs, slog.Int("attempt", attempt))
	}

	c.logger.WarnContext(ctx, "bitbrowser: request failed", attrs...)
}

// logRetry logs a retry attempt.
func (c *Client) logRetry(ctx context.Context, path string, attempt int, delay time.Duration, err error) {
	if c.logger == nil {
		return
	}

	c.logger.InfoContext(ctx, "bitbrowser: retrying request",
		slog.String("path", path),
		slog.Int("attempt", attempt),
		slog.Duration("delay", delay),
		slog.String("previous_error", err.Error()),
	)
}
