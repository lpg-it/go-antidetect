package bitbrowser

import (
	"fmt"
	"math/rand/v2"
	"net"
	"time"
)

// PortManager handles port allocation in Managed Mode.
// It uses a stateless "random probe + TCP check" mechanism
// to avoid conflicts in multi-service environments.
//
// The algorithm:
//  1. Generate a shuffled list of all ports in the range
//  2. For each port, perform a TCP probe to check availability
//  3. Return the first available port
//  4. If all ports are busy, return an error
//
// This approach is stateless and concurrency-safe, as it doesn't
// rely on memory-based bookkeeping that could become stale.
type PortManager struct {
	config *PortConfig
	host   string // Remote host to probe (extracted from API URL)
}

// NewPortManager creates a new PortManager with the given configuration.
// The host parameter is the BitBrowser server host (extracted from API URL),
// used for remote port availability probing.
//
// Returns (nil, nil) if Managed Mode is not enabled (config is nil or port range not configured).
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
// It uses random shuffling to distribute load evenly and avoid
// collisions when multiple services are running concurrently.
//
// The method:
//  1. Creates a shuffled list of all ports in the range
//  2. Probes each port with TCP to check if it's in use
//  3. Returns the first port that passes the probe
//
// Returns an error if no available port is found.
func (pm *PortManager) PickPort() (int, error) {
	if pm == nil || pm.config == nil || !pm.config.IsManaged() {
		return 0, fmt.Errorf("port manager not configured")
	}

	ports := pm.generateShuffledPorts()

	for _, port := range ports {
		if pm.isPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available port in range [%d, %d]", pm.config.MinPort, pm.config.MaxPort)
}

// generateShuffledPorts creates a randomly shuffled slice of all ports in the range.
func (pm *PortManager) generateShuffledPorts() []int {
	size := pm.config.PortRangeSize()
	ports := make([]int, size)

	for i := 0; i < size; i++ {
		ports[i] = pm.config.MinPort + i
	}

	// Fisher-Yates shuffle
	rand.Shuffle(len(ports), func(i, j int) {
		ports[i], ports[j] = ports[j], ports[i]
	})

	return ports
}

// isPortAvailable checks if a port is available by attempting a TCP connection.
// Returns true if the port is not in use (connection refused or timeout).
func (pm *PortManager) isPortAvailable(port int) bool {
	address := net.JoinHostPort(pm.host, fmt.Sprintf("%d", port))

	conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
	if err != nil {
		// Connection failed = port is available (not listening)
		return true
	}

	// Connection succeeded = port is in use
	conn.Close()
	return false
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
