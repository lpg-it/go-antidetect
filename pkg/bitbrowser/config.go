package bitbrowser

// PortConfig configures the port management behavior.
//
// The SDK supports two working modes:
//
// # Managed Mode (recommended for remote/distributed control)
//
// When MinPort and MaxPort are set (non-zero), the SDK takes control of port allocation:
//   - Randomly selects ports from the specified range
//   - Forces binding to 0.0.0.0 for remote access
//   - Uses TCP probe + retry mechanism for conflict resolution
//   - Recommended range: MinPort=50000, MaxPort=51000
//
// Example:
//
//	client := bitbrowser.New(apiURL,
//	    bitbrowser.WithPortRange(50000, 51000),
//	)
//
// # Native Mode (default, for local development)
//
// When MinPort or MaxPort is 0, the SDK does not manage ports:
//   - BitBrowser assigns ports automatically (usually 127.0.0.1:random)
//   - No 0.0.0.0 binding is forced
//   - Simpler but not suitable for remote access
//
// WARNING: If you need to control browsers remotely (across machines),
// you MUST configure MinPort and MaxPort to enable Managed Mode.
// Otherwise, the returned WebSocket URL (127.0.0.1) will be unreachable.
type PortConfig struct {
	// MinPort is the minimum port number in the range (inclusive).
	// Set to 0 to disable Managed Mode.
	MinPort int

	// MaxPort is the maximum port number in the range (inclusive).
	// Set to 0 to disable Managed Mode.
	MaxPort int

	// MaxRetries is the maximum number of retry attempts when port conflicts occur.
	// Only applicable in Managed Mode. Default is 10.
	MaxRetries int
}

// DefaultPortConfig returns a PortConfig with Native Mode (no port management).
func DefaultPortConfig() *PortConfig {
	return &PortConfig{
		MinPort:    0,
		MaxPort:    0,
		MaxRetries: 10,
	}
}

// IsManaged returns true if Managed Mode is enabled (port range is configured).
func (c *PortConfig) IsManaged() bool {
	return c != nil && c.MinPort > 0 && c.MaxPort > 0 && c.MinPort <= c.MaxPort
}

// PortRangeSize returns the number of ports in the configured range.
// Returns 0 if not in Managed Mode.
func (c *PortConfig) PortRangeSize() int {
	if !c.IsManaged() {
		return 0
	}
	return c.MaxPort - c.MinPort + 1
}

// WithPortRange sets the port range for Managed Mode.
// When configured, the SDK will:
//   - Randomly select ports from the range [minPort, maxPort]
//   - Force binding to 0.0.0.0 for remote access
//   - Automatically retry on port conflicts
//
// Recommended for remote/distributed browser control:
//
//	client := bitbrowser.New(apiURL, bitbrowser.WithPortRange(50000, 51000))
//
// If minPort or maxPort is 0, Managed Mode is disabled.
func WithPortRange(minPort, maxPort int) ClientOption {
	return func(c *Client) {
		if c.portConfig == nil {
			c.portConfig = DefaultPortConfig()
		}
		c.portConfig.MinPort = minPort
		c.portConfig.MaxPort = maxPort
	}
}

// WithPortRetries sets the maximum number of retry attempts for port allocation.
// Only applicable when Managed Mode is enabled via WithPortRange.
// Default is 10 retries.
func WithPortRetries(maxRetries int) ClientOption {
	return func(c *Client) {
		if c.portConfig == nil {
			c.portConfig = DefaultPortConfig()
		}
		c.portConfig.MaxRetries = maxRetries
	}
}
