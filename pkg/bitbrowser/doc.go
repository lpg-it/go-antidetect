// Package bitbrowser provides a client for the BitBrowser API.
//
// BitBrowser (比特浏览器) is an antidetect browser that allows creating
// multiple browser profiles with unique fingerprints. This package provides
// a complete Go client for the BitBrowser local API.
//
// # Usage
//
// While you can use this package directly, it's recommended to use the main
// antidetect package for a simpler import:
//
//	import antidetect "github.com/lpg-it/go-antidetect"
//	client := antidetect.NewBitBrowser("http://127.0.0.1:54345")
//
// # Direct Usage
//
// If you prefer to use the bitbrowser package directly:
//
//	import "github.com/lpg-it/go-antidetect/pkg/bitbrowser"
//	client := bitbrowser.New("http://127.0.0.1:54345")
//
// # API Coverage
//
// This package implements all 35 BitBrowser API endpoints including:
//   - Profile management (create, update, delete, list)
//   - Browser control (open, close, process management)
//   - Proxy configuration and checking
//   - Cookie operations
//   - Window arrangement
//   - Cache management
//   - RPA task control
//   - And more
//
// # Configuration
//
// The ProfileConfig struct supports all 50+ fingerprint options available
// in the BitBrowser API, including OS type, browser version, WebRTC settings,
// canvas fingerprint, WebGL, audio context, and more.
package bitbrowser
