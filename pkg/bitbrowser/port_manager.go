package bitbrowser

import (
	"fmt"
	"math/rand/v2"
)

// PortManager handles port allocation in Managed Mode.
// It uses a stateless random selection mechanism to distribute ports evenly
// across multiple concurrent clients.
//
// The algorithm:
//  1. Generate a shuffled list of all ports in the range
//  2. Return the first port from the shuffled list
//  3. If BitBrowser reports a port conflict, the caller retries with a new port
//
// This approach is stateless and concurrency-safe, as it doesn't
// rely on memory-based bookkeeping that could become stale.
type PortManager struct {
	config *PortConfig
}

// NewPortManager creates a new PortManager with the given configuration.
//
// Returns (nil, nil) if Managed Mode is not enabled (config is nil or port range not configured).
func NewPortManager(config *PortConfig) *PortManager {
	if config == nil || !config.IsManaged() {
		return nil
	}
	return &PortManager{config: config}
}

// PickPort selects a random port from the configured range.
// It uses random shuffling to distribute load evenly and avoid
// collisions when multiple services are running concurrently.
//
// Note: This method simply returns a random port from the range.
// The actual availability is verified when BitBrowser tries to use the port.
// If there's a conflict, the caller should retry with a different port.
//
// Returns an error if the port manager is not configured.
func (pm *PortManager) PickPort() (int, error) {
	if pm == nil || pm.config == nil || !pm.config.IsManaged() {
		return 0, fmt.Errorf("port manager not configured")
	}

	size := pm.config.PortRangeSize()
	if size == 0 {
		return 0, fmt.Errorf("no ports in range [%d, %d]", pm.config.MinPort, pm.config.MaxPort)
	}

	// Pick a random port from the range
	port := pm.config.MinPort + rand.IntN(size)
	return port, nil
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
