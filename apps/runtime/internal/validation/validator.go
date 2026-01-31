package validation

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type Validator struct {
	maxStringLength int
	maxRequestSize  int64
}

func NewValidator() *Validator {
	return &Validator{
		maxStringLength: 100000,
		maxRequestSize:  10 * 1024 * 1024,
	}
}

func (v *Validator) WithMaxStringLength(length int) *Validator {
	v.maxStringLength = length
	return v
}

func (v *Validator) WithMaxRequestSize(size int64) *Validator {
	v.maxRequestSize = size
	return v
}

func (v *Validator) ValidateString(field, value string, opts ...StringOption) error {
	config := &stringConfig{
		minLength:  0,
		maxLength:  v.maxStringLength,
		pattern:    nil,
		allowEmpty: true,
	}

	for _, opt := range opts {
		opt(config)
	}

	length := utf8.RuneCountInString(value)
	if length < config.minLength {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be at least %d characters", config.minLength)}
	}
	if length > config.maxLength {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be at most %d characters", config.maxLength)}
	}

	if !config.allowEmpty && strings.TrimSpace(value) == "" {
		return ValidationError{Field: field, Message: "cannot be empty"}
	}

	if config.pattern != nil && !config.pattern.MatchString(value) {
		return ValidationError{Field: field, Message: "contains invalid characters"}
	}

	return nil
}

func (v *Validator) ValidateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: field, Message: "is required"}
	}
	return nil
}

func (v *Validator) ValidateID(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: field, Message: "cannot be empty"}
	}

	validID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validID.MatchString(value) {
		return ValidationError{Field: field, Message: "must contain only letters, numbers, hyphens, and underscores"}
	}

	return v.ValidateString(field, value, MaxLength(256))
}

func (v *Validator) ValidateSessionID(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	if !uuidPattern.MatchString(value) {
		return ValidationError{Field: "session_id", Message: "must be a valid UUID"}
	}

	return nil
}

func (v *Validator) ValidateToolName(value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: "tool", Message: "cannot be empty"}
	}

	validTool := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validTool.MatchString(value) {
		return ValidationError{Field: "tool", Message: "must contain only letters, numbers, dots, hyphens, and underscores"}
	}

	return v.ValidateString("tool", value, MaxLength(256))
}

func (v *Validator) ValidateFilePath(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: field, Message: "cannot be empty"}
	}

	cleanPath := filepath.Clean(value)

	if strings.Contains(cleanPath, "..") {
		return ValidationError{Field: field, Message: "path traversal not allowed"}
	}

	if strings.Contains(value, "\x00") {
		return ValidationError{Field: field, Message: "contains null bytes"}
	}

	if filepath.IsAbs(cleanPath) {
		return ValidationError{Field: field, Message: "absolute paths not allowed"}
	}

	return v.ValidateString(field, value, MaxLength(4096))
}

func (v *Validator) ValidateURL(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: field, Message: "cannot be empty"}
	}

	u, err := url.Parse(value)
	if err != nil {
		return ValidationError{Field: field, Message: "invalid URL format"}
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return ValidationError{Field: field, Message: "only HTTP and HTTPS URLs allowed"}
	}

	if isPrivateIP(u.Hostname()) {
		return ValidationError{Field: field, Message: "private IP addresses not allowed"}
	}

	return v.ValidateString(field, value, MaxLength(2048))
}

func (v *Validator) ValidateCommand(value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: "command", Message: "cannot be empty"}
	}

	dangerousPatterns := []string{
		";",
		"&&",
		"||",
		"|",
		"`",
		"$",
		"$(",
		"${",
		">",
		">>",
		"<",
		"\x00",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(value, pattern) {
			return ValidationError{Field: "command", Message: fmt.Sprintf("contains dangerous character: %s", pattern)}
		}
	}

	return v.ValidateString("command", value, MaxLength(4096))
}

func (v *Validator) ValidateChatContent(value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: "content", Message: "cannot be empty"}
	}

	if len(value) > v.maxStringLength {
		return ValidationError{Field: "content", Message: fmt.Sprintf("exceeds maximum length of %d characters", v.maxStringLength)}
	}

	if strings.Contains(value, "\x00") {
		return ValidationError{Field: "content", Message: "contains null bytes"}
	}

	return nil
}

func (v *Validator) ValidateMap(field string, m map[string]interface{}) error {
	if m == nil {
		return nil
	}

	for key, value := range m {
		if err := v.ValidateID(field+"."+key, key); err != nil {
			return err
		}

		if str, ok := value.(string); ok {
			if err := v.ValidateString(field+"."+key, str, MaxLength(v.maxStringLength)); err != nil {
				return err
			}
		}
	}

	return nil
}

type stringConfig struct {
	minLength  int
	maxLength  int
	pattern    *regexp.Regexp
	allowEmpty bool
}

type StringOption func(*stringConfig)

func MinLength(n int) StringOption {
	return func(c *stringConfig) {
		c.minLength = n
	}
}

func MaxLength(n int) StringOption {
	return func(c *stringConfig) {
		c.maxLength = n
	}
}

func Pattern(r *regexp.Regexp) StringOption {
	return func(c *stringConfig) {
		c.pattern = r
	}
}

func AllowEmpty(allow bool) StringOption {
	return func(c *stringConfig) {
		c.allowEmpty = allow
	}
}

func isPrivateIP(hostname string) bool {
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		return true
	}

	privateRanges := []string{
		"10.",
		"172.16.",
		"172.17.",
		"172.18.",
		"172.19.",
		"172.20.",
		"172.21.",
		"172.22.",
		"172.23.",
		"172.24.",
		"172.25.",
		"172.26.",
		"172.27.",
		"172.28.",
		"172.29.",
		"172.30.",
		"172.31.",
		"192.168.",
		"127.",
		"0.",
		"::1",
		"fc00:",
		"fe80:",
	}

	for _, prefix := range privateRanges {
		if strings.HasPrefix(hostname, prefix) {
			return true
		}
	}

	return false
}

func SanitizeString(value string) string {
	value = strings.ReplaceAll(value, "\x00", "")

	var result strings.Builder
	for _, r := range value {
		if r == '\n' || r == '\r' || r == '\t' || (r >= 32 && r < 127) || r > 127 {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func TruncateString(value string, maxLen int) string {
	if utf8.RuneCountInString(value) <= maxLen {
		return value
	}

	return string([]rune(value)[:maxLen])
}
