package security

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// SandboxConfig defines sandbox restrictions for MCP server execution
type SandboxConfig struct {
	// Filesystem restrictions
	AllowedDirs  []string `json:"allowed_dirs,omitempty"`
	BlockedDirs  []string `json:"blocked_dirs,omitempty"`
	ReadOnly     bool     `json:"read_only,omitempty"`
	MaxFileSize  int64    `json:"max_file_size,omitempty"` // bytes
	AllowTempDir bool     `json:"allow_temp_dir,omitempty"`

	// Network restrictions
	AllowedHosts       []string `json:"allowed_hosts,omitempty"`
	BlockedHosts       []string `json:"blocked_hosts,omitempty"`
	AllowOutboundHTTP  bool     `json:"allow_outbound_http,omitempty"`
	AllowOutboundHTTPS bool     `json:"allow_outbound_https,omitempty"`

	// Resource limits
	MaxCPUTime   time.Duration `json:"max_cpu_time,omitempty"`
	MaxMemory    int64         `json:"max_memory,omitempty"` // bytes
	MaxProcesses int           `json:"max_processes,omitempty"`

	// Environment restrictions
	SanitizeEnv    bool     `json:"sanitize_env,omitempty"`
	AllowedEnvVars []string `json:"allowed_env_vars,omitempty"`
	BlockedEnvVars []string `json:"blocked_env_vars,omitempty"`

	// Execution restrictions
	AllowShell      bool     `json:"allow_shell,omitempty"`
	AllowedCommands []string `json:"allowed_commands,omitempty"`
	BlockSudo       bool     `json:"block_sudo,omitempty"`
}

// DefaultSandboxConfig returns a secure default sandbox configuration
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		// Filesystem - allow reading but not writing by default
		ReadOnly:     true,
		MaxFileSize:  100 * 1024 * 1024, // 100MB
		AllowTempDir: true,

		// Network - block by default
		AllowOutboundHTTP:  false,
		AllowOutboundHTTPS: false,

		// Resources - reasonable limits
		MaxCPUTime:   30 * time.Second,
		MaxMemory:    512 * 1024 * 1024, // 512MB
		MaxProcesses: 10,

		// Environment - sanitize by default
		SanitizeEnv: true,
		BlockedEnvVars: []string{
			"PATH",
			"HOME",
			"USER",
			"SHELL",
			"SSH_AUTH_SOCK",
			"AWS_ACCESS_KEY_ID",
			"AWS_SECRET_ACCESS_KEY",
		},

		// Execution - restrict by default
		AllowShell: false,
		BlockSudo:  true,
	}
}

// Sandbox defines the interface for MCP server sandboxing
type Sandbox struct {
	config SandboxConfig
}

// NewSandbox creates a new sandbox with the given configuration
func NewSandbox(config SandboxConfig) *Sandbox {
	return &Sandbox{config: config}
}

// Apply applies sandbox restrictions to a context
func (s *Sandbox) Apply(ctx context.Context) (context.Context, error) {
	// Add sandbox context values
	ctx = context.WithValue(ctx, sandboxKey{}, s.config)

	// Set up timeout
	if s.config.MaxCPUTime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.config.MaxCPUTime)
		_ = cancel // We'll handle cancellation via parent context
	}

	return ctx, nil
}

// ValidateTool checks if a tool is allowed in the sandbox
func (s *Sandbox) ValidateTool(name, description string) error {
	// Analyze tool risk
	risk := ClassifyToolRisk(name, description)

	// Check against sandbox restrictions
	switch risk.Level {
	case ToolRiskWrite:
		if s.config.ReadOnly {
			return fmt.Errorf("write operation not allowed in sandbox: %s", name)
		}

	case ToolRiskExec:
		if !s.config.AllowShell {
			return fmt.Errorf("command execution not allowed in sandbox: %s", name)
		}

	case ToolRiskNetwork:
		if !s.config.AllowOutboundHTTP && !s.config.AllowOutboundHTTPS {
			return fmt.Errorf("network operations not allowed in sandbox: %s", name)
		}

	case ToolRiskSystem:
		return fmt.Errorf("system-level operations not allowed: %s", name)

	case ToolRiskDangerous:
		return fmt.Errorf("dangerous operation blocked by sandbox: %s", name)
	}

	return nil
}

// ValidateFilePath checks if a file path is allowed
func (s *Sandbox) ValidateFilePath(path string, write bool) error {
	if write && s.config.ReadOnly {
		return fmt.Errorf("write access denied (sandbox is read-only): %s", path)
	}

	// Check allowed directories
	if len(s.config.AllowedDirs) > 0 {
		allowed := false
		for _, dir := range s.config.AllowedDirs {
			if isSubpath(path, dir) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("access denied (not in allowed directories): %s", path)
		}
	}

	// Check blocked directories
	for _, dir := range s.config.BlockedDirs {
		if isSubpath(path, dir) {
			return fmt.Errorf("access denied (blocked directory): %s", path)
		}
	}

	return nil
}

// ValidateHost checks if a host is allowed for network connections
func (s *Sandbox) ValidateHost(host string, https bool) error {
	if !s.config.AllowOutboundHTTP && !https {
		return fmt.Errorf("outbound HTTP connections not allowed: %s", host)
	}

	if !s.config.AllowOutboundHTTPS && https {
		return fmt.Errorf("outbound HTTPS connections not allowed: %s", host)
	}

	// Check blocked hosts
	for _, blocked := range s.config.BlockedHosts {
		if matchesHost(host, blocked) {
			return fmt.Errorf("host is blocked: %s", host)
		}
	}

	// Check allowed hosts (if specified)
	if len(s.config.AllowedHosts) > 0 {
		allowed := false
		for _, allowedHost := range s.config.AllowedHosts {
			if matchesHost(host, allowedHost) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("host not in allowlist: %s", host)
		}
	}

	return nil
}

// ValidateEnvironment filters environment variables
func (s *Sandbox) ValidateEnvironment(env map[string]string) map[string]string {
	if !s.config.SanitizeEnv {
		return env
	}

	filtered := make(map[string]string)
	for key, value := range env {
		// Check blocked variables
		blocked := false
		for _, blockedVar := range s.config.BlockedEnvVars {
			if key == blockedVar {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}

		// Check allowed variables (if specified)
		if len(s.config.AllowedEnvVars) > 0 {
			allowed := false
			for _, allowedVar := range s.config.AllowedEnvVars {
				if key == allowedVar {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}

		filtered[key] = value
	}

	return filtered
}

// GetResourceLimits returns the resource limits for this sandbox
func (s *Sandbox) GetResourceLimits() (cpu time.Duration, memory int64, processes int) {
	return s.config.MaxCPUTime, s.config.MaxMemory, s.config.MaxProcesses
}

// SandboxFromRisk creates an appropriate sandbox based on risk rating
func SandboxFromRisk(risk RiskScore) *Sandbox {
	config := DefaultSandboxConfig()

	switch risk.Rating {
	case RiskRatingA:
		// Verified servers - allow more access
		config.ReadOnly = false
		config.AllowOutboundHTTPS = true
		config.AllowShell = true

	case RiskRatingB:
		// Community servers - moderate restrictions
		config.ReadOnly = false
		config.AllowOutboundHTTPS = true
		config.AllowShell = false

	case RiskRatingC:
		// Known but unverified - strict read-only
		config.ReadOnly = true
		config.AllowOutboundHTTPS = false
		config.AllowShell = false

	case RiskRatingD:
		// Unknown - maximum restrictions
		config.ReadOnly = true
		config.AllowOutboundHTTP = false
		config.AllowOutboundHTTPS = false
		config.AllowShell = false
		config.SanitizeEnv = true
		config.MaxCPUTime = 10 * time.Second
		config.MaxMemory = 128 * 1024 * 1024 // 128MB

	case RiskRatingF:
		// Blocked - should never connect, but just in case
		config.ReadOnly = true
		config.AllowTempDir = false
		config.AllowOutboundHTTP = false
		config.AllowOutboundHTTPS = false
		config.AllowShell = false
		config.MaxCPUTime = 0
		config.MaxMemory = 0
	}

	return NewSandbox(config)
}

// PlatformSandbox implements platform-specific sandboxing
type PlatformSandbox struct {
	base *Sandbox
}

// NewPlatformSandbox creates a sandbox with platform-specific restrictions
func NewPlatformSandbox(config SandboxConfig) (*PlatformSandbox, error) {
	switch runtime.GOOS {
	case "linux":
		return newLinuxSandbox(config)
	case "darwin":
		return newDarwinSandbox(config)
	default:
		// Generic sandbox for other platforms
		return &PlatformSandbox{base: NewSandbox(config)}, nil
	}
}

// Apply applies platform-specific sandbox restrictions
func (p *PlatformSandbox) Apply(ctx context.Context) (context.Context, error) {
	return p.base.Apply(ctx)
}

// newLinuxSandbox creates a Linux-specific sandbox (seccomp, cgroups, namespaces)
func newLinuxSandbox(config SandboxConfig) (*PlatformSandbox, error) {
	// On Linux we could use:
	// - seccomp-bpf for syscall filtering
	// - cgroups for resource limits
	// - namespaces for isolation
	// For now, use the base sandbox
	return &PlatformSandbox{base: NewSandbox(config)}, nil
}

// newDarwinSandbox creates a macOS-specific sandbox (seatbelt)
func newDarwinSandbox(config SandboxConfig) (*PlatformSandbox, error) {
	// On macOS we could use:
	// - Seatbelt (sandbox profile)
	// - Entitlements
	// For now, use the base sandbox
	return &PlatformSandbox{base: NewSandbox(config)}, nil
}

// SandboxMiddleware wraps tool execution with sandbox checks
type SandboxMiddleware struct {
	sandbox *Sandbox
}

// NewSandboxMiddleware creates a new sandbox middleware
func NewSandboxMiddleware(sandbox *Sandbox) *SandboxMiddleware {
	return &SandboxMiddleware{sandbox: sandbox}
}

// Wrap wraps a tool function with sandbox validation
func (m *SandboxMiddleware) Wrap(toolName, toolDesc string, fn func(context.Context, map[string]interface{}) (interface{}, error)) func(context.Context, map[string]interface{}) (interface{}, error) {
	return func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		// Validate tool is allowed
		if err := m.sandbox.ValidateTool(toolName, toolDesc); err != nil {
			return nil, fmt.Errorf("sandbox violation: %w", err)
		}

		// Apply sandbox context
		sandboxCtx, err := m.sandbox.Apply(ctx)
		if err != nil {
			return nil, err
		}

		// Execute with sandbox context
		return fn(sandboxCtx, args)
	}
}

type sandboxKey struct{}

// GetSandboxConfig retrieves sandbox config from context
func GetSandboxConfig(ctx context.Context) (SandboxConfig, bool) {
	config, ok := ctx.Value(sandboxKey{}).(SandboxConfig)
	return config, ok
}

func isSubpath(path, parent string) bool {
	// Simple subpath check - in production, use filepath.IsLocal or proper path validation
	return len(path) >= len(parent) && path[:len(parent)] == parent
}

func matchesHost(host, pattern string) bool {
	// Simple host matching - exact or wildcard
	if pattern == host {
		return true
	}
	if len(pattern) > 2 && pattern[0] == '*' && pattern[1] == '.' {
		suffix := pattern[2:]
		return len(host) > len(suffix) && host[len(host)-len(suffix):] == suffix
	}
	return false
}
