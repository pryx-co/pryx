package skills

import (
	"context"
	"fmt"
	"os"
	"sync"
)

type UnifiedManager struct {
	mu       sync.RWMutex
	registry *Registry
	opts     Options
	bridges  map[string]*AgentBridge
}

func NewUnifiedManager(opts Options) *UnifiedManager {
	return &UnifiedManager{
		registry: NewRegistry(),
		opts:     opts,
		bridges:  make(map[string]*AgentBridge),
	}
}

func (m *UnifiedManager) Initialize(ctx context.Context) error {
	registry, err := Discover(ctx, m.opts)
	if err != nil {
		return fmt.Errorf("discover skills: %w", err)
	}
	m.registry = registry
	return nil
}

func (m *UnifiedManager) InstallFromURL(ctx context.Context, url string) error {
	result, err := InstallFromURL(ctx, url, m.opts)
	if err != nil {
		return err
	}

	m.registry.Upsert(result.Skill)
	return nil
}

func (m *UnifiedManager) Uninstall(skillID string) error {
	if err := UninstallSkill(skillID, m.opts); err != nil {
		return err
	}

	return nil
}

func (m *UnifiedManager) GetBridge(skillID string) (*AgentBridge, error) {
	m.mu.RLock()
	bridge, exists := m.bridges[skillID]
	m.mu.RUnlock()

	if exists {
		return bridge, nil
	}

	skill, ok := m.registry.Get(skillID)
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", skillID)
	}

	meta := skill.Frontmatter.Metadata.Pryx
	if meta.Type == "" || meta.Type == "tool" {
		return nil, fmt.Errorf("skill %s is a local tool, not a remote agent", skillID)
	}

	endpoint := meta.Endpoint
	if endpoint == "" {
		return nil, fmt.Errorf("skill %s missing endpoint configuration", skillID)
	}

	var authToken string
	var authType string

	if len(meta.Requires.Env) > 0 {
		authToken = os.Getenv(meta.Requires.Env[0])
		authType = meta.Auth
	}

	bridge = NewAgentBridge(skill, endpoint, authToken, authType)

	m.mu.Lock()
	m.bridges[skillID] = bridge
	m.mu.Unlock()

	return bridge, nil
}

func (m *UnifiedManager) ListSkills() []Skill {
	return m.registry.List()
}

func (m *UnifiedManager) GetSkill(id string) (Skill, bool) {
	return m.registry.Get(id)
}

func (m *UnifiedManager) IsAvailable(skillID string) bool {
	skill, ok := m.registry.Get(skillID)
	if !ok {
		return false
	}

	if skill.Source == SourceRemote {
		return true
	}

	return skill.Eligible
}

func (m *UnifiedManager) MetadataSummary() string {
	return m.registry.MetadataSummary()
}
