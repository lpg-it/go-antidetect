// Package antidetect provides a unified SDK for controlling antidetect browsers.
//
// This package allows you to interact with various fingerprint browsers through
// a single, consistent API. Currently supported browsers:
//   - BitBrowser (比特浏览器)
//
// Basic usage:
//
//	import antidetect "github.com/lpg-it/go-antidetect"
//
//	// Create a BitBrowser client
//	client := antidetect.NewBitBrowser("http://127.0.0.1:54345")
//
//	// Check connection
//	if err := client.Health(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create a profile
//	id, err := client.CreateProfile(ctx, antidetect.ProfileConfig{
//	    Name: "my-profile",
//	})
//
//	// Open the browser with options
//	result, err := client.Open(ctx, id, &antidetect.OpenOptions{
//	    AllowLAN:          true,
//	    IgnoreDefaultUrls: true,
//	})
//	fmt.Println("WebSocket:", result.Ws)
package antidetect

import (
	"github.com/lpg-it/go-antidetect/pkg/bitbrowser"
)

// ============================================================================
// Browser Type Constants
// ============================================================================

const (
	// TypeBitBrowser represents BitBrowser (比特浏览器)
	TypeBitBrowser = "bitbrowser"
	// TypeAdsPower represents AdsPower (coming soon)
	TypeAdsPower = "adspower"
)

// ============================================================================
// BitBrowser Client
// ============================================================================

// BitBrowserClient is an alias for the BitBrowser client.
type BitBrowserClient = bitbrowser.Client

// BitBrowserOption is a function that configures a BitBrowser client.
type BitBrowserOption = bitbrowser.ClientOption

// WithHTTPClient sets a custom HTTP client for the BitBrowser client.
// Use this to configure custom transport settings.
//
// Note: Timeouts should be controlled via context.Context, not HTTP client timeout.
var WithHTTPClient = bitbrowser.WithHTTPClient

// WithAPIKey sets the API key for authentication.
// The key will be sent in the "x-api-key" header with each request.
// You can find your API key in BitBrowser settings.
//
// Example:
//
//	client := antidetect.NewBitBrowser(apiURL, antidetect.WithAPIKey("56d2b7c905"))
var WithAPIKey = bitbrowser.WithAPIKey

// WithLogger sets the logger for the client.
// If nil, logging is disabled.
var WithLogger = bitbrowser.WithLogger

// WithRetry is a convenience option to enable retries with default settings.
// maxAttempts specifies the maximum number of attempts (including the initial attempt).
// For example, WithRetry(3) means 1 initial attempt + 2 retries.
var WithRetry = bitbrowser.WithRetry

// WithRetryConfig sets the retry configuration for the client.
// If nil, no retries will be performed (MaxAttempts=1).
var WithRetryConfig = bitbrowser.WithRetryConfig

// WithPortRange sets the port range for Managed Mode.
// When configured, the SDK will:
//   - Randomly select ports from the range [minPort, maxPort]
//   - Force binding to 0.0.0.0 for remote access
//   - Automatically retry on port conflicts
//
// Recommended for remote/distributed browser control:
//
//	client := antidetect.NewBitBrowser(apiURL, antidetect.WithPortRange(50000, 51000))
//
// If minPort or maxPort is 0, Managed Mode is disabled (Native Mode).
//
// WARNING: For cross-machine remote control, you MUST configure port range
// to enable Managed Mode. Otherwise, the WebSocket URL (127.0.0.1) will be
// unreachable from remote hosts. Recommended range: MinPort=50000, MaxPort=51000.
var WithPortRange = bitbrowser.WithPortRange

// WithPortRetries sets the maximum number of retry attempts for port allocation.
// Only applicable when Managed Mode is enabled via WithPortRange.
// Default is 10 retries.
var WithPortRetries = bitbrowser.WithPortRetries

// NewBitBrowser creates a new BitBrowser client.
// apiURL should be the BitBrowser API endpoint, e.g., "http://127.0.0.1:54345".
//
// By default, no timeout is set. Timeouts should be controlled via context:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	client.Open(ctx, id, opts)
//
// To customize the HTTP client:
//
//	client := antidetect.NewBitBrowser(apiURL, antidetect.WithHTTPClient(&http.Client{
//	    Transport: customTransport,
//	}))
func NewBitBrowser(apiURL string, opts ...BitBrowserOption) *BitBrowserClient {
	return bitbrowser.New(apiURL, opts...)
}

// ============================================================================
// Re-export BitBrowser Types
// ============================================================================

// ProfileConfig represents the full configuration for creating/updating a browser profile.
type ProfileConfig = bitbrowser.ProfileConfig

// Fingerprint represents the browser fingerprint configuration.
type Fingerprint = bitbrowser.Fingerprint

// OpenOptions provides convenient options for opening a browser.
// This is the recommended way to open browsers with common settings.
type OpenOptions = bitbrowser.OpenOptions

// OpenConfig represents the raw API request for opening a browser.
// For most use cases, prefer using OpenOptions with the Open method.
type OpenConfig = bitbrowser.OpenConfig

// OpenResult contains the browser connection information after opening.
type OpenResult = bitbrowser.OpenResult

// BrowserVersion contains browser version information from CDP.
type BrowserVersion = bitbrowser.BrowserVersion

// ProfileDetail contains detailed information about a browser profile.
type ProfileDetail = bitbrowser.ProfileDetail

// ListRequest represents a request to list browser profiles.
type ListRequest = bitbrowser.ListRequest

// ListResult contains the paginated list response.
type ListResult = bitbrowser.ListResult

// PartialUpdateRequest represents a batch partial update request.
type PartialUpdateRequest = bitbrowser.PartialUpdateRequest

// ProxyUpdateRequest represents a batch proxy update request.
type ProxyUpdateRequest = bitbrowser.ProxyUpdateRequest

// ProxyCheckRequest represents a proxy check request.
type ProxyCheckRequest = bitbrowser.ProxyCheckRequest

// ProxyCheckResult contains proxy check results.
type ProxyCheckResult = bitbrowser.ProxyCheckResult

// WindowBoundsRequest represents a window arrangement request.
type WindowBoundsRequest = bitbrowser.WindowBoundsRequest

// Cookie represents a browser cookie.
type Cookie = bitbrowser.Cookie

// Display represents a monitor display.
type Display = bitbrowser.Display

// Rect represents a rectangle area.
type Rect = bitbrowser.Rect

// RetryConfig configures the retry behavior.
type RetryConfig = bitbrowser.RetryConfig

// PortConfig configures the port management behavior.
// See the package documentation for detailed usage of Managed Mode vs Native Mode.
type PortConfig = bitbrowser.PortConfig

// DefaultRetryConfig returns a RetryConfig with sensible defaults.
// By default, MaxAttempts is 1 (no retries) for backward compatibility.
var DefaultRetryConfig = bitbrowser.DefaultRetryConfig

// DefaultPortConfig returns a PortConfig with Native Mode (no port management).
var DefaultPortConfig = bitbrowser.DefaultPortConfig

// ============================================================================
// Error Types
// ============================================================================

// Sentinel errors for error type checking using errors.Is().
var (
	// ErrNetwork indicates a network-level error (connection, DNS, etc.).
	ErrNetwork = bitbrowser.ErrNetwork

	// ErrAPI indicates an API-level error (non-2xx response, business logic error).
	ErrAPI = bitbrowser.ErrAPI

	// ErrValidation indicates a validation error (invalid input).
	ErrValidation = bitbrowser.ErrValidation

	// ErrTimeout indicates a timeout error.
	ErrTimeout = bitbrowser.ErrTimeout

	// ErrRetryExhausted indicates all retry attempts have been exhausted.
	ErrRetryExhausted = bitbrowser.ErrRetryExhausted
)

// NetworkError represents a network-level error.
type NetworkError = bitbrowser.NetworkError

// APIError represents an API-level error from BitBrowser.
type APIError = bitbrowser.APIError

// ValidationError represents an input validation error.
type ValidationError = bitbrowser.ValidationError

// TimeoutError represents a timeout error.
type TimeoutError = bitbrowser.TimeoutError

// RetryError represents an error after all retry attempts have been exhausted.
type RetryError = bitbrowser.RetryError

// IsRetryable determines if an error is retryable.
// Network errors and certain HTTP status codes are considered retryable.
// API business logic errors (e.g., "profile not found") are not retryable.
var IsRetryable = bitbrowser.IsRetryable

// ============================================================================
// Constants
// ============================================================================

const (
	// DefaultCoreVersion is the default Chrome kernel version.
	DefaultCoreVersion = bitbrowser.DefaultCoreVersion
	// ProxyMethodCustom indicates using a custom proxy (value: 2).
	ProxyMethodCustom = bitbrowser.ProxyMethodCustom
	// ProxyMethodExtract indicates using extracted IP (value: 3).
	ProxyMethodExtract = bitbrowser.ProxyMethodExtract
)
