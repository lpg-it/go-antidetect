package bitbrowser

// BitBrowser API request and response structures.
// Based on BitBrowser's official API documentation.
// All endpoints use POST method with JSON body.

import "encoding/json"

// ============================================================================
// Common Response Structure
// ============================================================================

// Response is the base response structure for all BitBrowser API calls.
type Response struct {
	Success bool            `json:"success"`
	Msg     string          `json:"msg,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ============================================================================
// Fingerprint Configuration
// ============================================================================

// Fingerprint represents the browser fingerprint configuration.
type Fingerprint struct {
	// Core settings
	CoreProduct string `json:"coreProduct,omitempty"` // "chrome" | "firefox", default: chrome
	CoreVersion string `json:"coreVersion,omitempty"` // e.g., "130" for Chrome, "128" for Firefox

	// OS settings
	OSType    string `json:"ostype,omitempty"`    // "PC" | "Android" | "IOS"
	OS        string `json:"os,omitempty"`        // "Win32", "MacIntel", "Linux x86_64", "iPhone", "Linux armv81"
	OSVersion string `json:"osVersion,omitempty"` // e.g., "11,10" for Windows

	// Browser version
	Version   string `json:"version,omitempty"`   // Browser version, recommend same as CoreVersion
	UserAgent string `json:"userAgent,omitempty"` // Custom UA, leave empty for auto-generation

	// Timezone settings
	IsIpCreateTimeZone bool   `json:"isIpCreateTimeZone,omitempty"` // Generate timezone based on IP
	TimeZone           string `json:"timeZone,omitempty"`           // Manual timezone
	TimeZoneOffset     int    `json:"timeZoneOffset,omitempty"`     // Timezone offset

	// WebRTC settings: "0"=replace, "1"=allow, "2"=disable, "3"=privacy
	WebRTC string `json:"webRTC,omitempty"`

	// HTTPS settings
	IgnoreHttpsErrors bool `json:"ignoreHttpsErrors,omitempty"`

	// Geolocation: "0"=ask, "1"=allow, "2"=disable
	Position           string `json:"position,omitempty"`
	IsIpCreatePosition bool   `json:"isIpCreatePosition,omitempty"` // Generate position based on IP
	Lat                string `json:"lat,omitempty"`                // Latitude
	Lng                string `json:"lng,omitempty"`                // Longitude
	PrecisionData      string `json:"precisionData,omitempty"`      // Precision in meters

	// Language settings
	IsIpCreateLanguage        bool   `json:"isIpCreateLanguage,omitempty"`
	Languages                 string `json:"languages,omitempty"`
	IsIpCreateDisplayLanguage bool   `json:"isIpCreateDisplayLanguage,omitempty"`
	DisplayLanguages          string `json:"displayLanguages,omitempty"`

	// Window size (not fingerprint, just opening size)
	OpenWidth  int `json:"openWidth,omitempty"`  // Default 1280
	OpenHeight int `json:"openHeight,omitempty"` // Default 720

	// Resolution: "0"=follow system, "1"=custom
	ResolutionType  string `json:"resolutionType,omitempty"`
	Resolution      string `json:"resolution,omitempty"`      // e.g., "1920 x 1080"
	WindowSizeLimit bool   `json:"windowSizeLimit,omitempty"` // Limit window size to resolution

	// Display
	DevicePixelRatio float64 `json:"devicePixelRatio,omitempty"` // 1, 1.5, 2, 2.5, 3

	// Font: "0"=system default, "2"=random
	FontType string `json:"fontType,omitempty"`

	// Canvas: "0"=random, "1"=disable
	Canvas string `json:"canvas,omitempty"`

	// WebGL: "0"=random, "1"=disable
	WebGL            string `json:"webGL,omitempty"`
	WebGLMeta        string `json:"webGLMeta,omitempty"`        // "0"=custom, "1"=disable
	WebGLManufacturer string `json:"webGLManufacturer,omitempty"`
	WebGLRender      string `json:"webGLRender,omitempty"`

	// Audio: "0"=random, "1"=disable
	AudioContext string `json:"audioContext,omitempty"`

	// Media devices: "0"=random, "1"=disable
	MediaDevice string `json:"mediaDevice,omitempty"`

	// Speech voices: "0"=random, "1"=disable
	SpeechVoices string `json:"speechVoices,omitempty"`

	// Hardware
	HardwareConcurrency string `json:"hardwareConcurrency,omitempty"` // e.g., "4"
	DeviceMemory        string `json:"deviceMemory,omitempty"`        // "4" or "8", max 8

	// Do Not Track: "1"=enabled, "0"=disabled
	DoNotTrack string `json:"doNotTrack,omitempty"`

	// ClientRects noise
	ClientRectNoiseEnabled bool `json:"clientRectNoiseEnabled,omitempty"`

	// Port scan protection: "0"=enabled, "1"=disabled
	PortScanProtect string `json:"portScanProtect,omitempty"`
	PortWhiteList   string `json:"portWhiteList,omitempty"` // Comma-separated ports

	// Device info
	DeviceInfoEnabled bool   `json:"deviceInfoEnabled,omitempty"`
	ComputerName      string `json:"computerName,omitempty"`
	MacAddr           string `json:"macAddr,omitempty"`
	HostIP            string `json:"hostIP,omitempty"`

	// SSL
	DisableSslCipherSuitesFlag bool   `json:"disableSslCipherSuitesFlag,omitempty"`
	DisableSslCipherSuites     string `json:"disableSslCipherSuites,omitempty"`

	// Plugins
	EnablePlugins bool   `json:"enablePlugins,omitempty"`
	Plugins       string `json:"plugins,omitempty"`

	// Launch arguments (comma-separated, e.g., "--incognito,--no-sandbox")
	LaunchArgs string `json:"launchArgs,omitempty"`
}

// ============================================================================
// Profile Configuration (Create/Update)
// ============================================================================

// ProfileConfig represents the full configuration for creating/updating a browser profile.
type ProfileConfig struct {
	// Basic info
	ID      string `json:"id,omitempty"`      // Only for updates
	Name    string `json:"name,omitempty"`
	GroupID string `json:"groupId,omitempty"` // Group ID, defaults to "API" group
	Remark  string `json:"remark,omitempty"`

	// Platform account info
	Platform    string `json:"platform,omitempty"`    // Platform URL, e.g., "https://www.facebook.com"
	URL         string `json:"url,omitempty"`         // Additional URLs to open (comma-separated)
	UserName    string `json:"userName,omitempty"`    // Platform username for autofill
	Password    string `json:"password,omitempty"`    // Platform password for autofill
	Cookie      string `json:"cookie,omitempty"`      // JSON format cookie string
	FaSecretKey string `json:"faSecretKey,omitempty"` // 2FA secret key

	// Multi-open setting
	IsSynOpen bool `json:"isSynOpen,omitempty"` // Allow multiple opens of same profile

	// Proxy settings
	ProxyMethod   int    `json:"proxyMethod,omitempty"`   // 2=custom, 3=extract IP
	ProxyType     string `json:"proxyType,omitempty"`     // "noproxy", "http", "https", "socks5", "ssh"
	Host          string `json:"host,omitempty"`
	Port          int    `json:"port,omitempty"`
	ProxyUserName string `json:"proxyUserName,omitempty"`
	ProxyPassword string `json:"proxyPassword,omitempty"`

	// IP settings
	IpCheckService  string `json:"ipCheckService,omitempty"`  // "ip123in", "ip-api", "luminati"
	IsIpv6          bool   `json:"isIpv6,omitempty"`
	RefreshProxyUrl string `json:"refreshProxyUrl,omitempty"` // Proxy refresh URL
	EnableSocks5Udp bool   `json:"enableSocks5Udp,omitempty"` // Enable UDP for SOCKS5

	// Location for dynamic proxy
	Country  string `json:"country,omitempty"`
	Province string `json:"province,omitempty"`
	City     string `json:"city,omitempty"`

	// Dynamic IP settings
	DynamicIpUrl       string `json:"dynamicIpUrl,omitempty"`       // Extract IP URL
	DynamicIpChannel   string `json:"dynamicIpChannel,omitempty"`   // "rola", "doveip", "cloudam", "common"
	IsDynamicIpChangeIp bool   `json:"isDynamicIpChangeIp,omitempty"` // Extract new IP on each open
	DuplicateCheck     int    `json:"duplicateCheck,omitempty"`     // 1=check, 0=no check
	IsGlobalProxyInfo  bool   `json:"isGlobalProxyInfo,omitempty"`  // Use global dynamic proxy info

	// Workbench: "localserver" or "disable"
	Workbench string `json:"workbench,omitempty"`

	// Media settings
	AbortImage        bool `json:"abortImage,omitempty"`        // Block images
	AbortImageMaxSize int  `json:"abortImageMaxSize,omitempty"` // Block images larger than X KB
	AbortMedia        bool `json:"abortMedia,omitempty"`        // Block video autoplay
	MuteAudio         bool `json:"muteAudio,omitempty"`         // Mute audio

	// Network checks
	StopWhileNetError      bool `json:"stopWhileNetError,omitempty"`
	StopWhileIpChange      bool `json:"stopWhileIpChange,omitempty"`
	StopWhileCountryChange bool `json:"stopWhileCountryChange,omitempty"`

	// Sync settings
	SyncTabs          bool `json:"syncTabs,omitempty"`          // Default true
	SyncCookies       bool `json:"syncCookies,omitempty"`       // Default true
	SyncIndexedDb     bool `json:"syncIndexedDb,omitempty"`
	SyncLocalStorage  bool `json:"syncLocalStorage,omitempty"`
	SyncBookmarks     bool `json:"syncBookmarks,omitempty"`
	SyncAuthorization bool `json:"syncAuthorization,omitempty"` // Sync saved passwords
	SyncHistory       bool `json:"syncHistory,omitempty"`
	SyncExtensions    bool `json:"syncExtensions,omitempty"`

	// Credentials
	CredentialsEnableService bool `json:"credentialsEnableService,omitempty"` // Disable save password popup
	AllowedSignin            bool `json:"allowedSignin,omitempty"`            // Allow Google account login

	// Validation
	IsValidUsername bool `json:"isValidUsername,omitempty"` // Check duplicate by platform/username/password

	// Clear before launch
	ClearCacheFilesBeforeLaunch    bool `json:"clearCacheFilesBeforeLaunch,omitempty"`
	ClearCacheWithoutExtensions    bool `json:"clearCacheWithoutExtensions,omitempty"`
	ClearCookiesBeforeLaunch       bool `json:"clearCookiesBeforeLaunch,omitempty"`
	ClearHistoriesBeforeLaunch     bool `json:"clearHistoriesBeforeLaunch,omitempty"`

	// Random fingerprint on each launch
	RandomFingerprint bool `json:"randomFingerprint,omitempty"`

	// Browser settings
	DisableGpu            bool `json:"disableGpu,omitempty"`
	DisableTranslatePopup bool `json:"disableTranslatePopup,omitempty"`
	DisableNotifications  bool `json:"disableNotifications,omitempty"`
	DisableClipboard      bool `json:"disableClipboard,omitempty"`
	MemorySaver           bool `json:"memorySaver,omitempty"` // Memory saver mode

	// Fingerprint object (required, can be empty {})
	BrowserFingerPrint *Fingerprint `json:"browserFingerPrint"`
}

// ============================================================================
// Profile Partial Update
// ============================================================================

// PartialUpdateRequest represents a batch partial update request.
type PartialUpdateRequest struct {
	IDs []string `json:"ids"` // Profile IDs to update
	ProfileConfig
}

// ============================================================================
// Open Browser
// ============================================================================

// OpenOptions provides convenient options for opening a browser.
// This is the recommended way to open browsers with common settings.
type OpenOptions struct {
	// Headless runs the browser in headless mode (no GUI).
	// Note: Headless mode only supports one tab, synced URLs will be ignored.
	Headless bool

	// AllowLAN allows connections from LAN/public IP instead of only localhost.
	// This adds --remote-debugging-address=0.0.0.0 to enable remote access.
	// Useful for accessing the browser from other machines.
	AllowLAN bool

	// Incognito opens the browser in incognito/private mode.
	Incognito bool

	// IgnoreDefaultUrls ignores synced URLs and opens a blank page or workbench.
	// Recommended to set true for automation scripts.
	IgnoreDefaultUrls bool

	// StartURL specifies a URL to open when the browser starts.
	// Only works when IgnoreDefaultUrls is true.
	StartURL string

	// CustomPort specifies a fixed debugging port.
	// If 0, a random port will be assigned by BitBrowser.
	// Useful when you need a predictable port.
	CustomPort int

	// DisableGPU disables GPU hardware acceleration.
	DisableGPU bool

	// LoadExtensions specifies extension paths to load.
	// Multiple paths should be comma-separated.
	// Example: "/path/to/ext1,/path/to/ext2"
	LoadExtensions string

	// ExtraArgs allows passing additional Chrome arguments.
	// These are appended to the args array.
	ExtraArgs []string

	// WaitReady waits for the browser to be fully ready before returning.
	// If the browser is still starting, it will poll until ready.
	// Default timeout is 30 seconds.
	WaitReady bool

	// WaitTimeout specifies the maximum time in seconds to wait for browser ready.
	// Only used when WaitReady is true. Default is 30 seconds.
	WaitTimeout int

	// PollInterval specifies the interval in seconds between browser ready checks.
	// Only used when WaitReady is true. Default is 2 seconds.
	PollInterval int
}

// OpenConfig represents the raw API request for opening a browser.
// For most use cases, prefer using OpenOptions with the Open method.
type OpenConfig struct {
	ID                string   `json:"id"`                          // Profile ID (required)
	Args              []string `json:"args,omitempty"`              // Chromium launch arguments
	Queue             bool     `json:"queue,omitempty"`             // Queue mode to prevent concurrent errors
	IgnoreDefaultUrls bool     `json:"ignoreDefaultUrls,omitempty"` // Ignore synced URLs
	NewPageUrl        string   `json:"newPageUrl,omitempty"`        // URL to open (requires IgnoreDefaultUrls)
}

// OpenResult contains the browser connection information after opening.
type OpenResult struct {
	Ws          string `json:"ws"`          // WebSocket URL
	Http        string `json:"http"`        // HTTP address (host:port)
	CoreVersion string `json:"coreVersion"` // Browser kernel version
	Driver      string `json:"driver"`      // ChromeDriver path
	Seq         int    `json:"seq"`         // Window sequence number
	Name        string `json:"name"`        // Profile name
	Remark      string `json:"remark"`      // Profile remark
	GroupID     string `json:"groupId"`     // Group ID
	PID         int    `json:"pid"`         // Process ID
}

// ============================================================================
// Browser Version (CDP)
// ============================================================================

// BrowserVersion contains browser version information from CDP.
type BrowserVersion struct {
	Browser              string `json:"Browser"`
	ProtocolVersion      string `json:"Protocol-Version"`
	UserAgent            string `json:"User-Agent"`
	V8Version            string `json:"V8-Version"`
	WebKitVersion        string `json:"WebKit-Version"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// ============================================================================
// Profile Detail
// ============================================================================

// ProfileDetail contains detailed information about a browser profile.
type ProfileDetail struct {
	ID                   string       `json:"id"`
	Seq                  int          `json:"seq"`
	Name                 string       `json:"name"`
	Remark               string       `json:"remark"`
	Platform             string       `json:"platform"`
	URL                  string       `json:"url"`
	UserName             string       `json:"userName"`
	Password             string       `json:"password"`
	Cookie               string       `json:"cookie"`
	Status               int          `json:"status"`
	GroupID              string       `json:"groupId"`
	CreatedTime          string       `json:"createdTime"`
	ProxyMethod          int          `json:"proxyMethod"`
	ProxyType            string       `json:"proxyType"`
	Host                 string       `json:"host"`
	Port                 int          `json:"port"`
	ProxyUserName        string       `json:"proxyUserName"`
	ProxyPassword        string       `json:"proxyPassword"`
	LastIp               string       `json:"lastIp"`
	LastCountry          string       `json:"lastCountry"`
	BrowserFingerPrint   *Fingerprint `json:"browserFingerPrint"`
	// ... many more fields available in full response
}

// ============================================================================
// Profile List
// ============================================================================

// ListRequest represents a request to list browser profiles.
type ListRequest struct {
	Page     int    `json:"page"`               // Page number, starts from 0
	PageSize int    `json:"pageSize"`           // Max 100
	GroupID  string `json:"groupId,omitempty"`  // Filter by group
	Name     string `json:"name,omitempty"`     // Filter by name (fuzzy match)
	Remark   string `json:"remark,omitempty"`   // Filter by remark
	Seq      int    `json:"seq,omitempty"`      // Filter by exact sequence number
	MinSeq   int    `json:"minSeq,omitempty"`   // Range query min
	MaxSeq   int    `json:"maxSeq,omitempty"`   // Range query max
	Sort     string `json:"sort,omitempty"`     // "asc" or "desc"
}

// ListResult contains the paginated list response.
type ListResult struct {
	List  []ProfileDetail `json:"list"`
	Page  int             `json:"page"`
	Total int             `json:"totalNum"`
}

// ============================================================================
// Proxy Update
// ============================================================================

// ProxyUpdateRequest represents a batch proxy update request.
type ProxyUpdateRequest struct {
	IDs               []string `json:"ids"`                         // Profile IDs
	IpCheckService    string   `json:"ipCheckService,omitempty"`    // "ip123in", "ip-api", "luminati"
	ProxyMethod       int      `json:"proxyMethod"`                 // 2=custom, 3=extract IP
	ProxyType         string   `json:"proxyType"`                   // "http", "https", "socks5", "ssh", "noproxy"
	Host              string   `json:"host"`
	Port              int      `json:"port"`
	ProxyUserName     string   `json:"proxyUserName"`
	ProxyPassword     string   `json:"proxyPassword"`
	RefreshProxyUrl   string   `json:"refreshProxyUrl,omitempty"`
	DynamicIpUrl      string   `json:"dynamicIpUrl,omitempty"`
	DynamicIpChannel  string   `json:"dynamicIpChannel,omitempty"`  // "rola", "ipidea", "doveip", "cloudam", "common"
	IsDynamicIpChangeIp bool   `json:"isDynamicIpChangeIp,omitempty"`
	IsIpv6            bool     `json:"isIpv6,omitempty"`
}

// ============================================================================
// Proxy Check
// ============================================================================

// ProxyCheckRequest represents a proxy check request.
type ProxyCheckRequest struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	ProxyType      string `json:"proxyType"`      // "http", "socks5", "ssh"
	ProxyUserName  string `json:"proxyUserName"`
	ProxyPassword  string `json:"proxyPassword"`
	IpCheckService string `json:"ipCheckService"` // "ip123in", "ip-api"
	CheckExists    int    `json:"checkExists"`    // 1=check if used, 0=no check
}

// ProxyCheckResult contains proxy check results.
type ProxyCheckResult struct {
	Success bool `json:"success"`
	Data    struct {
		IP          string  `json:"ip"`
		CountryName string  `json:"countryName"`
		CountryCode string  `json:"countryCode"`
		StateProv   string  `json:"stateProv"`
		Region      string  `json:"region"`
		City        string  `json:"city"`
		Languages   string  `json:"languages"`
		TimeZone    string  `json:"timeZone"`
		Offset      string  `json:"offset"`
		Longitude   string  `json:"longitude"`
		Latitude    string  `json:"latitude"`
		Zip         string  `json:"zip"`
		Status      int     `json:"status"`
		Used        bool    `json:"used"`
		UsedTime    *string `json:"usedTime"`
	} `json:"data"`
}

// ============================================================================
// Window Arrangement
// ============================================================================

// WindowBoundsRequest represents a window arrangement request.
type WindowBoundsRequest struct {
	Type     string   `json:"type"`              // "box" or "diagonal"
	StartX   int      `json:"startX"`            // Starting X position
	StartY   int      `json:"startY"`            // Starting Y position
	Width    int      `json:"width"`             // Min 500
	Height   int      `json:"height"`            // Min 200
	Col      int      `json:"col"`               // Columns for box layout
	SpaceX   int      `json:"spaceX"`            // Horizontal spacing
	SpaceY   int      `json:"spaceY"`            // Vertical spacing
	OffsetX  int      `json:"offsetX"`           // Diagonal X offset
	OffsetY  int      `json:"offsetY"`           // Diagonal Y offset
	OrderBy  string   `json:"orderBy"`           // "asc" or "desc"
	IDs      []string `json:"ids,omitempty"`     // Profile IDs (overrides SeqList)
	SeqList  []int    `json:"seqlist,omitempty"` // Sequence numbers
	ScreenID int      `json:"screenId,omitempty"` // Display screen ID
}

// ============================================================================
// Cookie Operations
// ============================================================================

// Cookie represents a browser cookie.
type Cookie struct {
	Name      string  `json:"name"`
	Value     string  `json:"value"`
	Domain    string  `json:"domain"`
	Path      string  `json:"path,omitempty"`
	Expires   float64 `json:"expires,omitempty"`
	HttpOnly  bool    `json:"httpOnly,omitempty"`
	Secure    bool    `json:"secure,omitempty"`
	Session   bool    `json:"session,omitempty"`
	SameSite  string  `json:"sameSite,omitempty"`
	SameParty bool    `json:"sameParty,omitempty"`
}

// SetCookiesRequest represents a request to set cookies.
type SetCookiesRequest struct {
	BrowserID string   `json:"browserId"`
	Cookies   []Cookie `json:"cookies"`
}

// ClearCookiesRequest represents a request to clear cookies.
type ClearCookiesRequest struct {
	BrowserID  string `json:"browserId"`
	SaveSynced bool   `json:"saveSynced"` // Keep server-synced cookies
}

// FormatCookiesRequest represents a request to format cookies.
type FormatCookiesRequest struct {
	Cookie   any    `json:"cookie"`   // String or array
	Hostname string `json:"hostname"` // Domain for cookies without domain
}

// ============================================================================
// Display Information
// ============================================================================

// Display represents a monitor display.
type Display struct {
	ID               int    `json:"id"`
	Label            string `json:"label"`
	Bounds           Rect   `json:"bounds"`
	WorkArea         Rect   `json:"workArea"`
	ColorDepth       int    `json:"colorDepth"`
	DisplayFrequency int    `json:"displayFrequency"`
	ScaleFactor      int    `json:"scaleFactor"`
	Rotation         int    `json:"rotation"`
	Internal         bool   `json:"internal"`
}

// Rect represents a rectangle area.
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ============================================================================
// RPA Operations
// ============================================================================

// RPARequest represents an RPA task request.
type RPARequest struct {
	ID string `json:"id"` // RPA task ID
}

// ============================================================================
// Auto Paste
// ============================================================================

// AutoPasteRequest represents an auto paste request.
type AutoPasteRequest struct {
	BrowserID string `json:"browserId"`
	URL       string `json:"url"` // Must match exactly
}

// ============================================================================
// File Operations
// ============================================================================

// FileRequest represents a file read request.
type FileRequest struct {
	FilePath string `json:"filepath"` // Absolute path
}
