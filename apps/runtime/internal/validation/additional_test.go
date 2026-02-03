package validation

import (
	"testing"
)

// Additional validation tests for complete coverage
// These test edge cases and functions that may not have been fully covered

func TestValidatePrivateIPRanges(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// 10.x.x.x - Private Class A
		{"10.0.0.1", "http://10.0.0.1:8080", true},
		{"10.255.255.255", "http://10.255.255.255/api", true},
		{"10.1.2.3", "https://10.1.2.3:443/v1", true},

		// 172.16-31.x.x - Private Class B
		{"172.16.0.1", "http://172.16.0.1:3000", true},
		{"172.31.255.254", "http://172.31.255.254/api", true},
		{"172.20.50.75", "https://172.20.50.75/endpoint", true},

		// 192.168.x.x - Private Class C
		{"192.168.1.1", "http://192.168.1.1:8080", true},
		{"192.168.0.100", "http://192.168.0.100/api", true},
		{"192.168.255.255", "https://192.168.255.255/v1", true},

		// 127.x.x.x - Loopback
		{"127.0.0.1", "http://127.0.0.1:3000", true},
		{"127.0.0.2", "http://127.0.0.2/api", true},

		// ::1 - IPv6 Loopback
		{"::1", "http://[::1]:3000/api", true},

		// fc00: and fe80: - IPv6 Private
		{"fc00::1", "http://[fc00::1]:8080", true},
		{"fe80::1", "http://[fe80::1%eth0]", true},

		// 0.0.0.0 - Wildcard (should be blocked)
		{"0.0.0.0", "http://0.0.0.0:80", true},

		// Valid public IPs should not error (we only check the hostname format)
		{"8.8.8.8", "http://8.8.8.8:8080", false},
		{"1.1.1.1", "https://1.1.1.1:443/dns", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateURL("url", tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestValidateIDEdgeCases(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Valid cases
		{"simple", "test-id", false},
		{"with underscores", "test_id_123", false},
		{"with hyphens", "test-id-456", false},
		{"numbers only", "123456", false},
		{"mixed case", "Test-Id_789", false},
		{"single char", "a", false},
		{"max length 256", generateMaxLengthString(256, "a-z0-9_-"), false},

		// Invalid cases
		{"empty", "", true},
		{"with spaces", "test id", true},
		{"with @", "test@id", true},
		{"with #", "test#id", true},
		{"with $", "test$id", true},
		{"with %", "test%id", true},
		{"with ^", "test^id", true},
		{"with &", "test&id", true},
		{"with *", "test*id", true},
		{"with (", "test(id", true},
		{"with )", "test)id", true},
		{"with +", "test+id", true},
		{"with =", "test=id", true},
		{"with [", "test[id", true},
		{"with ]", "test]id", true},
		{"with {", "test{id", true},
		{"with }", "test}id", true},
		{"with |", "test|id", true},
		{"with \\", "test\\id", true},
		{"with /", "test/id", true},
		{"with ?", "test?id", true},
		{"with !", "test!id", true},
		{"with ~", "test~id", true},
		{"with `", "test`id", true},
		{"with '", "test'id", true},
		{"with \"", "test\"id", true},
		{"with <", "test<id", true},
		{"with >", "test>id", true},
		{"with comma", "test,id", true},
		{"with semicolon", "test;id", true},
		{"with colon", "test:id", true},
		{"unicode chars", "test✓id", true},
		{"chinese chars", "测试ID", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateID("id", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSessionIDFormats(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Empty is valid (optional)
		{"empty", "", false},

		// Valid UUID v4 format
		{"valid v4 - lowercase", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid v4 - uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"valid v4 - mixed", "550e8400-E29b-41d4-a716-446655440000", false},

		// Invalid formats
		{"not uuid", "not-a-valid-uuid", true},
		{"missing hyphens", "550e8400e29b41d4a716446655440000", true},
		{"too short", "550e8400-e29b-41d4-a716-44665544", true},
		{"too long", "550e8400-e29b-41d4-a716-4466554400000", true},
		{"v1 format", "550e8400-e29b-11d4-a716-446655440000", true}, // 11d4 not 41d4
		{"v2 format", "550e8400-e29b-21d4-a716-446655440000", true}, // 21d4 not 41d4
		{"v3 format", "550e8400-e29b-31d4-a716-446655440000", true}, // 31d4 not 41d4
		{"v5 format", "550e8400-e29b-51d4-a716-446655440000", true}, // 51d4 not 41d4
		{"all zeros version", "00000000-0000-0000-0000-000000000000", true},
		{"invalid chars", "550e8400-e29b-41d4-a716-44665544zzzz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateSessionID(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSessionID(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateToolNamePatterns(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Valid tool names (MCP format: namespace.name)
		{"simple", "read", false},
		{"with dots", "filesystem.read", false},
		{"multiple dots", "aws.s3.bucket.list", false},
		{"with hyphens", "git-branch-create", false},
		{"with underscores", "database_query_execute", false},
		{"mixed format", "provider.tool-name.sub_tool", false},
		{"single letter", "a", false},
		{"numbers", "tool123", false},
		{"max length", generateMaxLengthString(256, "a-z0-9._-"), false},

		// Invalid
		{"empty", "", true},
		{"with spaces", "read file", true},
		{"with @", "read@file", true},
		{"with /", "read/file", true},
		{"with \\", "read\\file", true},
		{"with :", "read:file", true},
		{"with #", "read#file", true},
		{"with $", "read$file", true},
		{"with %", "read%file", true},
		{"with ^", "read^file", true},
		{"with &", "read&file", true},
		{"with *", "read*file", true},
		{"with (", "read(file", true},
		{"with )", "read)file", true},
		{"with +", "read+file", true},
		{"with =", "read=file", true},
		{"with [", "read[file", true},
		{"with ]", "read]file", true},
		{"with {", "read{file", true},
		{"with }", "read}file", true},
		{"with |", "read|file", true},
		{"with ?", "read?file", true},
		{"with !", "read!file", true},
		{"with ~", "read~file", true},
		{"with `", "read`file", true},
		{"with '", "read'file", true},
		{"with \"", "read\"file", true},
		{"with <", "read<file", true},
		{"with >", "read>file", true},
		{"chinese chars", "文件系统.读取", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateToolName(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToolName(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateFilePathSecurity(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Valid relative paths
		{"simple file", "file.txt", false},
		{"with subdir", "dir/subdir/file.txt", false},
		{"many levels", "a/b/c/d/e/f/g.txt", false},
		{"with spaces", "my documents/file.txt", false},
		{"with hyphens", "my-project/file.txt", false},
		{"with underscores", "my_project/file.txt", false},
		{"dots in name", "file.name.txt", false},
		{"hidden file", ".hidden", false},
		{"hidden with path", "path/.hidden", false},
		{"trailing slash", "path/", false},

		// Path traversal patterns - these should be blocked
		{"single dot traversal", "../etc/passwd", true},
		{"double dot traversal", "../../../etc/passwd", true},
		{"encoded traversal", "%2e%2e/passwd", false}, // URL encoded - not our concern
		{"traversal with slashes", "./../../../etc/passwd", true},

		// Null byte injection
		{"null byte", "file.txt\x00", true},
		{"null in middle", "file\x00name.txt", true},

		// Absolute paths should be blocked (unix-style)
		{"absolute unix", "/etc/passwd", true},
		{"absolute with drive", "/C:/Windows/System32", true},
		{"root reference", "/home/user/file", true},

		// Invalid - empty
		{"empty", "", true},
		{"whitespace only", "   ", true},

		// Edge cases
		{"self reference", ".", false}, // Current directory reference
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateFilePath("path", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeStringEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Normal cases
		{"normal text", "hello world", "hello world"},
		{"empty string", "", ""},

		// Null byte removal
		{"single null byte", "hello\x00world", "helloworld"},
		{"multiple null bytes", "hel\x00lo\x00wor\x00ld", "helloworld"},
		{"null at start", "\x00hello", "hello"},
		{"null at end", "hello\x00", "hello"},

		// Control character removal
		{"tab preserved", "hello\tworld", "hello\tworld"},
		{"newline preserved", "hello\nworld", "hello\nworld"},
		{"carriage return preserved", "hello\r\nworld", "hello\r\nworld"},
		{"control chars removed", "hello\x01\x02\x03world", "helloworld"},
		{"mix of control and whitespace", "hello\x01 \x02\n", "hello \n"},

		// Extended ASCII
		{"extended ascii", "café", "café"}, // Valid UTF-8
		{"emoji", "Hello 🌍", "Hello 🌍"},    // Unicode preserved
		{"chinese", "你好世界", "你好世界"},        // Chinese preserved

		// Special edge cases
		{"all null bytes", "\x00\x00\x00", ""},
		{"all control chars", "\x01\x02\x03\x04\x05", ""},
		{"null and valid", "\x00hello\x00world\x00", "helloworld"},
		{"null in unicode", "café\x00", "café"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateCommandInjection(t *testing.T) {
	v := NewValidator()

	// Comprehensive command injection patterns to block
	dangerousPatterns := []struct {
		name     string
		command  string
		shouldBe bool // true = should be blocked
	}{
		// Shell metacharacters
		{"semicolon", "ls; rm -rf /", true},
		{"double ampersand", "echo hello && rm -rf /", true},
		{"double pipe", "cat file.txt || echo fail", true},
		{"single pipe", "cat file.txt | grep pattern", true},
		{"backtick command", "echo `whoami`", true},
		{"$() subshell", "echo $(whoami)", true},
		{"${} expansion", "echo ${HOME}", true},
		{"output redirect", "ls > /tmp/output", true},
		{"append redirect", "ls >> /tmp/output", true},
		{"input redirect", "cat < /tmp/input", true},
		{"heredoc", "cat <<EOF\nhello\nEOF", true},

		// Command separators with spaces
		{"semicolon with spaces", "ls ; echo hello", true},
		{"pipe with spaces", "ls | cat", true},

		// Null byte injection
		{"null byte in command", "ls\x00 -la", true},

		// Valid commands (should pass)
		{"simple command", "ls -la", false},
		{"command with args", "git commit -m 'message'", false},
		{"path with spaces", "/path/to/file with spaces.txt", false},
		{"command with dots", "python script.py", false},
		{"command with dashes", "my-script --option value", false},
		{"command with underscores", "my_command --option", false},
		{"npm script", "npm run build", false},
		{"docker command", "docker ps -a", false},
		{"curl command", "curl -s https://api.example.com", false},
		{"git command", "git status", false},
		{"make command", "make install", false},
	}

	for _, tt := range dangerousPatterns {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateCommand(tt.command)
			hasError := err != nil

			if hasError != tt.shouldBe {
				if tt.shouldBe {
					t.Errorf("ValidateCommand(%q) should have been blocked but passed", tt.command)
				} else {
					t.Errorf("ValidateCommand(%q) should have passed but got error: %v", tt.command, err)
				}
			}
		})
	}
}

// Helper function to generate strings of specific length
func generateMaxLengthString(maxLen int, allowedChars string) string {
	result := ""
	for len(result) < maxLen {
		result += allowedChars
	}
	if len(result) > maxLen {
		result = result[:maxLen]
	}
	return result
}
