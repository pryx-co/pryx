package universal

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Detector manages auto-detection of running agents
type Detector struct {
	mu      sync.RWMutex
	ports   []int
	running bool
	stopCh  chan struct{}
}

// NewDetector creates a new detector
func NewDetector(scanPorts []int) *Detector {
	if len(scanPorts) == 0 {
		scanPorts = []int{
			18789, // OpenClaw default port
			8080,  // HTTP agents
			3000,  // Common development port
			4000,  // Alternative WebSocket port
		}
	}

	return &Detector{
		ports:  scanPorts,
		stopCh: make(chan struct{}),
	}
}

// Start initializes the detector
func (d *Detector) Start(ctx context.Context) {
	d.running = true
}

// Stop gracefully shuts down the detector
func (d *Detector) Stop(ctx context.Context) {
	d.running = false
	close(d.stopCh)
}

func safeProtocol(protocols []string) string {
	if len(protocols) > 0 {
		return protocols[0]
	}
	return "stdio"
}

// DetectAll scans for agents using all methods
func (d *Detector) DetectAll(ctx context.Context) ([]DetectedAgent, error) {
	var agents []DetectedAgent
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Scan ports
	wg.Add(1)
	go func() {
		defer wg.Done()
		portAgents, err := d.scanPorts(ctx)
		if err == nil {
			mu.Lock()
			agents = append(agents, portAgents...)
			mu.Unlock()
		}
	}()

	// Scan filesystem
	wg.Add(1)
	go func() {
		defer wg.Done()
		fsAgents, err := d.scanFilesystem(ctx)
		if err == nil {
			mu.Lock()
			agents = append(agents, fsAgents...)
			mu.Unlock()
		}
	}()

	wg.Wait()
	return agents, nil
}

// DetectByProtocol scans for agents using a specific protocol
func (d *Detector) DetectByProtocol(ctx context.Context, protocol string) ([]DetectedAgent, error) {
	switch protocol {
	case "websocket":
		return d.scanPorts(ctx)
	case "filesystem":
		return d.scanFilesystem(ctx)
	default:
		return d.DetectAll(ctx)
	}
}

// scanPorts scans common ports for agent connections
func (d *Detector) scanPorts(ctx context.Context) ([]DetectedAgent, error) {
	var agents []DetectedAgent

	for _, port := range d.ports {
		agents = append(agents, d.checkPort(ctx, port)...)
	}

	return agents, nil
}

// checkPort checks if a port has an agent running
func (d *Detector) checkPort(ctx context.Context, port int) []DetectedAgent {
	var agents []DetectedAgent

	// Try common local IPs
	addresses := []string{
		fmt.Sprintf("127.0.0.1:%d", port),
		fmt.Sprintf("localhost:%d", port),
	}

	for _, addr := range addresses {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			continue
		}
		conn.Close()

		// Detect agent type based on port
		agent := d.detectByPort(port, addr)
		if agent.Confidence > 0 {
			agents = append(agents, agent)
		}
	}

	return agents
}

// detectByPort identifies agent type based on port number
func (d *Detector) detectByPort(port int, addr string) DetectedAgent {
	agent := DetectedAgent{
		DetectionMethod: "port",
		Confidence:      0.5,
	}

	switch port {
	case 18789:
		// OpenClaw WebSocket gateway
		agent.AgentInfo = AgentInfo{
			Identity: AgentIdentity{
				ID:      fmt.Sprintf("openclaw-%d", port),
				Name:    "OpenClaw Agent",
				Version: "unknown",
			},
			Protocol: "websocket",
			Endpoint: EndpointInfo{
				Type: "websocket",
				URL:  fmt.Sprintf("ws://%s", addr),
				Host: "127.0.0.1",
				Port: port,
			},
			Capabilities: []string{"messaging", "tools", "sessions"},
			HealthStatus: "unknown",
		}
		agent.Confidence = 0.9
		agent.HandshakeData = map[string]string{
			"handshake": "connect.challenge",
		}

	case 8080, 3000, 4000:
		// Generic HTTP/WebSocket agent
		agent.AgentInfo = AgentInfo{
			Identity: AgentIdentity{
				ID:      fmt.Sprintf("http-agent-%d", port),
				Name:    fmt.Sprintf("HTTP Agent on port %d", port),
				Version: "unknown",
			},
			Protocol: "http",
			Endpoint: EndpointInfo{
				Type: "http",
				URL:  fmt.Sprintf("http://%s", addr),
				Host: "127.0.0.1",
				Port: port,
			},
			Capabilities: []string{"http"},
			HealthStatus: "unknown",
		}
		agent.Confidence = 0.7
	}

	return agent
}

// scanFilesystem searches for agent manifests
func (d *Detector) scanFilesystem(ctx context.Context) ([]DetectedAgent, error) {
	var agents []DetectedAgent

	// Common locations for agent manifests
	searchPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".pryx", "agents"),
		filepath.Join(os.Getenv("HOME"), ".config", "pryx", "agents"),
		"/usr/local/share/pryx/agents",
	}

	for _, path := range searchPaths {
		pathAgents, err := d.scanPath(ctx, path)
		if err == nil {
			agents = append(agents, pathAgents...)
		}
	}

	return agents, nil
}

// scanPath searches a directory for agent manifests
func (d *Detector) scanPath(ctx context.Context, path string) ([]DetectedAgent, error) {
	var agents []DetectedAgent

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		agentDir := filepath.Join(path, entry.Name())
		agent := d.parseManifest(agentDir)
		if agent.Confidence > 0 {
			agents = append(agents, agent)
		}
	}

	return agents, nil
}

// parseManifest parses an agent manifest file
func (d *Detector) parseManifest(agentDir string) DetectedAgent {
	manifestPaths := []string{
		filepath.Join(agentDir, "agent.json"),
		filepath.Join(agentDir, "package.json"),
		filepath.Join(agentDir, "agent.yaml"),
	}

	for _, manifestPath := range manifestPaths {
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		// Try to parse as AgentPackage
		var pkg AgentPackage
		if err := json.Unmarshal(data, &pkg); err == nil {
			return DetectedAgent{
				AgentInfo: AgentInfo{
					Identity: AgentIdentity{
						ID:        pkg.Name,
						Name:      pkg.Name,
						Version:   pkg.Version,
						Namespace: "local",
					},
					Protocol:     safeProtocol(pkg.Protocols),
					Capabilities: pkg.Capabilities,
					HealthStatus: "unknown",
					Metadata:     pkg.Metadata,
				},
				DetectionMethod: "filesystem",
				Confidence:      0.8,
			}
		}
	}

	return DetectedAgent{
		DetectionMethod: "filesystem",
		Confidence:      0,
	}
}
