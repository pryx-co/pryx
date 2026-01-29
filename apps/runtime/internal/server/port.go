package server

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

// GetAvailablePort finds an available port on the system
func GetAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// WritePortFile writes the port number to ~/.pryx/runtime.port
func WritePortFile(port int) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pryxDir := filepath.Join(homeDir, ".pryx")
	if err := os.MkdirAll(pryxDir, 0755); err != nil {
		return fmt.Errorf("failed to create .pryx directory: %w", err)
	}

	portFile := filepath.Join(pryxDir, "runtime.port")
	content := fmt.Sprintf("%d", port)

	if err := os.WriteFile(portFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write port file: %w", err)
	}

	return nil
}

// ReadPortFile reads the port number from ~/.pryx/runtime.port
func ReadPortFile() (int, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, fmt.Errorf("failed to get home directory: %w", err)
	}

	portFile := filepath.Join(homeDir, ".pryx", "runtime.port")
	content, err := os.ReadFile(portFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read port file: %w", err)
	}

	port, err := strconv.Atoi(string(content))
	if err != nil {
		return 0, fmt.Errorf("invalid port in file: %w", err)
	}

	return port, nil
}

// CleanupPortFile removes the port file (call on shutdown)
func CleanupPortFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	portFile := filepath.Join(homeDir, ".pryx", "runtime.port")
	return os.Remove(portFile)
}
