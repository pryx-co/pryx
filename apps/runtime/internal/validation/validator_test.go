package validation

import (
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator() returned nil")
	}
	if v.maxStringLength != 100000 {
		t.Errorf("expected maxStringLength to be 100000, got %d", v.maxStringLength)
	}
	if v.maxRequestSize != 10*1024*1024 {
		t.Errorf("expected maxRequestSize to be 10MB, got %d", v.maxRequestSize)
	}
}

func TestValidateString(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		field   string
		value   string
		opts    []StringOption
		wantErr bool
	}{
		{
			name:    "valid string",
			field:   "test",
			value:   "hello world",
			wantErr: false,
		},
		{
			name:    "empty string allowed",
			field:   "test",
			value:   "",
			wantErr: false,
		},
		{
			name:    "empty string not allowed",
			field:   "test",
			value:   "",
			opts:    []StringOption{AllowEmpty(false)},
			wantErr: true,
		},
		{
			name:    "whitespace only not allowed",
			field:   "test",
			value:   "   ",
			opts:    []StringOption{AllowEmpty(false)},
			wantErr: true,
		},
		{
			name:    "min length",
			field:   "test",
			value:   "ab",
			opts:    []StringOption{MinLength(3)},
			wantErr: true,
		},
		{
			name:    "max length exceeded",
			field:   "test",
			value:   strings.Repeat("a", 11),
			opts:    []StringOption{MaxLength(10)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateString(tt.field, tt.value, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{
			name:    "valid ID",
			field:   "id",
			value:   "test-id_123",
			wantErr: false,
		},
		{
			name:    "empty ID",
			field:   "id",
			value:   "",
			wantErr: true,
		},
		{
			name:    "ID with spaces",
			field:   "id",
			value:   "test id",
			wantErr: true,
		},
		{
			name:    "ID with special chars",
			field:   "id",
			value:   "test@id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateID(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSessionID(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "empty session ID",
			value:   "",
			wantErr: false,
		},
		{
			name:    "valid UUID v4",
			value:   "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "invalid UUID",
			value:   "not-a-uuid",
			wantErr: true,
		},
		{
			name:    "UUID v1",
			value:   "550e8400-e29b-11d4-a716-446655440000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateSessionID(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSessionID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateToolName(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid tool name",
			value:   "filesystem.read",
			wantErr: false,
		},
		{
			name:    "empty tool name",
			value:   "",
			wantErr: true,
		},
		{
			name:    "tool name with spaces",
			value:   "file read",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateToolName(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToolName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{
			name:    "valid relative path",
			field:   "path",
			value:   "src/main.go",
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			field:   "path",
			value:   "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "absolute path",
			field:   "path",
			value:   "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "null byte injection",
			field:   "path",
			value:   "file.txt\x00",
			wantErr: true,
		},
		{
			name:    "empty path",
			field:   "path",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateFilePath(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid command",
			value:   "ls -la",
			wantErr: false,
		},
		{
			name:    "command with semicolon",
			value:   "ls; rm -rf /",
			wantErr: true,
		},
		{
			name:    "command with pipe",
			value:   "cat file | grep text",
			wantErr: true,
		},
		{
			name:    "command with backtick",
			value:   "echo `whoami`",
			wantErr: true,
		},
		{
			name:    "empty command",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateCommand(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateChatContent(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid content",
			value:   "Hello, world!",
			wantErr: false,
		},
		{
			name:    "empty content",
			value:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
		},
		{
			name:    "null byte",
			value:   "hello\x00world",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateChatContent(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateChatContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "null bytes removed",
			input:    "hello\x00world",
			expected: "helloworld",
		},
		{
			name:     "control characters removed",
			input:    "hello\x01\x02world",
			expected: "helloworld",
		},
		{
			name:     "newlines preserved",
			input:    "hello\nworld",
			expected: "hello\nworld",
		},
		{
			name:     "tabs preserved",
			input:    "hello\tworld",
			expected: "hello\tworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "no truncation needed",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "truncate to maxLen",
			input:    "hello world",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "unicode truncation",
			input:    "hello 🌍 world",
			maxLen:   7,
			expected: "hello 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{
			name:    "valid non-empty string",
			field:   "name",
			value:   "test value",
			wantErr: false,
		},
		{
			name:    "empty string",
			field:   "name",
			value:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			field:   "name",
			value:   "   ",
			wantErr: true,
		},
		{
			name:    "string with whitespace",
			field:   "name",
			value:   "  test  ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequired(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequired() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
