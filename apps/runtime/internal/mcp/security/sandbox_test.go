package security

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDefaultSandboxConfig(t *testing.T) {
	config := DefaultSandboxConfig()

	if !config.ReadOnly {
		t.Error("Expected sandbox to be read-only by default")
	}

	if config.AllowOutboundHTTP {
		t.Error("Expected HTTP to be blocked by default")
	}

	if config.AllowOutboundHTTPS {
		t.Error("Expected HTTPS to be blocked by default")
	}

	if config.MaxMemory != 512*1024*1024 {
		t.Errorf("Expected 512MB memory limit, got %d", config.MaxMemory)
	}

	if config.MaxCPUTime != 30*time.Second {
		t.Errorf("Expected 30s CPU limit, got %v", config.MaxCPUTime)
	}

	if !config.SanitizeEnv {
		t.Error("Expected environment sanitization by default")
	}
}

func TestNewSandbox(t *testing.T) {
	config := DefaultSandboxConfig()
	sandbox := NewSandbox(config)

	if sandbox == nil {
		t.Fatal("NewSandbox returned nil")
	}

	if sandbox.config.ReadOnly != config.ReadOnly {
		t.Error("Sandbox config not set correctly")
	}
}

func TestSandboxApply(t *testing.T) {
	config := DefaultSandboxConfig()
	config.MaxCPUTime = 1 * time.Second
	sandbox := NewSandbox(config)

	ctx := context.Background()
	newCtx, err := sandbox.Apply(ctx)

	if err != nil {
		t.Errorf("Apply returned error: %v", err)
	}

	if newCtx == nil {
		t.Error("Apply returned nil context")
	}

	// Check that sandbox config is in context
	sandboxConfig, ok := GetSandboxConfig(newCtx)
	if !ok {
		t.Error("Sandbox config not found in context")
	}

	if sandboxConfig.ReadOnly != config.ReadOnly {
		t.Error("Sandbox config in context doesn't match")
	}
}

func TestSandboxValidateTool(t *testing.T) {
	tests := []struct {
		name        string
		readOnly    bool
		allowShell  bool
		toolName    string
		description string
		shouldError bool
	}{
		{
			name:        "Read tool in read-only sandbox",
			readOnly:    true,
			toolName:    "read_file",
			description: "Read a file",
			shouldError: false,
		},
		{
			name:        "Write tool in read-only sandbox",
			readOnly:    true,
			toolName:    "write_file",
			description: "Write to a file",
			shouldError: true,
		},
		{
			name:        "Exec tool when shell disabled",
			readOnly:    false,
			allowShell:  false,
			toolName:    "execute",
			description: "Execute a shell command",
			shouldError: true,
		},
		{
			name:        "Exec tool when shell enabled",
			readOnly:    false,
			allowShell:  true,
			toolName:    "execute",
			description: "Execute a shell command",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSandboxConfig()
			config.ReadOnly = tt.readOnly
			config.AllowShell = tt.allowShell
			sandbox := NewSandbox(config)

			err := sandbox.ValidateTool(tt.toolName, tt.description)
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestSandboxValidateFilePath(t *testing.T) {
	tests := []struct {
		name        string
		allowedDirs []string
		blockedDirs []string
		readOnly    bool
		path        string
		write       bool
		shouldError bool
	}{
		{
			name:        "Read in allowed directory",
			allowedDirs: []string{"/home/user"},
			path:        "/home/user/file.txt",
			write:       false,
			shouldError: false,
		},
		{
			name:        "Read in blocked directory",
			blockedDirs: []string{"/etc"},
			path:        "/etc/passwd",
			write:       false,
			shouldError: true,
		},
		{
			name:        "Write in read-only sandbox",
			readOnly:    true,
			path:        "/tmp/file.txt",
			write:       true,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSandboxConfig()
			config.AllowedDirs = tt.allowedDirs
			config.BlockedDirs = tt.blockedDirs
			config.ReadOnly = tt.readOnly
			sandbox := NewSandbox(config)

			err := sandbox.ValidateFilePath(tt.path, tt.write)
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestSandboxValidateHost(t *testing.T) {
	tests := []struct {
		name       string
		allowHTTP  bool
		allowHTTPS bool
		allowed    []string
		blocked    []string
		host       string
		https      bool
		shouldErr  bool
	}{
		{
			name:       "HTTPS allowed",
			allowHTTPS: true,
			host:       "api.example.com",
			https:      true,
			shouldErr:  false,
		},
		{
			name:      "HTTPS blocked",
			host:      "api.example.com",
			https:     true,
			shouldErr: true,
		},
		{
			name:       "HTTP not allowed",
			allowHTTP:  false,
			allowHTTPS: true,
			host:       "api.example.com",
			https:      false,
			shouldErr:  true,
		},
		{
			name:       "Blocked host",
			allowHTTPS: true,
			blocked:    []string{"evil.com"},
			host:       "evil.com",
			https:      true,
			shouldErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSandboxConfig()
			config.AllowOutboundHTTP = tt.allowHTTP
			config.AllowOutboundHTTPS = tt.allowHTTPS
			config.AllowedHosts = tt.allowed
			config.BlockedHosts = tt.blocked
			sandbox := NewSandbox(config)

			err := sandbox.ValidateHost(tt.host, tt.https)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestSandboxValidateEnvironment(t *testing.T) {
	config := DefaultSandboxConfig()
	config.SanitizeEnv = true
	config.BlockedEnvVars = []string{"SECRET", "PASSWORD"}
	config.AllowedEnvVars = []string{"PATH", "HOME"}

	sandbox := NewSandbox(config)

	env := map[string]string{
		"PATH":     "/usr/bin",
		"HOME":     "/home/user",
		"SECRET":   "hidden",
		"PASSWORD": "secret123",
		"CUSTOM":   "value",
	}

	filtered := sandbox.ValidateEnvironment(env)

	// PATH and HOME should be present
	if _, ok := filtered["PATH"]; !ok {
		t.Error("PATH should be in filtered env")
	}

	// SECRET and PASSWORD should be removed
	if _, ok := filtered["SECRET"]; ok {
		t.Error("SECRET should be filtered out")
	}

	// CUSTOM should be filtered (not in allowed list)
	if _, ok := filtered["CUSTOM"]; ok {
		t.Error("CUSTOM should be filtered out (not in allowed list)")
	}
}

func TestSandboxFromRisk(t *testing.T) {
	tests := []struct {
		rating         RiskRating
		expectReadOnly bool
		expectHTTPS    bool
		expectShell    bool
	}{
		{RiskRatingA, false, true, true},
		{RiskRatingB, false, true, false},
		{RiskRatingC, true, false, false},
		{RiskRatingD, true, false, false},
		{RiskRatingF, true, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.rating), func(t *testing.T) {
			risk := RiskScore{Rating: tt.rating}
			sandbox := SandboxFromRisk(risk)

			if sandbox.config.ReadOnly != tt.expectReadOnly {
				t.Errorf("Expected ReadOnly=%v, got %v", tt.expectReadOnly, sandbox.config.ReadOnly)
			}
			if sandbox.config.AllowOutboundHTTPS != tt.expectHTTPS {
				t.Errorf("Expected HTTPS=%v, got %v", tt.expectHTTPS, sandbox.config.AllowOutboundHTTPS)
			}
			if sandbox.config.AllowShell != tt.expectShell {
				t.Errorf("Expected Shell=%v, got %v", tt.expectShell, sandbox.config.AllowShell)
			}
		})
	}
}

func TestIsSubpath(t *testing.T) {
	tests := []struct {
		path   string
		parent string
		result bool
	}{
		{"/home/user/file.txt", "/home/user", true},
		{"/home/user/docs/file.txt", "/home/user", true},
		{"/home/other/file.txt", "/home/user", false},
		{"/home/user", "/home/user", true},
		{"/home/us", "/home/user", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isSubpath(tt.path, tt.parent)
			if result != tt.result {
				t.Errorf("Expected %v, got %v", tt.result, result)
			}
		})
	}
}

func TestMatchesHost(t *testing.T) {
	tests := []struct {
		host    string
		pattern string
		result  bool
	}{
		{"example.com", "example.com", true},
		{"api.example.com", "*.example.com", true},
		{"example.com", "*.example.com", false},
		{"deep.sub.example.com", "*.example.com", true},
		{"other.com", "example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.host+"_"+tt.pattern, func(t *testing.T) {
			result := matchesHost(tt.host, tt.pattern)
			if result != tt.result {
				t.Errorf("Expected %v, got %v", tt.result, result)
			}
		})
	}
}

func TestSandboxMiddleware(t *testing.T) {
	config := DefaultSandboxConfig()
	sandbox := NewSandbox(config)
	middleware := NewSandboxMiddleware(sandbox)

	executed := false
	fn := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		executed = true
		return "result", nil
	}

	wrapped := middleware.Wrap("safe_tool", "A safe tool", fn)

	ctx := context.Background()
	_, err := wrapped(ctx, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !executed {
		t.Error("Wrapped function was not executed")
	}

	// Test with blocked tool
	config.ReadOnly = true
	sandbox2 := NewSandbox(config)
	middleware2 := NewSandboxMiddleware(sandbox2)

	wrapped2 := middleware2.Wrap("write_file", "Write to file", fn)
	_, err = wrapped2(ctx, nil)

	if err == nil {
		t.Error("Expected error for blocked tool")
	}

	if !strings.Contains(err.Error(), "sandbox violation") {
		t.Errorf("Expected sandbox violation error, got: %v", err)
	}
}
