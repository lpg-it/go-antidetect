package bitbrowser

import (
	"testing"
)

func TestPortConfig(t *testing.T) {
	t.Run("DefaultPortConfig returns Native Mode", func(t *testing.T) {
		config := DefaultPortConfig()

		if config.MinPort != 0 {
			t.Errorf("MinPort = %d, want 0", config.MinPort)
		}
		if config.MaxPort != 0 {
			t.Errorf("MaxPort = %d, want 0", config.MaxPort)
		}
		if config.MaxRetries != 10 {
			t.Errorf("MaxRetries = %d, want 10", config.MaxRetries)
		}
		if config.IsManaged() {
			t.Error("IsManaged() should return false for default config")
		}
	})

	t.Run("IsManaged returns true for valid range", func(t *testing.T) {
		config := &PortConfig{
			MinPort: 50000,
			MaxPort: 51000,
		}

		if !config.IsManaged() {
			t.Error("IsManaged() should return true")
		}
	})

	t.Run("IsManaged returns false for zero values", func(t *testing.T) {
		tests := []struct {
			name    string
			config  *PortConfig
			managed bool
		}{
			{"nil config", nil, false},
			{"zero MinPort", &PortConfig{MinPort: 0, MaxPort: 51000}, false},
			{"zero MaxPort", &PortConfig{MinPort: 50000, MaxPort: 0}, false},
			{"both zero", &PortConfig{MinPort: 0, MaxPort: 0}, false},
			{"inverted range", &PortConfig{MinPort: 51000, MaxPort: 50000}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.config.IsManaged() != tt.managed {
					t.Errorf("IsManaged() = %v, want %v", tt.config.IsManaged(), tt.managed)
				}
			})
		}
	})

	t.Run("PortRangeSize calculates correctly", func(t *testing.T) {
		tests := []struct {
			name     string
			config   *PortConfig
			expected int
		}{
			{"nil config", nil, 0},
			{"zero range", &PortConfig{MinPort: 0, MaxPort: 0}, 0},
			{"single port", &PortConfig{MinPort: 50000, MaxPort: 50000}, 1},
			{"10 ports", &PortConfig{MinPort: 50000, MaxPort: 50009}, 10},
			{"1001 ports", &PortConfig{MinPort: 50000, MaxPort: 51000}, 1001},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				size := tt.config.PortRangeSize()
				if size != tt.expected {
					t.Errorf("PortRangeSize() = %d, want %d", size, tt.expected)
				}
			})
		}
	})
}

func TestNewPortManager(t *testing.T) {
	t.Run("returns nil for nil config", func(t *testing.T) {
		pm := NewPortManager(nil)
		if pm != nil {
			t.Error("NewPortManager(nil) should return nil")
		}
	})

	t.Run("returns nil for Native Mode config", func(t *testing.T) {
		config := DefaultPortConfig()
		pm := NewPortManager(config)
		if pm != nil {
			t.Error("NewPortManager should return nil for Native Mode")
		}
	})

	t.Run("returns manager for Managed Mode config", func(t *testing.T) {
		config := &PortConfig{
			MinPort: 50000,
			MaxPort: 51000,
		}
		pm := NewPortManager(config)
		if pm == nil {
			t.Error("NewPortManager should return non-nil for Managed Mode")
		}
		if !pm.IsActive() {
			t.Error("IsActive() should return true")
		}
	})
}

func TestPortManager_PickPort(t *testing.T) {
	t.Run("returns error for nil manager", func(t *testing.T) {
		var pm *PortManager
		_, err := pm.PickPort()
		if err == nil {
			t.Error("PickPort should return error for nil manager")
		}
	})

	t.Run("picks port from range", func(t *testing.T) {
		config := &PortConfig{
			MinPort: 59000,
			MaxPort: 59010,
		}
		pm := NewPortManager(config)

		port, err := pm.PickPort()
		if err != nil {
			t.Errorf("PickPort failed: %v", err)
		}

		if port < config.MinPort || port > config.MaxPort {
			t.Errorf("port %d is outside range [%d, %d]", port, config.MinPort, config.MaxPort)
		}
	})

	t.Run("returns different ports on multiple calls (randomness)", func(t *testing.T) {
		config := &PortConfig{
			MinPort: 59000,
			MaxPort: 59100,
		}
		pm := NewPortManager(config)

		ports := make(map[int]bool)
		for i := 0; i < 20; i++ {
			port, err := pm.PickPort()
			if err != nil {
				t.Fatalf("PickPort failed: %v", err)
			}
			ports[port] = true
		}

		// With 100 ports and 20 picks, we should have multiple unique ports
		// (statistically very likely)
		if len(ports) < 2 {
			t.Error("PickPort should return varied ports due to randomization")
		}
	})
}

func TestPortManager_IsActive(t *testing.T) {
	t.Run("returns false for nil manager", func(t *testing.T) {
		var pm *PortManager
		if pm.IsActive() {
			t.Error("IsActive() should return false for nil manager")
		}
	})

	t.Run("returns true for configured manager", func(t *testing.T) {
		config := &PortConfig{
			MinPort: 50000,
			MaxPort: 51000,
		}
		pm := NewPortManager(config)
		if !pm.IsActive() {
			t.Error("IsActive() should return true for configured manager")
		}
	})
}

func TestPortManager_GetConfig(t *testing.T) {
	t.Run("returns nil for nil manager", func(t *testing.T) {
		var pm *PortManager
		if pm.GetConfig() != nil {
			t.Error("GetConfig() should return nil for nil manager")
		}
	})

	t.Run("returns config for configured manager", func(t *testing.T) {
		config := &PortConfig{
			MinPort: 50000,
			MaxPort: 51000,
		}
		pm := NewPortManager(config)
		got := pm.GetConfig()
		if got != config {
			t.Error("GetConfig() should return the same config")
		}
	})
}

func TestWithPortRange(t *testing.T) {
	t.Run("configures port range", func(t *testing.T) {
		client, err := New("http://localhost:54345", WithPortRange(50000, 51000))
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		if client.portConfig == nil {
			t.Fatal("portConfig should not be nil")
		}
		if client.portConfig.MinPort != 50000 {
			t.Errorf("MinPort = %d, want 50000", client.portConfig.MinPort)
		}
		if client.portConfig.MaxPort != 51000 {
			t.Errorf("MaxPort = %d, want 51000", client.portConfig.MaxPort)
		}
		if client.portManager == nil {
			t.Error("portManager should be initialized")
		}
		if !client.portManager.IsActive() {
			t.Error("portManager should be active")
		}
	})

	t.Run("zero range keeps Native Mode", func(t *testing.T) {
		client, err := New("http://localhost:54345", WithPortRange(0, 0))
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		if client.portManager != nil {
			t.Error("portManager should be nil in Native Mode")
		}
	})
}

func TestWithPortRetries(t *testing.T) {
	t.Run("configures max retries", func(t *testing.T) {
		client, err := New("http://localhost:54345",
			WithPortRange(50000, 51000),
			WithPortRetries(20),
		)
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		if client.portConfig.MaxRetries != 20 {
			t.Errorf("MaxRetries = %d, want 20", client.portConfig.MaxRetries)
		}
	})
}
