# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-21

### Added

- Initial release with full BitBrowser API support
- **Profile Management**
  - `CreateProfile` - Create new browser profiles
  - `UpdateProfile` - Update existing profiles
  - `UpdateProfilePartial` - Batch partial updates
  - `GetProfileDetail` - Get profile details
  - `ListProfiles` - Paginated profile listing
  - `DeleteProfile` - Delete single profile
  - `DeleteProfiles` - Batch delete profiles
  - `ResetClosingState` - Reset stuck browser state

- **Browser Control**
  - `Open` - Open browser with full configuration
  - `OpenWithPort` - Convenience method for custom debug port
  - `Close` - Close browser instance
  - `CloseBySeqs` - Close by sequence numbers
  - `CloseAll` - Close all browsers

- **Process Management**
  - `GetPIDs` - Get process IDs
  - `GetAllPIDs` - Get all running PIDs
  - `GetAlivePIDs` - Get alive process IDs
  - `GetPorts` - Get debugging ports

- **Proxy Management**
  - `UpdateProxy` - Batch update proxy settings
  - `CheckProxy` - Check proxy connectivity

- **Cookie Management**
  - `SetCookies` - Set cookies in real-time
  - `GetCookies` - Get current cookies
  - `ClearCookies` - Clear cookies
  - `FormatCookies` - Convert cookie formats

- **Window Management**
  - `ArrangeWindows` - Arrange by grid/diagonal layout
  - `ArrangeWindowsFlexible` - Auto-arrange windows

- **Cache Management**
  - `ClearCache` - Clear profile cache
  - `ClearCacheExceptExtensions` - Clear cache preserving extensions

- **Other Features**
  - `Health` - API health check
  - `UpdateGroup` - Move profiles between groups
  - `UpdateRemark` - Update profile remarks
  - `RandomizeFingerprint` - Randomize browser fingerprint
  - `GetAllDisplays` - Get monitor information
  - `RunRPA` / `StopRPA` - RPA task control
  - `AutoPaste` - Simulated clipboard typing
  - `ReadExcel` / `ReadFile` - File operations

- Full support for 50+ fingerprint configuration options
- Comprehensive proxy configuration including dynamic IP extraction
- Single import usage: `import antidetect "github.com/lpg-it/go-antidetect"`

### Notes

- Requires BitBrowser client with API enabled
- Compatible with chromedp, playwright-go, rod, and other CDP libraries
