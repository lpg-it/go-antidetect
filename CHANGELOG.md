# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-21

### Added

- **Full BitBrowser API Support**
  - Profile Management: Create, update, delete, list profiles
  - Browser Control: Open, close, manage browser instances
  - Process Management: Get PIDs, ports, alive status
  - Proxy Management: Configure and check proxies
  - Cookie Management: Set, get, clear, format cookies
  - Window Management: Arrange windows in grid/diagonal layout
  - Cache Management: Clear cache with extension preservation option
  - RPA Control: Run and stop RPA tasks

- **OpenOptions** - Convenient way to open browsers:
  - `Headless`, `AllowLAN`, `Incognito`, `IgnoreDefaultUrls`
  - `WaitReady` with configurable `WaitTimeout` and `PollInterval`
  - `CustomPort`, `DisableGPU`, `LoadExtensions`, `ExtraArgs`

- **Connection Verification**
  - `VerifyDebugURL` - Check if debug URL is accessible
  - `GetBrowserVersion` - Get browser version via CDP
  - `WaitForReady` - Wait until browser is fully ready

- **Managed Mode** for remote/distributed browser control:
  - `WithPortRange(min, max)` - Enable SDK-managed port allocation
  - Automatic port probing on remote BitBrowser host
  - Random port selection with retry on conflict
  - Supports multiple machines accessing same BitBrowser instance

- **Error Handling**
  - Custom error types: `ValidationError`, `NetworkError`, `APIError`, `TimeoutError`
  - `IsRetryable()` method for retry decisions
  - No silent fallbacks - configuration errors fail explicitly

- **Retry Mechanism**
  - `WithRetry(maxAttempts)` - Simple retry configuration
  - `WithRetryConfig(config)` - Advanced retry with backoff and jitter
  - Configurable retry conditions via `RetryIf` function

- **Logging Support**
  - `WithLogger(logger)` - Integrate with `slog.Logger`
  - Request/response logging for debugging

- **API Authentication**
  - `WithAPIKey(key)` - Set API key for authenticated requests

- **Flexible HTTP Client**
  - `WithHTTPClient(client)` - Custom HTTP client configuration
  - No hardcoded timeouts - use `context.Context` for timeout control

- **50+ Fingerprint Options** - Full browser fingerprint configuration

- **Single Import Usage**
  ```go
  import antidetect "github.com/lpg-it/go-antidetect"

  client, err := antidetect.NewBitBrowser("http://127.0.0.1:54345")
  ```

### Notes

- Requires Go 1.24.0 or higher
- Requires BitBrowser client with API enabled
- Compatible with chromedp, playwright-go, rod, and other CDP libraries
