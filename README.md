# go-antidetect

[![Go Reference](https://pkg.go.dev/badge/github.com/lpg-it/go-antidetect.svg)](https://pkg.go.dev/github.com/lpg-it/go-antidetect)
[![Go Report Card](https://goreportcard.com/badge/github.com/lpg-it/go-antidetect)](https://goreportcard.com/report/github.com/lpg-it/go-antidetect)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go SDK for controlling antidetect browsers. Provides a unified interface to interact with various fingerprint browsers through a simple, consistent API.

## Supported Browsers

| Browser | Status | Version |
|---------|--------|---------|
| [BitBrowser](https://www.bitbrowser.cn/) (比特浏览器) | ✅ Fully Supported | v1.0.0 |
| [AdsPower](https://www.adspower.com/) | 🚧 Coming Soon | - |

## Installation

```bash
go get github.com/lpg-it/go-antidetect
```

## Requirements

- **Go 1.24.0 or higher** - Uses modern Go features:
  - `max()` builtin function (Go 1.21+)
  - `for range int` syntax (Go 1.22+)
  - Latest toolchain optimizations (Go 1.24+)
- BitBrowser client running with API enabled (default: `http://127.0.0.1:54345`)

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    antidetect "github.com/lpg-it/go-antidetect"
)

func main() {
    // Create client - single import, no sub-packages needed!
    client, err := antidetect.NewBitBrowser("http://127.0.0.1:54345")
    if err != nil {
        log.Fatal(err)
    }

    // Control timeout via context (recommended approach)
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    // Check connection
    if err := client.Health(ctx); err != nil {
        log.Fatal(err)
    }

    // Create a browser profile
    profileID, err := client.CreateProfile(ctx, antidetect.ProfileConfig{
        Name: "my-profile",
        BrowserFingerPrint: &antidetect.Fingerprint{
            CoreVersion: "130",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // Open browser with convenient options
    result, err := client.Open(ctx, profileID, &antidetect.OpenOptions{
        AllowLAN:          true,  // Allow LAN/remote access
        IgnoreDefaultUrls: true,  // Start with blank page
        WaitReady:         true,  // Wait for browser to be ready
        WaitTimeout:       60,    // Wait up to 60 seconds (default: 30)
        PollInterval:      2,     // Check every 2 seconds (default: 2)
        // Headless:       true,  // Optional: headless mode
        // Incognito:      true,  // Optional: incognito mode
        // CustomPort:     9222,  // Optional: fixed debug port
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("WebSocket: %s\n", result.Ws)
    fmt.Printf("HTTP: %s\n", result.Http)

    // Verify connection is valid (timeout controlled by ctx)
    if client.VerifyDebugURL(ctx, result.Http) {
        fmt.Println("Browser is accessible!")
    }

    // Use with chromedp, playwright-go, or other CDP libraries
    // chromedp.NewRemoteAllocator(ctx, result.Ws)

    // Close browser when done
    client.Close(ctx, profileID)
}
```

## Timeout Control

The SDK follows Go best practices - **no internal hardcoded timeouts**. All timeouts are controlled by the user via `context.Context`:

```go
// Control timeout via context (recommended)
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

result, err := client.Open(ctx, profileID, opts)
```

### WaitReady Options

When using `WaitReady: true`, you can configure polling behavior:

```go
result, err := client.Open(ctx, profileID, &antidetect.OpenOptions{
    WaitReady:    true,
    WaitTimeout:  60,  // Maximum wait time in seconds (default: 30)
    PollInterval: 2,   // Check interval in seconds (default: 2)
})
```

### Custom HTTP Client

For advanced scenarios, you can provide a custom HTTP client:

```go
client, err := antidetect.NewBitBrowser(apiURL, antidetect.WithHTTPClient(&http.Client{
    Transport: &http.Transport{
        MaxIdleConns:       10,
        IdleConnTimeout:    90 * time.Second,
        DisableCompression: true,
    },
}))
if err != nil {
    log.Fatal(err)
}
```

### Managed Mode (Remote/Distributed Control)

For controlling browsers remotely across multiple machines, use Managed Mode:

```go
// Multiple machines (A, C, D, E) can connect to BitBrowser on machine B
// SDK automatically allocates ports from the range and probes remote availability
client, err := antidetect.NewBitBrowser("http://192.168.1.100:54345",
    antidetect.WithPortRange(50000, 51000), // Enable Managed Mode
    antidetect.WithAPIKey("your-api-key"),
)
if err != nil {
    log.Fatal(err)
}
```

## Features

### Profile Management
- Create, update, and delete browser profiles
- Batch operations support
- Full fingerprint configuration

### Browser Control
- Open/close browsers with custom arguments
- Custom debugging port for firewall traversal
- Headless mode support
- Queue mode for concurrent operations
- Wait for browser ready with configurable polling

### Connection Verification
- `VerifyDebugURL`: Check if debug URL is accessible
- `GetBrowserVersion`: Get browser version via CDP
- `WaitForReady`: Wait until browser is fully ready

### Proxy Management
- Configure HTTP/HTTPS/SOCKS5/SSH proxies
- Dynamic IP extraction support
- Proxy health checking

### Cookie Management
- Set/get/clear cookies in real-time
- Cookie format conversion

### Window Management
- Arrange windows in grid or diagonal layout
- Flexible auto-arrangement

### And More
- RPA task control
- Cache management
- Fingerprint randomization
- Display information
- File operations

## API Reference

### Client Methods

<details>
<summary><b>Profile Management</b></summary>

| Method | Description |
|--------|-------------|
| `CreateProfile(ctx, config)` | Create a new browser profile |
| `UpdateProfile(ctx, config)` | Update an existing profile |
| `UpdateProfilePartial(ctx, req)` | Batch update specific fields |
| `GetProfileDetail(ctx, id)` | Get profile details |
| `ListProfiles(ctx, req)` | List profiles with pagination |
| `DeleteProfile(ctx, id)` | Delete a single profile |
| `DeleteProfiles(ctx, ids)` | Batch delete profiles |
| `ResetClosingState(ctx, id)` | Reset stuck closing state |

</details>

<details>
<summary><b>Browser Control</b></summary>

| Method | Description |
|--------|-------------|
| `Open(ctx, id, opts)` | Open browser with OpenOptions (recommended) |
| `OpenRaw(ctx, config)` | Open browser with raw OpenConfig |
| `Close(ctx, id)` | Close a browser |
| `CloseBySeqs(ctx, seqs)` | Close browsers by sequence numbers |
| `CloseAll(ctx)` | Close all open browsers |
| `VerifyDebugURL(ctx, url)` | Check if debug URL is accessible |
| `GetBrowserVersion(ctx, url)` | Get browser version via CDP |
| `WaitForReady(ctx, id, timeout)` | Wait for browser to be ready |

</details>

<details>
<summary><b>Process Management</b></summary>

| Method | Description |
|--------|-------------|
| `GetPIDs(ctx, ids)` | Get process IDs for profiles |
| `GetAllPIDs(ctx)` | Get all running process IDs |
| `GetAlivePIDs(ctx, ids)` | Get alive process IDs |
| `GetPorts(ctx)` | Get debugging ports |

</details>

<details>
<summary><b>Proxy Management</b></summary>

| Method | Description |
|--------|-------------|
| `UpdateProxy(ctx, req)` | Update proxy for profiles |
| `CheckProxy(ctx, req)` | Check proxy connectivity |

</details>

<details>
<summary><b>Cookie Management</b></summary>

| Method | Description |
|--------|-------------|
| `SetCookies(ctx, id, cookies)` | Set cookies for open browser |
| `GetCookies(ctx, id)` | Get real-time cookies |
| `ClearCookies(ctx, id, saveSynced)` | Clear cookies |
| `FormatCookies(ctx, cookie, hostname)` | Format cookies |

</details>

<details>
<summary><b>Window Management</b></summary>

| Method | Description |
|--------|-------------|
| `ArrangeWindows(ctx, req)` | Arrange windows by layout |
| `ArrangeWindowsFlexible(ctx, seqList)` | Auto-arrange windows |

</details>

<details>
<summary><b>Other Operations</b></summary>

| Method | Description |
|--------|-------------|
| `Health(ctx)` | Check API connection |
| `UpdateGroup(ctx, groupID, ids)` | Move profiles to group |
| `UpdateRemark(ctx, remark, ids)` | Update profile remarks |
| `ClearCache(ctx, ids)` | Clear profile cache |
| `ClearCacheExceptExtensions(ctx, ids)` | Clear cache keeping extensions |
| `RandomizeFingerprint(ctx, id)` | Randomize fingerprint |
| `GetAllDisplays(ctx)` | Get display information |
| `RunRPA(ctx, taskID)` | Run RPA task |
| `StopRPA(ctx, taskID)` | Stop RPA task |
| `AutoPaste(ctx, id, url)` | Simulate typing from clipboard |
| `ReadExcel(ctx, filepath)` | Read Excel file |
| `ReadFile(ctx, filepath)` | Read text file |

</details>

## Configuration Types

### OpenOptions (Recommended)

```go
antidetect.OpenOptions{
    Headless:          false,        // Run in headless mode
    AllowLAN:          true,         // Allow LAN/remote access
    Incognito:         false,        // Incognito mode
    IgnoreDefaultUrls: true,         // Start with blank page
    StartURL:          "",           // URL to open on start
    CustomPort:        0,            // Fixed debug port (0 = random)
    DisableGPU:        false,        // Disable GPU acceleration
    LoadExtensions:    "",           // Extension paths (comma-separated)
    ExtraArgs:         []string{},   // Additional Chrome args
    WaitReady:         true,         // Wait for browser ready
    WaitTimeout:       30,           // Seconds to wait (default: 30)
    PollInterval:      2,            // Poll interval seconds (default: 2)
}
```

### ProfileConfig

```go
antidetect.ProfileConfig{
    Name:      "profile-name",
    GroupID:   "group-id",
    Remark:    "description",

    // Proxy settings
    ProxyMethod:   antidetect.ProxyMethodCustom, // 2=custom, 3=extract
    ProxyType:     "socks5",                     // noproxy, http, https, socks5, ssh
    Host:          "127.0.0.1",
    Port:          1080,
    ProxyUserName: "user",
    ProxyPassword: "pass",

    // Browser fingerprint
    BrowserFingerPrint: &antidetect.Fingerprint{
        CoreVersion: "130",
        OSType:      "PC",
        OS:          "Win32",
        OSVersion:   "10",
    },
}
```

## Integration with CDP Libraries

### chromedp

```go
import "github.com/chromedp/chromedp"

result, _ := client.Open(ctx, profileID, &antidetect.OpenOptions{
    AllowLAN: true,
    WaitReady: true,
})

allocatorCtx, cancel := chromedp.NewRemoteAllocator(ctx, result.Ws)
defer cancel()

taskCtx, cancel := chromedp.NewContext(allocatorCtx)
defer cancel()

chromedp.Run(taskCtx,
    chromedp.Navigate("https://example.com"),
)
```

### playwright-go

```go
import "github.com/playwright-community/playwright-go"

result, _ := client.Open(ctx, profileID, &antidetect.OpenOptions{
    AllowLAN: true,
    WaitReady: true,
})

browser, _ := pw.Chromium.ConnectOverCDP(result.Ws)
page, _ := browser.NewPage()
page.Goto("https://example.com")
```

### rod

```go
import "github.com/go-rod/rod"

result, _ := client.Open(ctx, profileID, &antidetect.OpenOptions{
    AllowLAN: true,
    WaitReady: true,
})

browser := rod.New().ControlURL(result.Ws).MustConnect()
page := browser.MustPage("https://example.com")
```

## Examples

See the [example](./example) directory for complete examples.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [BitBrowser](https://www.bitbrowser.cn/) for their excellent antidetect browser and API documentation
