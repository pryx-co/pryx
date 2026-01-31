package security

import (
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()

	if v == nil {
		t.Fatal("NewValidator returned nil")
	}

	if !v.requireHTTPS {
		t.Error("Expected HTTPS requirement by default")
	}

	if v.allowSelfSigned {
		t.Error("Expected self-signed certs to be disallowed by default")
	}
}

func TestValidatorWithOptions(t *testing.T) {
	v := NewValidator(
		WithLocalhost(true),
		WithHTTPSRequirement(false),
		WithSelfSignedCerts(true),
		WithAllowlist([]string{"example.com"}),
		WithBlocklist([]string{"evil.com"}),
	)

	if !v.allowLocalhost {
		t.Error("Expected localhost to be allowed")
	}

	if v.requireHTTPS {
		t.Error("Expected HTTPS requirement to be disabled")
	}

	if !v.allowSelfSigned {
		t.Error("Expected self-signed certs to be allowed")
	}

	if len(v.allowlist) != 1 || v.allowlist[0] != "example.com" {
		t.Error("Expected custom allowlist")
	}

	if len(v.blocklist) != 1 || v.blocklist[0] != "evil.com" {
		t.Error("Expected custom blocklist")
	}
}

func TestValidatorValidate(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name          string
		url           string
		transport     string
		expectValid   bool
		expectHTTPS   bool
		expectBlocked bool
	}{
		{
			name:        "Valid HTTPS URL",
			url:         "https://api.example.com",
			transport:   "http",
			expectValid: true,
			expectHTTPS: true,
		},
		{
			name:        "HTTP URL rejected by default",
			url:         "http://api.example.com",
			transport:   "http",
			expectValid: false,
			expectHTTPS: false,
		},
		{
			name:          "Blocked localhost",
			url:           "http://localhost:8080",
			transport:     "http",
			expectValid:   false,
			expectBlocked: true,
		},
		{
			name:        "Bundled transport",
			url:         "bundled",
			transport:   "bundled",
			expectValid: true,
		},
		{
			name:        "Invalid URL",
			url:         "://invalid-url",
			transport:   "http",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.Validate(tt.url, tt.transport)

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, result.Valid)
			}

			if result.Security.HTTPS != tt.expectHTTPS {
				t.Errorf("Expected HTTPS=%v, got %v", tt.expectHTTPS, result.Security.HTTPS)
			}

			if tt.expectBlocked && result.Security.DomainBlocked {
				t.Error("Expected domain to be blocked")
			}
		})
	}
}

func TestValidatorValidateWithTools(t *testing.T) {
	v := NewValidator()

	tools := []ToolInfo{
		{Name: "read_file", Description: "Read a file"},
		{Name: "write_file", Description: "Write a file"},
	}

	result := v.ValidateWithTools("https://example.com", "http", tools)

	if !result.Valid {
		t.Error("Expected valid result")
	}

	if !result.RiskScore.RequiresApproval {
		t.Error("Expected approval required due to write tool")
	}

	hasWriteWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "write") || strings.Contains(w, "Write") {
			hasWriteWarning = true
			break
		}
	}
	if !hasWriteWarning {
		t.Error("Expected warning about write capability")
	}
}

func TestCheckDomain(t *testing.T) {
	v := NewValidator(
		WithAllowlist([]string{"*.example.com", "trusted.org"}),
		WithBlocklist([]string{"evil.com", "malware"}),
	)

	tests := []struct {
		domain        string
		expectAllowed bool
		expectBlocked bool
	}{
		{"api.example.com", true, false},
		{"sub.api.example.com", true, false},
		{"trusted.org", true, false},
		{"evil.com", false, true},
		{"sub.evil.com", false, true}, // Subdomain of blocked domain should also be blocked
		{"malware-site.com", false, true},
		{"unknown.com", false, false}, // Not in allowlist
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			allowed, blocked := v.CheckDomain(tt.domain)
			if allowed != tt.expectAllowed {
				t.Errorf("Expected allowed=%v, got %v", tt.expectAllowed, allowed)
			}
			if blocked != tt.expectBlocked {
				t.Errorf("Expected blocked=%v, got %v", tt.expectBlocked, blocked)
			}
		})
	}
}

func TestMatchesDomainPattern(t *testing.T) {
	tests := []struct {
		domain  string
		pattern string
		matches bool
	}{
		{"example.com", "example.com", true},
		{"example.com", "other.com", false},
		{"api.example.com", "*.example.com", true},
		{"example.com", "*.example.com", false},
		{"deep.sub.example.com", "*.example.com", true},
		{"test123.com", "regex:test[0-9]+.com", true},
		{"testabc.com", "regex:test[0-9]+.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.domain+"_"+tt.pattern, func(t *testing.T) {
			result := matchesDomainPattern(tt.domain, tt.pattern)
			if result != tt.matches {
				t.Errorf("Expected match=%v, got %v", tt.matches, result)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"https://example.com", "https://example.com", false},
		{"http://example.com/path/", "http://example.com/path", false},
		{"example.com", "https://example.com", false},
		{"bundled", "bundled://localhost", false},
		{"://invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := v.normalizeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error=%v, got err=%v", tt.wantErr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDefaultAllowlist(t *testing.T) {
	allowlist := DefaultAllowlist()
	if len(allowlist) == 0 {
		t.Error("Expected non-empty default allowlist")
	}

	// Check for expected trusted domains
	hasGithub := false
	for _, domain := range allowlist {
		if domain == "*.github.com" {
			hasGithub = true
			break
		}
	}
	if !hasGithub {
		t.Error("Expected github.com in default allowlist")
	}
}

func TestDefaultBlocklist(t *testing.T) {
	blocklist := DefaultBlocklist()
	if len(blocklist) == 0 {
		t.Error("Expected non-empty default blocklist")
	}
}
