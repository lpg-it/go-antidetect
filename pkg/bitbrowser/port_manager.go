package bitbrowser

import (
	"fmt"
	"math/rand/v2"
)

// PortManager handles port allocation in Managed Mode.
// It uses a stateless random selection mechanism combined with
// the BitBrowser API to avoid port conflicts.
//
// The algorithm:
//  1. Query BitBrowser API to get currently used ports
//  2. Generate a shuffled list of all ports in the configured range
//  3. Exclude ports already used by BitBrowser
//  4. Return the first available port from the shuffled list
//
// This approach is stateless and concurrency-safe, as it queries
// the actual port usage from the API rather than relying on
// memory-based bookkeeping that could become stale.
type PortManager struct {
	config *PortConfig
	host   string // Remote host (extracted from API URL)
}

// NewPortManager creates a new PortManager with the given configuration.
// The host parameter is the BitBrowser server host.
//
// Returns nil if Managed Mode is not enabled (config is nil or port range not configured).
// Returns an error if Managed Mode is enabled but host is empty.
func NewPortManager(config *PortConfig, host string) (*PortManager, error) {
	if config == nil || !config.IsManaged() {
		return nil, nil
	}
	if host == "" {
		return nil, fmt.Errorf("bitbrowser: host is required for Managed Mode port probing")
	}
	return &PortManager{config: config, host: host}, nil
}

// PickPort selects an available port from the configured range.
// Deprecated: Use PickPortExcluding instead for better reliability.
func (pm *PortManager) PickPort() (int, error) {
	return pm.PickPortExcluding(nil)
}

// PickPortExcluding selects a random port from the configured range,
// excluding the ports in the provided set.
//
// The method:
//  1. Creates a shuffled list of all ports in the range
//  2. Filters out ports that are in the excluded set
//  3. Returns the first remaining port
//
// The excluded set should contain ports already used by BitBrowser
// (obtained via GetPorts API).
//
// Returns an error if no available port is found.
func (pm *PortManager) PickPortExcluding(excluded map[int]bool) (int, error) {
	if pm == nil || pm.config == nil || !pm.config.IsManaged() {
		return 0, fmt.Errorf("port manager not configured")
	}

	ports := pm.generateShuffledPorts()
	if len(ports) == 0 {
		return 0, fmt.Errorf("no ports in range [%d, %d]", pm.config.MinPort, pm.config.MaxPort)
	}

	// Find first port not in excluded set
	for _, port := range ports {
		if excluded != nil && excluded[port] {
			continue
		}
		return port, nil
	}

	return 0, fmt.Errorf("no available port in range [%d, %d]: all %d ports are excluded (BitBrowser is using them)",
		pm.config.MinPort, pm.config.MaxPort, len(excluded))
}

// generateShuffledPorts creates a randomly shuffled slice of all ports in the range.
func (pm *PortManager) generateShuffledPorts() []int {
	size := pm.config.PortRangeSize()
	ports := make([]int, size)

	for i := range size {
		ports[i] = pm.config.MinPort + i
	}

	// Fisher-Yates shuffle
	rand.Shuffle(len(ports), func(i, j int) {
		ports[i], ports[j] = ports[j], ports[i]
	})

	return ports
}


// IsActive returns true if the PortManager is configured and active.
func (pm *PortManager) IsActive() bool {
	return pm != nil && pm.config != nil && pm.config.IsManaged()
}

// GetConfig returns the port configuration.
func (pm *PortManager) GetConfig() *PortConfig {
	if pm == nil {
		return nil
	}
	return pm.config
}

// GetHost returns the host being probed.
func (pm *PortManager) GetHost() string {
	if pm == nil {
		return ""
	}
	return pm.host
}
