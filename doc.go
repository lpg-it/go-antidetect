// Package antidetect provides a unified SDK for controlling antidetect browsers.
//
// This package allows developers to interact with various fingerprint browsers
// (such as BitBrowser, AdsPower, etc.) through a single, consistent API.
//
// # Supported Browsers
//
//   - BitBrowser (比特浏览器) - Fully supported
//   - AdsPower - Coming soon
//
// # Installation
//
//	go get github.com/lpg-it/go-antidetect
//
// # Quick Start
//
// Create a BitBrowser client and start controlling browsers:
//
//	package main
//
//	import (
//	    "context"
//	    antidetect "github.com/lpg-it/go-antidetect"
//	)
//
//	func main() {
//	    client := antidetect.NewBitBrowser("http://127.0.0.1:54345")
//	    ctx := context.Background()
//
//	    // Create a profile
//	    id, _ := client.CreateProfile(ctx, antidetect.ProfileConfig{
//	        Name: "my-profile",
//	    })
//
//	    // Open browser with custom port
//	    result, _ := client.OpenWithPort(ctx, id, 9222, false)
//
//	    // Use result.Ws with chromedp, playwright-go, or rod
//	    // ...
//
//	    // Close when done
//	    client.Close(ctx, id)
//	}
//
// # Features
//
// Profile Management: Create, update, delete, and list browser profiles with
// full fingerprint configuration support.
//
// Browser Control: Open and close browsers with custom debugging ports,
// headless mode, and various Chrome arguments.
//
// Proxy Management: Configure HTTP/HTTPS/SOCKS5/SSH proxies with support for
// dynamic IP extraction and proxy health checking.
//
// Cookie Management: Set, get, and clear cookies in real-time for running
// browser instances.
//
// Window Management: Arrange browser windows in grid or diagonal layouts.
//
// # Integration
//
// The SDK returns WebSocket debugging URLs that can be used with popular
// browser automation libraries:
//
//   - chromedp: chromedp.NewRemoteAllocator(ctx, result.Ws)
//   - playwright-go: pw.Chromium.ConnectOverCDP(result.Ws)
//   - rod: rod.New().ControlURL(result.Ws).MustConnect()
package antidetect
