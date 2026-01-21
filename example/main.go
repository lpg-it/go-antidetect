package main

import (
	"context"
	"fmt"
	"log"
	"time"

	antidetect "github.com/lpg-it/go-antidetect"
)

func main() {
	// Create a BitBrowser client - single import, no sub-packages needed!
	client := antidetect.NewBitBrowser("http://127.0.0.1:54345")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// ========================================================================
	// Health Check
	// ========================================================================
	fmt.Println("Checking BitBrowser connection...")
	if err := client.Health(ctx); err != nil {
		log.Fatalf("BitBrowser is not running: %v", err)
	}
	fmt.Println("BitBrowser is connected!")

	// ========================================================================
	// Create a Profile
	// ========================================================================
	fmt.Println("\nCreating a new browser profile...")
	profileID, err := client.CreateProfile(ctx, antidetect.ProfileConfig{
		Name:   "test-profile-sdk",
		Remark: "Created by go-antidetect SDK",
		// Configure proxy (optional)
		ProxyMethod: antidetect.ProxyMethodCustom,
		ProxyType:   "socks5",
		Host:        "127.0.0.1",
		Port:        1080,
		// Fingerprint configuration (optional, will use defaults if not set)
		BrowserFingerPrint: &antidetect.Fingerprint{
			CoreVersion: "130",
			OSType:      "PC",
			OS:          "Win32",
			OSVersion:   "10",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create profile: %v", err)
	}
	fmt.Printf("Profile created with ID: %s\n", profileID)

	// ========================================================================
	// Open Browser with Custom Port (for firewall traversal)
	// ========================================================================
	fmt.Println("\nOpening browser with custom debugging port...")
	result, err := client.OpenWithPort(ctx, profileID, 9222, false)
	if err != nil {
		log.Fatalf("Failed to open browser: %v", err)
	}

	fmt.Println("Browser opened successfully!")
	fmt.Printf("  WebSocket: %s\n", result.Ws)
	fmt.Printf("  HTTP:      %s\n", result.Http)
	fmt.Printf("  Driver:    %s\n", result.Driver)
	fmt.Printf("  PID:       %d\n", result.PID)

	// ========================================================================
	// Use with chromedp or other CDP libraries
	// ========================================================================
	// Example with chromedp:
	//
	// import "github.com/chromedp/chromedp"
	//
	// allocatorCtx, cancel := chromedp.NewRemoteAllocator(ctx, result.Ws)
	// defer cancel()
	//
	// taskCtx, cancel := chromedp.NewContext(allocatorCtx)
	// defer cancel()
	//
	// chromedp.Run(taskCtx,
	//     chromedp.Navigate("https://www.google.com"),
	// )

	// Simulate some work
	fmt.Println("\nBrowser is running for 5 seconds...")
	time.Sleep(5 * time.Second)

	// ========================================================================
	// Get Cookies (while browser is open)
	// ========================================================================
	fmt.Println("\nGetting cookies...")
	cookies, err := client.GetCookies(ctx, profileID)
	if err != nil {
		fmt.Printf("Failed to get cookies: %v\n", err)
	} else {
		fmt.Printf("Found %d cookies\n", len(cookies))
	}

	// ========================================================================
	// Close Browser
	// ========================================================================
	fmt.Println("\nClosing browser...")
	if err := client.Close(ctx, profileID); err != nil {
		log.Fatalf("Failed to close browser: %v", err)
	}
	fmt.Println("Browser closed!")

	// Wait for process to fully exit
	time.Sleep(5 * time.Second)

	// ========================================================================
	// Delete Profile (optional)
	// ========================================================================
	fmt.Println("\nDeleting profile...")
	if err := client.DeleteProfile(ctx, profileID); err != nil {
		log.Fatalf("Failed to delete profile: %v", err)
	}
	fmt.Println("Profile deleted!")

	fmt.Println("\n=== Demo Complete ===")
}
