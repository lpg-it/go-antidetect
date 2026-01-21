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

- Go 1.21 or higher
- BitBrowser client running with API enabled (default: `http://127.0.0.1:54345`)

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    antidetect "github.com/lpg-it/go-antidetect"
)

func main() {
    // Create client - single import, no sub-packages needed!
    client := antidetect.NewBitBrowser("http://127.0.0.1:54345")
    ctx := context.Background()

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
        // Headless:       true,  // Optional: headless mode
        // Incognito:      true,  // Optional: incognito mode
        // CustomPort:     9222,  // Optional: fixed debug port
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("WebSocket: %s\n", result.Ws)
    fmt.Printf("HTTP: %s\n", result.Http)

    // Verify connection is valid
    if client.VerifyDebugURL(ctx, result.Http) {
        fmt.Println("Browser is accessible!")
    }

    // Use with chromedp, playwright-go, or other CDP libraries
    // chromedp.NewRemoteAllocator(ctx, result.Ws)

    // Close browser when done
    client.Close(ctx, profileID)
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
| `Open(ctx, config)` | Open a browser with full config |
| `OpenWithPort(ctx, id, port, headless)` | Open with custom debug port |
| `Close(ctx, id)` | Close a browser |
| `CloseBySeqs(ctx, seqs)` | Close browsers by sequence numbers |
| `CloseAll(ctx)` | Close all open browsers |

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

### OpenConfig

```go
antidetect.OpenConfig{
    ID:    "profile-id",
    Args:  []string{"--remote-debugging-port=9222"},
    Queue: true,  // Prevent concurrent startup errors
}
```

## Integration with CDP Libraries

### chromedp

```go
import "github.com/chromedp/chromedp"

result, _ := client.OpenWithPort(ctx, profileID, 9222, false)

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

result, _ := client.OpenWithPort(ctx, profileID, 9222, false)

browser, _ := pw.Chromium.ConnectOverCDP(result.Ws)
page, _ := browser.NewPage()
page.Goto("https://example.com")
```

### rod

```go
import "github.com/go-rod/rod"

result, _ := client.OpenWithPort(ctx, profileID, 9222, false)

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
