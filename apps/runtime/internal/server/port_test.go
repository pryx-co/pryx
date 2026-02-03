package server_test

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPortAlreadyInUse tests port conflict handling
func TestPortAlreadyInUse(t *testing.T) {
	t.Run("detect_port_in_use", func(t *testing.T) {
		// Find an available port
		listener, err := net.Listen("tcp", ":0")
		require.NoError(t, err)
		defer listener.Close()

		addr := listener.Addr().(*net.TCPAddr)
		port := addr.Port

		t.Logf("Using port: %d", port)

		// Try to listen on the same port - should fail
		_, err = net.Listen("tcp", ":"+itow(port))
		assert.Error(t, err, "Should fail when port is already in use")

		if err != nil {
			assert.Contains(t, err.Error(), "address already in use",
				"Error should mention address in use")
		}
	})

	t.Run("port_selection_strategy", func(t *testing.T) {
		ports := []int{0, 3000, 8080, 0}

		for _, port := range ports {
			addr := ":" + itow(port)
			listener, err := net.Listen("tcp", addr)
			require.NoError(t, err)
			defer listener.Close()

			boundPort := listener.Addr().(*net.TCPAddr).Port
			t.Logf("Requested port %d, got %d", port, boundPort)

			if port == 0 {
				assert.Greater(t, boundPort, 0,
					"Port 0 should bind to random available port")
			} else {
				assert.Equal(t, port, boundPort,
					"Should bind to requested port if available")
			}
		}
	})

	t.Run("random_port_selection", func(t *testing.T) {
		ports := make(map[int]bool)

		// Try to get multiple random ports
		for i := 0; i < 10; i++ {
			listener, err := net.Listen("tcp", ":0")
			require.NoError(t, err)
			defer listener.Close()

			port := listener.Addr().(*net.TCPAddr).Port
			ports[port] = true

			t.Logf("Random port %d: %d", i+1, port)
		}

		assert.Len(t, ports, 10, "Should get unique ports")
	})
}

// TestPortConflictRecovery tests recovery from port conflicts
func TestPortConflictRecovery(t *testing.T) {
	t.Run("try_multiple_ports", func(t *testing.T) {
		basePort := int(20000 + (time.Now().Unix() % 1000))
		attempts := []int{
			basePort,
			basePort + 1,
			basePort + 2,
			0, // random
		}

		boundPort := 0
		for _, port := range attempts {
			addr := ":" + itow(port)
			listener, err := net.Listen("tcp", addr)
			if err == nil {
				defer listener.Close()
				boundPort = listener.Addr().(*net.TCPAddr).Port
				t.Logf("Successfully bound to port %d (requested: %d)", boundPort, port)
				break
			}
			t.Logf("Port %d unavailable: %v", port, err)
		}

		assert.Greater(t, boundPort, 0, "Should eventually find available port")
	})

	t.Run("concurrent_port_attempts", func(t *testing.T) {
		attempts := 5

		// Try to bind to the same random port range concurrently
		for i := 0; i < attempts; i++ {
			go func(attemptNum int) {
				listener, err := net.Listen("tcp", ":0")
				if err == nil {
					defer listener.Close()
					port := listener.Addr().(*net.TCPAddr).Port
					t.Logf("Goroutine %d got port %d", attemptNum, port)
				}
			}(i)
		}

		// Give goroutines time to complete
		time.Sleep(100 * time.Millisecond)

		t.Logf("Concurrent port binding test completed")
	})
}

// TestPortRangeValidation tests valid port ranges
func TestPortRangeValidation(t *testing.T) {
	t.Run("well_known_ports", func(t *testing.T) {
		wellKnownPorts := []int{80, 443, 22, 25, 53}

		for _, port := range wellKnownPorts {
			addr := ":" + itow(port)
			listener, err := net.Listen("tcp", addr)
			if err == nil {
				defer listener.Close()
				t.Logf("Port %d available (well-known)", port)
			} else {
				t.Logf("Port %d in use (well-known)", port)
			}
		}
	})

	t.Run("ephemeral_ports", func(t *testing.T) {
		ephemeralRange := []int{32768, 60999, 49152, 65535}

		for _, port := range ephemeralRange {
			addr := ":" + itow(port)
			listener, err := net.Listen("tcp", addr)
			if err == nil {
				defer listener.Close()
				t.Logf("Port %d available (ephemeral)", port)
			} else {
				t.Logf("Port %d may be in use (ephemeral)", port)
			}
		}
	})

	t.Run("invalid_ports", func(t *testing.T) {
		invalidPorts := []int{-1, 0, 65536, 100000}

		for _, port := range invalidPorts {
			addr := ":" + itow(port)
			_, err := net.Listen("tcp", addr)
			if err != nil {
				t.Logf("Port %d correctly rejected: %v", port, err)
			}
		}
	})
}

// Helper function for int to string conversion
func itow(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
