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
