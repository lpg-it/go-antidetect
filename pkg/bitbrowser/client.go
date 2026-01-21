package bitbrowser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Default values for BitBrowser configuration.
const (
	// DefaultCoreVersion is the default Chrome kernel version.
	DefaultCoreVersion = "130"
	// ProxyMethodCustom indicates using a custom proxy.
	ProxyMethodCustom = 2
	// ProxyMethodExtract indicates using extracted IP.
	ProxyMethodExtract = 3
)

// Client is the BitBrowser API client.
//
// The client supports two working modes for port management:
//
// # Managed Mode (for remote/distributed control)
//
// Configure with WithPortRange to enable SDK-managed port allocation:
//
//	client, err := bitbrowser.New(apiURL,
//	    bitbrowser.WithPortRange(50000, 51000),
//	)
//
// In this mode:
//   - SDK randomly selects ports from the range
//   - Forces binding to 0.0.0.0 for remote access
//   - Automatically retries on port conflicts
//
// # Native Mode (default, for local development)
//
// Without port range configuration, BitBrowser assigns ports automatically:
//
//	client, err := bitbrowser.New(apiURL) // Native Mode
//
// WARNING: For remote browser control across machines, you MUST use Managed Mode.
// Otherwise, the WebSocket URL (127.0.0.1) will be unreachable from remote hosts.
type Client struct {
	apiURL      string
	httpClient  *http.Client
	apiKey      string // API token for authentication (x-api-key header)
	logger      *slog.Logger
	retryConfig *RetryConfig
	portConfig  *PortConfig  // Port management configuration
	portManager *PortManager // Port manager (nil in Native Mode)
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
// Use this to configure timeouts, transport settings, etc.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithAPIKey sets the API key for authentication.
// The key will be sent in the "x-api-key" header with each request.
// You can find your API key in BitBrowser settings.
//
// Example:
//
//	client, err := bitbrowser.New(apiURL, bitbrowser.WithAPIKey("56d2b7c905"))
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// New creates a new BitBrowser client.
// apiURL should be the BitBrowser API endpoint, e.g., "http://127.0.0.1:54345".
//
// By default, no timeout is set on the HTTP client. Timeouts should be controlled
// via context.Context passed to each method call:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	client.Open(ctx, id, opts)
//
// To customize the HTTP client (e.g., for custom transport or timeouts):
//
//	client, err := bitbrowser.New(apiURL, bitbrowser.WithHTTPClient(&http.Client{
//	    Timeout: 2 * time.Minute,
//	}))
//
// For remote/distributed browser control, configure port management:
//
//	client, err := bitbrowser.New(apiURL,
//	    bitbrowser.WithPortRange(50000, 51000), // Enable Managed Mode
//	    bitbrowser.WithAPIKey("your-api-key"),
//	)
func New(apiURL string, opts ...ClientOption) (*Client, error) {
	c := &Client{
		apiURL:      strings.TrimRight(apiURL, "/"),
		httpClient:  &http.Client{}, // No timeout - controlled by context
		retryConfig: DefaultRetryConfig(),
		portConfig:  DefaultPortConfig(),
	}

	for _, opt := range opts {
		opt(c)
	}

	// Initialize port manager if Managed Mode is enabled
	if c.portConfig.IsManaged() {
		// Extract host from API URL for remote port probing
		host, err := extractHost(c.apiURL)
		if err != nil {
			return nil, fmt.Errorf("bitbrowser: invalid API URL for Managed Mode: %w", err)
		}

		pm, err := NewPortManager(c.portConfig, host)
		if err != nil {
			return nil, err
		}
		c.portManager = pm

		if c.logger != nil {
			c.logger.Info("bitbrowser: Managed Mode enabled",
				slog.Int("min_port", c.portConfig.MinPort),
				slog.Int("max_port", c.portConfig.MaxPort),
				slog.String("probe_host", host),
			)
		}
	} else {
		if c.logger != nil {
			c.logger.Debug("bitbrowser: Native Mode (no port management)")
		}
	}

	return c, nil
}

// ============================================================================
// Health Check
// ============================================================================

// Health checks if the BitBrowser local server is running.
// POST /health
func (c *Client) Health(ctx context.Context) error {
	var resp Response
	if err := c.doRequest(ctx, "/health", struct{}{}, &resp); err != nil {
		return fmt.Errorf("bitbrowser: health check failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: health check returned false")
	}
	return nil
}

// ============================================================================
// Profile Management
// ============================================================================

// CreateProfile creates a new browser profile.
// POST /browser/update
func (c *Client) CreateProfile(ctx context.Context, config ProfileConfig) (string, error) {
	// Ensure fingerprint is set (required by API)
	if config.BrowserFingerPrint == nil {
		config.BrowserFingerPrint = &Fingerprint{
			CoreVersion: DefaultCoreVersion,
		}
	}

	var resp Response
	if err := c.doRequest(ctx, "/browser/update", config, &resp); err != nil {
		return "", fmt.Errorf("bitbrowser: create profile failed: %w", err)
	}
	if !resp.Success {
		return "", fmt.Errorf("bitbrowser: create profile failed: %s", resp.Msg)
	}

	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return data.ID, nil
}

// UpdateProfile updates an existing browser profile.
// POST /browser/update
func (c *Client) UpdateProfile(ctx context.Context, config ProfileConfig) error {
	if config.ID == "" {
		return NewValidationError("id", "profile ID is required for update")
	}

	var resp Response
	if err := c.doRequest(ctx, "/browser/update", config, &resp); err != nil {
		return fmt.Errorf("bitbrowser: update profile failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: update profile failed: %s", resp.Msg)
	}
	return nil
}

// UpdateProfilePartial updates specific fields of one or more profiles.
// POST /browser/update/partial
func (c *Client) UpdateProfilePartial(ctx context.Context, req PartialUpdateRequest) error {
	var resp Response
	if err := c.doRequest(ctx, "/browser/update/partial", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: partial update failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: partial update failed: %s", resp.Msg)
	}
	return nil
}

// GetProfileDetail gets detailed information about a browser profile.
// POST /browser/detail
func (c *Client) GetProfileDetail(ctx context.Context, id string) (*ProfileDetail, error) {
	req := struct {
		ID string `json:"id"`
	}{ID: id}

	var resp Response
	if err := c.doRequest(ctx, "/browser/detail", req, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: get profile detail failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: get profile detail failed: %s", resp.Msg)
	}

	var detail ProfileDetail
	if err := json.Unmarshal(resp.Data, &detail); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return &detail, nil
}

// ListProfiles gets a paginated list of browser profiles.
// POST /browser/list
func (c *Client) ListProfiles(ctx context.Context, req ListRequest) (*ListResult, error) {
	var resp Response
	if err := c.doRequest(ctx, "/browser/list", req, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: list profiles failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: list profiles failed: %s", resp.Msg)
	}

	var result ListResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return &result, nil
}

// DeleteProfile deletes a single browser profile permanently.
// POST /browser/delete
func (c *Client) DeleteProfile(ctx context.Context, id string) error {
	req := struct {
		ID string `json:"id"`
	}{ID: id}

	var resp Response
	if err := c.doRequest(ctx, "/browser/delete", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: delete profile failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: delete profile failed: %s", resp.Msg)
	}
	return nil
}

// DeleteProfiles deletes multiple browser profiles permanently (max 100).
// POST /browser/delete/ids
func (c *Client) DeleteProfiles(ctx context.Context, ids []string) error {
	req := struct {
		IDs []string `json:"ids"`
	}{IDs: ids}

	var resp Response
	if err := c.doRequest(ctx, "/browser/delete/ids", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: batch delete failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: batch delete failed: %s", resp.Msg)
	}
	return nil
}

// ResetClosingState resets a profile's closing state when it's stuck.
// POST /browser/closing/reset
func (c *Client) ResetClosingState(ctx context.Context, id string) error {
	req := struct {
		ID string `json:"id"`
	}{ID: id}

	var resp Response
	if err := c.doRequest(ctx, "/browser/closing/reset", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: reset closing state failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: reset closing state failed: %s", resp.Msg)
	}
	return nil
}

// ============================================================================
// Browser Control
// ============================================================================

// Open opens a browser instance with the specified options.
// This is the recommended method for opening browsers with convenient options.
//
// # Port Management Modes
//
// The behavior depends on whether Managed Mode is enabled:
//
// Managed Mode (WithPortRange configured):
//   - SDK allocates a port from the configured range
//   - Automatically binds to 0.0.0.0 for remote access
//   - Retries with different ports on conflict
//   - opts.CustomPort and opts.AllowLAN are ignored
//
// Native Mode (default, no port range):
//   - BitBrowser assigns ports automatically
//   - opts.CustomPort and opts.AllowLAN are respected
//   - WARNING: May return 127.0.0.1 which is unreachable remotely
//
// Example:
//
//	result, err := client.Open(ctx, "profile-id", &bitbrowser.OpenOptions{
//	    Headless:          false,
//	    IgnoreDefaultUrls: true,
//	    WaitReady:         true,
//	})
func (c *Client) Open(ctx context.Context, id string, opts *OpenOptions) (*OpenResult, error) {
	if opts == nil {
		opts = &OpenOptions{}
	}

	// Check if Managed Mode is active
	if c.portManager != nil && c.portManager.IsActive() {
		return c.openWithManagedPort(ctx, id, opts)
	}

	// Native Mode: let BitBrowser handle port allocation
	return c.openNative(ctx, id, opts)
}

// openWithManagedPort opens a browser with SDK-managed port allocation.
// It uses the following strategy:
//  1. Get all ports currently used by BitBrowser via API
//  2. Exclude those ports from the configured range
//  3. Randomly pick a port from the remaining available ports
//  4. If another program is using the port, BitBrowser will fail and SDK retries
func (c *Client) openWithManagedPort(ctx context.Context, id string, opts *OpenOptions) (*OpenResult, error) {
	maxRetries := c.portConfig.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 10
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Get ports currently used by BitBrowser
		usedPorts, err := c.getUsedPortsSet(ctx)
		if err != nil {
			if c.logger != nil {
				c.logger.Warn("bitbrowser: failed to get used ports, proceeding with random selection",
					slog.String("error", err.Error()),
				)
			}
			// Continue with empty set - we'll rely on retry if there's a conflict
			usedPorts = make(map[int]bool)
		}

		// Pick an available port (excluding used ones)
		port, err := c.portManager.PickPortExcluding(usedPorts)
		if err != nil {
			return nil, fmt.Errorf("bitbrowser: failed to allocate port: %w", err)
		}

		if c.logger != nil {
			c.logger.Debug("bitbrowser: attempting to open browser with managed port",
				slog.Int("port", port),
				slog.Int("attempt", attempt),
				slog.Int("max_retries", maxRetries),
				slog.Int("excluded_ports", len(usedPorts)),
			)
		}

		// Build Chrome arguments with managed port
		args := c.buildManagedArgs(port, opts)

		// Build request
		// Note: In headless mode, NewPageUrl must be empty (official doc requirement)
		// "Multiple targets are not supported in headless mode"
		startURL := opts.StartURL
		if opts.Headless {
			startURL = ""
		}
		config := OpenConfig{
			ID:                id,
			Args:              args,
			Queue:             true,
			IgnoreDefaultUrls: opts.IgnoreDefaultUrls || opts.Headless,
			NewPageUrl:        startURL,
		}

		result, err := c.doOpenRequest(ctx, config)
		if err == nil {
			// CRITICAL: Verify this browser actually belongs to us
			// This prevents accidentally controlling another client's browser
			// The race condition:
			//   1. We call GetPorts() -> port 50001 is free
			//   2. Another client opens their browser on port 50001
			//   3. We request port 50001 -> BitBrowser might return their browser!
			//
			// Solution: After opening, call GetPorts() again and verify that
			// OUR profile ID is mapped to the port we requested.
			verifyPorts, verifyErr := c.GetPorts(ctx)
			if verifyErr != nil {
				if c.logger != nil {
					c.logger.Warn("bitbrowser: failed to verify port ownership",
						slog.String("error", verifyErr.Error()),
					)
				}
				// Continue anyway - verification failed but browser might be OK
			} else {
				actualPortStr := verifyPorts[id]
				var actualPort int
				if actualPortStr != "" {
					fmt.Sscanf(actualPortStr, "%d", &actualPort)
				}

				if actualPort != port {
					// DANGER: The browser we got is NOT ours!
					// Either:
					// 1. Our profile didn't actually open (actualPort == 0)
					// 2. Our profile opened on a different port (shouldn't happen)
					// 3. BitBrowser returned another browser's info
					if c.logger != nil {
						c.logger.Error("bitbrowser: CRITICAL port ownership mismatch",
							slog.String("profile_id", id),
							slog.Int("requested_port", port),
							slog.Int("actual_port", actualPort),
							slog.String("actual_port_str", actualPortStr),
							slog.Int("attempt", attempt),
						)
					}
					// Close the browser we opened to prevent orphan processes
					if closeErr := c.Close(ctx, id); closeErr != nil {
						if c.logger != nil {
							c.logger.Warn("bitbrowser: failed to close browser after port mismatch",
								slog.String("profile_id", id),
								slog.String("error", closeErr.Error()),
							)
						}
					}
					lastErr = fmt.Errorf("port ownership mismatch: profile %s should be on port %d but GetPorts shows port %d", id, port, actualPort)
					continue
				}
			}

			// Ensure HTTP endpoint has protocol prefix
			if result.Http != "" && !strings.HasPrefix(result.Http, "http://") {
				result.Http = "http://" + result.Http
			}
			return result, nil
		}

		lastErr = err

		// Check if it's a port conflict error (another program using this port)
		if c.isPortConflictError(err) {
			if c.logger != nil {
				c.logger.Warn("bitbrowser: port conflict, retrying with different port",
					slog.Int("port", port),
					slog.Int("attempt", attempt),
					slog.String("error", err.Error()),
				)
			}
			continue
		}

		// Non-retryable error
		return nil, err
	}

	return nil, fmt.Errorf("bitbrowser: failed to open browser after %d attempts: %w", maxRetries, lastErr)
}

// getUsedPortsSet returns a set of ports currently used by BitBrowser.
func (c *Client) getUsedPortsSet(ctx context.Context) (map[int]bool, error) {
	ports, err := c.GetPorts(ctx)
	if err != nil {
		return nil, err
	}

	usedPorts := make(map[int]bool)
	for _, portStr := range ports {
		if portStr == "" {
			continue
		}
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil && port > 0 {
			usedPorts[port] = true
		}
	}
	return usedPorts, nil
}


// openNative opens a browser using Native Mode (BitBrowser-managed ports).
func (c *Client) openNative(ctx context.Context, id string, opts *OpenOptions) (*OpenResult, error) {
	// Log warning if remote access might be needed
	if c.logger != nil && !opts.AllowLAN && opts.CustomPort == 0 {
		c.logger.Debug("bitbrowser: Native Mode - browser may bind to 127.0.0.1. " +
			"For remote access, use WithPortRange() to enable Managed Mode.")
	}

	// Build Chrome arguments from options
	args := c.buildNativeArgs(opts)

	// Build request
	// Note: In headless mode, NewPageUrl must be empty (official doc requirement)
	// "Multiple targets are not supported in headless mode"
	startURL := opts.StartURL
	if opts.Headless {
		startURL = ""
	}
	config := OpenConfig{
		ID:                id,
		Args:              args,
		Queue:             true,
		IgnoreDefaultUrls: opts.IgnoreDefaultUrls || opts.Headless,
		NewPageUrl:        startURL,
	}

	var resp Response
	if err := c.doRequest(ctx, "/browser/open", config, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: open browser failed: %w", err)
	}
	if !resp.Success {
		// Check if browser is still starting
		if opts.WaitReady && strings.Contains(resp.Msg, "正在打开") {
			return c.waitForBrowserReady(ctx, id, opts)
		}
		return nil, fmt.Errorf("bitbrowser: open browser failed: %s", resp.Msg)
	}

	var result OpenResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}

	// Ensure HTTP endpoint has protocol prefix
	if result.Http != "" && !strings.HasPrefix(result.Http, "http://") {
		result.Http = "http://" + result.Http
	}

	// If result is empty but WaitReady is enabled, wait for browser
	if result.Http == "" && opts.WaitReady {
		return c.waitForBrowserReady(ctx, id, opts)
	}

	return &result, nil
}

// buildManagedArgs builds Chrome arguments for Managed Mode.
// It always includes port binding to 0.0.0.0 for remote access.
func (c *Client) buildManagedArgs(port int, opts *OpenOptions) []string {
	var args []string

	// Managed port and address (always 0.0.0.0 for remote access)
	args = append(args, fmt.Sprintf("--remote-debugging-port=%d", port))
	args = append(args, "--remote-debugging-address=0.0.0.0")

	// Headless mode
	if opts.Headless {
		args = append(args, "--headless")
	}

	// Incognito mode
	if opts.Incognito {
		args = append(args, "--incognito")
	}

	// Disable GPU
	if opts.DisableGPU {
		args = append(args, "--disable-gpu")
	}

	// Load extensions
	if opts.LoadExtensions != "" {
		args = append(args, fmt.Sprintf("--load-extension=%s", opts.LoadExtensions))
	}

	// Extra args
	args = append(args, opts.ExtraArgs...)

	return args
}

// buildNativeArgs builds Chrome arguments for Native Mode.
// It respects user-specified CustomPort and AllowLAN options.
func (c *Client) buildNativeArgs(opts *OpenOptions) []string {
	var args []string

	// Custom port (user-specified)
	if opts.CustomPort > 0 {
		args = append(args, fmt.Sprintf("--remote-debugging-port=%d", opts.CustomPort))
	}

	// Allow LAN access (user-specified)
	if opts.AllowLAN {
		args = append(args, "--remote-debugging-address=0.0.0.0")
	}

	// Headless mode
	if opts.Headless {
		args = append(args, "--headless")
	}

	// Incognito mode
	if opts.Incognito {
		args = append(args, "--incognito")
	}

	// Disable GPU
	if opts.DisableGPU {
		args = append(args, "--disable-gpu")
	}

	// Load extensions
	if opts.LoadExtensions != "" {
		args = append(args, fmt.Sprintf("--load-extension=%s", opts.LoadExtensions))
	}

	// Extra args
	args = append(args, opts.ExtraArgs...)

	return args
}

// doOpenRequest performs the /browser/open API call and parses the response.
func (c *Client) doOpenRequest(ctx context.Context, config OpenConfig) (*OpenResult, error) {
	var resp Response
	if err := c.doRequest(ctx, "/browser/open", config, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Msg)
	}

	var result OpenResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// isPortConflictError checks if an error indicates a port conflict.
func (c *Client) isPortConflictError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	// Common port conflict indicators - be specific to avoid false positives
	// (e.g., "profile already opened" is NOT a port conflict)
	return strings.Contains(errMsg, "address already in use") ||
		strings.Contains(errMsg, "port is already in use") ||
		strings.Contains(errMsg, "端口被占用") ||
		strings.Contains(errMsg, "端口已被") ||
		strings.Contains(errMsg, "port occupied") ||
		strings.Contains(errMsg, "bind: address already in use")
}

// OpenRaw opens a browser using the raw API configuration.
// Use this when you need full control over the request parameters.
// For most cases, prefer using Open with OpenOptions instead.
func (c *Client) OpenRaw(ctx context.Context, config OpenConfig) (*OpenResult, error) {
	var resp Response
	if err := c.doRequest(ctx, "/browser/open", config, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: open browser failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: open browser failed: %s", resp.Msg)
	}

	var result OpenResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}

	// Ensure HTTP endpoint has protocol prefix
	if result.Http != "" && !strings.HasPrefix(result.Http, "http://") {
		result.Http = "http://" + result.Http
	}

	return &result, nil
}

// waitForBrowserReady polls until the browser is ready.
func (c *Client) waitForBrowserReady(ctx context.Context, id string, opts *OpenOptions) (*OpenResult, error) {
	timeout := opts.WaitTimeout
	if timeout <= 0 {
		timeout = 30 // Default 30 seconds
	}

	pollIntervalSec := opts.PollInterval
	if pollIntervalSec <= 0 {
		pollIntervalSec = 2 // Default 2 seconds
	}
	pollInterval := time.Duration(pollIntervalSec) * time.Second

	maxAttempts := max(timeout/pollIntervalSec, 1)

	for range maxAttempts {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}

		// Try to get browser ports to check if it's ready
		ports, err := c.GetPorts(ctx)
		if err == nil {
			if port, ok := ports[id]; ok && port != "" {
				// Browser is ready, construct result
				httpEndpoint := "http://127.0.0.1:" + port
				if opts.AllowLAN {
					httpEndpoint = "http://0.0.0.0:" + port
				}

				// Get WebSocket URL from browser
				version, verr := c.GetBrowserVersion(ctx, httpEndpoint)
				if verr == nil && version.WebSocketDebuggerURL != "" {
					return &OpenResult{
						Http: httpEndpoint,
						Ws:   version.WebSocketDebuggerURL,
					}, nil
				}

				return &OpenResult{
					Http: httpEndpoint,
				}, nil
			}
		}
	}

	return nil, NewTimeoutError("wait_for_browser_ready", (time.Duration(timeout) * time.Second).String(), nil)
}

// WaitForReady waits until the browser is fully ready and returns connection info.
// This is useful when you need to ensure the browser is ready before connecting.
func (c *Client) WaitForReady(ctx context.Context, id string, timeoutSeconds int) (*OpenResult, error) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	opts := &OpenOptions{
		WaitReady:   true,
		WaitTimeout: timeoutSeconds,
	}

	return c.waitForBrowserReady(ctx, id, opts)
}

// ============================================================================
// Connection Verification
// ============================================================================

// VerifyDebugURL checks if a browser debug URL is still valid and accessible.
// This is useful for verifying cached debug URLs before attempting to connect.
// The caller should set an appropriate timeout on the context.
func (c *Client) VerifyDebugURL(ctx context.Context, httpEndpoint string) bool {
	if httpEndpoint == "" {
		return false
	}

	// Try to access /json/version endpoint
	versionURL := strings.TrimSuffix(httpEndpoint, "/") + "/json/version"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, versionURL, nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetBrowserVersion gets the browser version information via CDP.
// The httpEndpoint should be the debug HTTP address (e.g., "http://127.0.0.1:9222").
func (c *Client) GetBrowserVersion(ctx context.Context, httpEndpoint string) (*BrowserVersion, error) {
	if httpEndpoint == "" {
		return nil, NewValidationError("httpEndpoint", "httpEndpoint is required")
	}

	versionURL := strings.TrimSuffix(httpEndpoint, "/") + "/json/version"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, versionURL, nil)
	if err != nil {
		return nil, NewNetworkError("create_request", versionURL, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewTimeoutError("get_browser_version", "", err)
		}
		return nil, NewNetworkError("http_request", versionURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewAPIError("/json/version", resp.StatusCode, "unexpected status code")
	}

	var version BrowserVersion
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return nil, NewAPIError("/json/version", resp.StatusCode, "failed to parse version: "+err.Error())
	}

	return &version, nil
}

// ============================================================================
// Close Browser
// ============================================================================

// Close closes a running browser instance.
// POST /browser/close
// Note: Wait at least 5 seconds before reopening or deleting the profile.
func (c *Client) Close(ctx context.Context, id string) error {
	req := struct {
		ID string `json:"id"`
	}{ID: id}

	var resp Response
	if err := c.doRequest(ctx, "/browser/close", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: close browser failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: close browser failed: %s", resp.Msg)
	}
	return nil
}

// CloseBySeqs closes browsers by their sequence numbers.
// POST /browser/close/byseqs
func (c *Client) CloseBySeqs(ctx context.Context, seqs []int) error {
	req := struct {
		Seqs []int `json:"seqs"`
	}{Seqs: seqs}

	var resp Response
	if err := c.doRequest(ctx, "/browser/close/byseqs", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: close by seqs failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: close by seqs failed: %s", resp.Msg)
	}
	return nil
}

// CloseAll closes all open browser windows.
// POST /browser/close/all
func (c *Client) CloseAll(ctx context.Context) error {
	var resp Response
	if err := c.doRequest(ctx, "/browser/close/all", struct{}{}, &resp); err != nil {
		return fmt.Errorf("bitbrowser: close all failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: close all failed: %s", resp.Msg)
	}
	return nil
}

// ============================================================================
// Process Management
// ============================================================================

// GetPIDs gets the process IDs for the specified browser profiles.
// POST /browser/pids
func (c *Client) GetPIDs(ctx context.Context, ids []string) (map[string]int, error) {
	req := struct {
		IDs []string `json:"ids"`
	}{IDs: ids}

	var resp Response
	if err := c.doRequest(ctx, "/browser/pids", req, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: get pids failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: get pids failed: %s", resp.Msg)
	}

	var result map[string]int
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return result, nil
}

// GetAllPIDs gets all running browser process IDs.
// POST /browser/pids/all
func (c *Client) GetAllPIDs(ctx context.Context) (map[string]int, error) {
	var resp Response
	if err := c.doRequest(ctx, "/browser/pids/all", struct{}{}, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: get all pids failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: get all pids failed: %s", resp.Msg)
	}

	var result map[string]int
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return result, nil
}

// GetAlivePIDs gets alive process IDs for the specified profiles.
// POST /browser/pids/alive
func (c *Client) GetAlivePIDs(ctx context.Context, ids []string) (map[string]int, error) {
	req := struct {
		IDs []string `json:"ids"`
	}{IDs: ids}

	var resp Response
	if err := c.doRequest(ctx, "/browser/pids/alive", req, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: get alive pids failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: get alive pids failed: %s", resp.Msg)
	}

	var result map[string]int
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return result, nil
}

// GetPorts gets the debugging ports for all open browsers.
// POST /browser/ports
func (c *Client) GetPorts(ctx context.Context) (map[string]string, error) {
	var resp Response
	if err := c.doRequest(ctx, "/browser/ports", struct{}{}, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: get ports failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: get ports failed: %s", resp.Msg)
	}

	var result map[string]string
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return result, nil
}

// ============================================================================
// Proxy Management
// ============================================================================

// UpdateProxy updates proxy settings for multiple profiles.
// POST /browser/proxy/update
func (c *Client) UpdateProxy(ctx context.Context, req ProxyUpdateRequest) error {
	var resp Response
	if err := c.doRequest(ctx, "/browser/proxy/update", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: update proxy failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: update proxy failed: %s", resp.Msg)
	}
	return nil
}

// CheckProxy checks if a proxy is working and gets its information.
// POST /checkagent
func (c *Client) CheckProxy(ctx context.Context, req ProxyCheckRequest) (*ProxyCheckResult, error) {
	var resp Response
	if err := c.doRequest(ctx, "/checkagent", req, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: check proxy failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: check proxy failed: %s", resp.Msg)
	}

	var result ProxyCheckResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return &result, nil
}

// ============================================================================
// Group Management
// ============================================================================

// UpdateGroup moves profiles to a specified group.
// POST /browser/group/update
func (c *Client) UpdateGroup(ctx context.Context, groupID string, browserIDs []string) error {
	req := struct {
		GroupID    string   `json:"groupId"`
		BrowserIDs []string `json:"browserIds"`
	}{
		GroupID:    groupID,
		BrowserIDs: browserIDs,
	}

	var resp Response
	if err := c.doRequest(ctx, "/browser/group/update", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: update group failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: update group failed: %s", resp.Msg)
	}
	return nil
}

// UpdateRemark updates the remark for multiple profiles.
// POST /browser/remark/update
func (c *Client) UpdateRemark(ctx context.Context, remark string, browserIDs []string) error {
	req := struct {
		Remark     string   `json:"remark"`
		BrowserIDs []string `json:"browserIds"`
	}{
		Remark:     remark,
		BrowserIDs: browserIDs,
	}

	var resp Response
	if err := c.doRequest(ctx, "/browser/remark/update", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: update remark failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: update remark failed: %s", resp.Msg)
	}
	return nil
}

// ============================================================================
// Window Arrangement
// ============================================================================

// ArrangeWindows arranges browser windows according to the specified layout.
// POST /windowbounds
func (c *Client) ArrangeWindows(ctx context.Context, req WindowBoundsRequest) error {
	var resp Response
	if err := c.doRequest(ctx, "/windowbounds", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: arrange windows failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: arrange windows failed: %s", resp.Msg)
	}
	return nil
}

// ArrangeWindowsFlexible auto-arranges windows flexibly.
// POST /windowbounds/flexable
func (c *Client) ArrangeWindowsFlexible(ctx context.Context, seqList []int) error {
	req := struct {
		SeqList []int `json:"seqlist"`
	}{SeqList: seqList}

	var resp Response
	if err := c.doRequest(ctx, "/windowbounds/flexable", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: flexible arrange failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: flexible arrange failed: %s", resp.Msg)
	}
	return nil
}

// ============================================================================
// Cache Management
// ============================================================================

// ClearCache clears all cache for the specified profiles.
// POST /cache/clear
func (c *Client) ClearCache(ctx context.Context, ids []string) error {
	req := struct {
		IDs []string `json:"ids"`
	}{IDs: ids}

	var resp Response
	if err := c.doRequest(ctx, "/cache/clear", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: clear cache failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: clear cache failed: %s", resp.Msg)
	}
	return nil
}

// ClearCacheExceptExtensions clears cache but keeps extension data.
// POST /cache/clear/exceptExtensions
func (c *Client) ClearCacheExceptExtensions(ctx context.Context, ids []string) error {
	req := struct {
		IDs []string `json:"ids"`
	}{IDs: ids}

	var resp Response
	if err := c.doRequest(ctx, "/cache/clear/exceptExtensions", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: clear cache except extensions failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: clear cache except extensions failed: %s", resp.Msg)
	}
	return nil
}

// ============================================================================
// Fingerprint Management
// ============================================================================

// RandomizeFingerprint randomizes the fingerprint for a profile.
// POST /browser/fingerprint/random
func (c *Client) RandomizeFingerprint(ctx context.Context, browserID string) (*Fingerprint, error) {
	req := struct {
		BrowserID string `json:"browserId"`
	}{BrowserID: browserID}

	var resp Response
	if err := c.doRequest(ctx, "/browser/fingerprint/random", req, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: randomize fingerprint failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: randomize fingerprint failed: %s", resp.Msg)
	}

	var result Fingerprint
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return &result, nil
}

// ============================================================================
// Cookie Management
// ============================================================================

// SetCookies sets cookies for an open browser.
// POST /browser/cookies/set
func (c *Client) SetCookies(ctx context.Context, browserID string, cookies []Cookie) error {
	req := SetCookiesRequest{
		BrowserID: browserID,
		Cookies:   cookies,
	}

	var resp Response
	if err := c.doRequest(ctx, "/browser/cookies/set", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: set cookies failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: set cookies failed: %s", resp.Msg)
	}
	return nil
}

// GetCookies gets real-time cookies from an open browser.
// POST /browser/cookies/get
func (c *Client) GetCookies(ctx context.Context, browserID string) ([]Cookie, error) {
	req := struct {
		BrowserID string `json:"browserId"`
	}{BrowserID: browserID}

	var resp Response
	if err := c.doRequest(ctx, "/browser/cookies/get", req, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: get cookies failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: get cookies failed: %s", resp.Msg)
	}

	var result []Cookie
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return result, nil
}

// ClearCookies clears cookies for a profile.
// POST /browser/cookies/clear
func (c *Client) ClearCookies(ctx context.Context, browserID string, saveSynced bool) error {
	req := ClearCookiesRequest{
		BrowserID:  browserID,
		SaveSynced: saveSynced,
	}

	var resp Response
	if err := c.doRequest(ctx, "/browser/cookies/clear", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: clear cookies failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: clear cookies failed: %s", resp.Msg)
	}
	return nil
}

// FormatCookies formats cookies to standard format.
// POST /browser/cookies/format
func (c *Client) FormatCookies(ctx context.Context, cookie any, hostname string) ([]Cookie, error) {
	req := FormatCookiesRequest{
		Cookie:   cookie,
		Hostname: hostname,
	}

	var resp Response
	if err := c.doRequest(ctx, "/browser/cookies/format", req, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: format cookies failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: format cookies failed: %s", resp.Msg)
	}

	var result []Cookie
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return result, nil
}

// ============================================================================
// Display Management
// ============================================================================

// GetAllDisplays gets information about all connected displays.
// POST /alldisplays
func (c *Client) GetAllDisplays(ctx context.Context) ([]Display, error) {
	var resp Response
	if err := c.doRequest(ctx, "/alldisplays", struct{}{}, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: get displays failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: get displays failed: %s", resp.Msg)
	}

	var result []Display
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return result, nil
}

// ============================================================================
// RPA Management
// ============================================================================

// RunRPA starts an RPA task.
// POST /rpa/run
func (c *Client) RunRPA(ctx context.Context, taskID string) error {
	req := RPARequest{ID: taskID}

	var resp Response
	if err := c.doRequest(ctx, "/rpa/run", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: run RPA failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: run RPA failed: %s", resp.Msg)
	}
	return nil
}

// StopRPA stops a running RPA task.
// POST /rpa/stop
func (c *Client) StopRPA(ctx context.Context, taskID string) error {
	req := RPARequest{ID: taskID}

	var resp Response
	if err := c.doRequest(ctx, "/rpa/stop", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: stop RPA failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: stop RPA failed: %s", resp.Msg)
	}
	return nil
}

// ============================================================================
// Utility Functions
// ============================================================================

// AutoPaste simulates typing from clipboard into the focused input field.
// POST /autopaste
func (c *Client) AutoPaste(ctx context.Context, browserID, url string) error {
	req := AutoPasteRequest{
		BrowserID: browserID,
		URL:       url,
	}

	var resp Response
	if err := c.doRequest(ctx, "/autopaste", req, &resp); err != nil {
		return fmt.Errorf("bitbrowser: auto paste failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("bitbrowser: auto paste failed: %s", resp.Msg)
	}
	return nil
}

// ReadExcel reads an Excel file from the local filesystem.
// POST /utils/readexcel
func (c *Client) ReadExcel(ctx context.Context, filepath string) (any, error) {
	req := FileRequest{FilePath: filepath}

	var resp Response
	if err := c.doRequest(ctx, "/utils/readexcel", req, &resp); err != nil {
		return nil, fmt.Errorf("bitbrowser: read excel failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("bitbrowser: read excel failed: %s", resp.Msg)
	}

	var result any
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("bitbrowser: failed to parse response: %w", err)
	}
	return result, nil
}

// ReadFile reads a text file from the local filesystem.
// POST /utils/readfile
func (c *Client) ReadFile(ctx context.Context, filepath string) (string, error) {
	req := FileRequest{FilePath: filepath}

	var resp Response
	if err := c.doRequest(ctx, "/utils/readfile", req, &resp); err != nil {
		return "", fmt.Errorf("bitbrowser: read file failed: %w", err)
	}
	if !resp.Success {
		return "", fmt.Errorf("bitbrowser: read file failed: %s", resp.Msg)
	}

	var result string
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		// If it's not a string, return the raw JSON
		return string(resp.Data), nil
	}
	return result, nil
}

// ============================================================================
// Internal HTTP Helper
// ============================================================================

// doRequest performs an HTTP POST request to the BitBrowser API with retry logic.
func (c *Client) doRequest(ctx context.Context, path string, reqBody any, respBody any) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return &ValidationError{
			Field:   "request_body",
			Message: "failed to marshal request: " + err.Error(),
		}
	}

	c.logRequest(ctx, http.MethodPost, path, reqBody)
	start := time.Now()

	r := newRetryer(c.retryConfig)
	attempt := 0

	err = r.do(ctx, func() error {
		attempt++
		execErr := c.executeRequest(ctx, path, jsonData, respBody)
		if execErr != nil {
			c.logError(ctx, path, execErr, attempt)
		}
		return execErr
	})

	duration := time.Since(start)
	success := err == nil

	// Log the final response
	c.logResponse(ctx, path, 0, duration, success)

	return err
}

// executeRequest performs a single HTTP POST request without retry.
func (c *Client) executeRequest(ctx context.Context, path string, jsonData []byte, respBody any) error {
	url := c.apiURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return NewNetworkError("create_request", url, err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add API key authentication header if configured
	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if it's a context error
		if errors.Is(err, context.DeadlineExceeded) {
			return NewTimeoutError("http_request", "", err)
		}
		if errors.Is(err, context.Canceled) {
			return err
		}
		return NewNetworkError("http_request", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewNetworkError("read_response", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return NewAPIError(path, resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, respBody); err != nil {
		return NewAPIError(path, resp.StatusCode, "failed to unmarshal response: "+err.Error())
	}

	return nil
}

// extractHost extracts the hostname from a URL string.
// Returns an error if the URL is invalid or has no host.
func extractHost(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL %q: %w", rawURL, err)
	}
	host := u.Hostname()
	if host == "" {
		return "", fmt.Errorf("URL %q has no host", rawURL)
	}
	return host, nil
}

