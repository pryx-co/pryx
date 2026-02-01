package agentbus

import (
	"context"
	"sync"

	"pryx-core/internal/bus"
)

// DetectionManager manages auto-detection of running agents
type DetectionManager struct {
	mu      sync.RWMutex
	bus     *bus.Bus
	logger  *StructuredLogger
	running bool
	stopCh  chan struct{}
}

// NewDetectionManager creates a new detection manager
func NewDetectionManager(b *bus.Bus) *DetectionManager {
	return &DetectionManager{
		bus:    b,
		logger: NewStructuredLogger("detection", "info"),
		stopCh: make(chan struct{}),
	}
}

// Start initializes the detection manager
func (dm *DetectionManager) Start(ctx context.Context) error {
	dm.mu.Lock()
	if dm.running {
		dm.mu.Unlock()
		return nil
	}
	dm.running = true
	dm.mu.Unlock()

	dm.logger.Info("detection manager started", nil)
	dm.bus.Publish(bus.NewEvent("agentbus.detection.started", "", nil))

	return nil
}

// Stop gracefully shuts down the detection manager
func (dm *DetectionManager) Stop(ctx context.Context) error {
	dm.mu.Lock()
	if !dm.running {
		dm.mu.Unlock()
		return nil
	}
	dm.running = false
	dm.mu.Unlock()

	close(dm.stopCh)

	dm.logger.Info("detection manager stopped", nil)
	dm.bus.Publish(bus.NewEvent("agentbus.detection.stopped", "", nil))

	return nil
}

// DetectAll runs detection using all available adapters
func (dm *DetectionManager) DetectAll(ctx context.Context, adapters map[string]AgentAdapter) ([]AgentInfo, error) {
	dm.logger.Info("running detection with all adapters", map[string]interface{}{
		"adapter_count": len(adapters),
	})

	var allAgents []AgentInfo
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, adapter := range adapters {
		wg.Add(1)
		go func(a AgentAdapter) {
			defer wg.Done()

			agents, err := a.Detect(ctx)
			if err != nil {
				dm.logger.Error("detection failed for adapter", map[string]interface{}{
					"protocol": a.Protocol(),
					"error":    err.Error(),
				})
				return
			}

			mu.Lock()
			allAgents = append(allAgents, agents...)
			mu.Unlock()
		}(adapter)
	}

	wg.Wait()

	dm.logger.Info("detection complete", map[string]interface{}{
		"agent_count": len(allAgents),
	})

	return allAgents, nil
}

// DetectProtocol runs detection using a specific protocol adapter
func (dm *DetectionManager) DetectProtocol(ctx context.Context, protocol string, adapters map[string]AgentAdapter) ([]AgentInfo, error) {
	adapter, exists := adapters[protocol]
	if !exists {
		return nil, nil
	}

	dm.logger.Info("running detection for protocol", map[string]interface{}{
		"protocol": protocol,
	})

	agents, err := adapter.Detect(ctx)
	if err != nil {
		dm.logger.Error("detection failed for protocol", map[string]interface{}{
			"protocol": protocol,
			"error":    err.Error(),
		})
		return nil, err
	}

	dm.logger.Info("detection complete for protocol", map[string]interface{}{
		"protocol":    protocol,
		"agent_count": len(agents),
	})

	return agents, nil
}

// DetectFilesystem searches for agents via filesystem indicators
func (dm *DetectionManager) DetectFilesystem(ctx context.Context, paths []string) ([]AgentInfo, error) {
	var agents []AgentInfo

	for _, path := range paths {
		// Search for agent indicators in the path
		// This would be implemented with actual filesystem scanning
		_ = path
	}

	return agents, nil
}

// DetectNetwork discovers agents via network scanning
func (dm *DetectionManager) DetectNetwork(ctx context.Context, ports []int) ([]AgentInfo, error) {
	var agents []AgentInfo

	// Network discovery implementation
	// Would scan specified ports for agent endpoints
	_ = ports

	return agents, nil
}
