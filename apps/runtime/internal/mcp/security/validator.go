package security

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ValidationResult contains the complete validation results for an MCP server URL
type ValidationResult struct {
	Valid         bool          `json:"valid"`
	URL           string        `json:"url"`
	NormalizedURL string        `json:"normalized_url"`
	Transport     string        `json:"transport"`
	Security      SecurityCheck `json:"security"`
	RiskScore     RiskScore     `json:"risk_score"`
	Errors        []string      `json:"errors,omitempty"`
	Warnings      []string      `json:"warnings,omitempty"`
}

// SecurityCheck contains security-specific validation results
type SecurityCheck struct {
	HTTPSRequired    bool `json:"https_required"`
	HTTPS            bool `json:"https"`
	DomainAllowed    bool `json:"domain_allowed"`
	DomainBlocked    bool `json:"domain_blocked"`
	SelfSignedCert   bool `json:"self_signed_cert"`
	LocalhostAllowed bool `json:"localhost_allowed"`
}

// Validator handles MCP server URL and security validation
type Validator struct {
	allowlist       []string
	blocklist       []string
	allowLocalhost  bool
	requireHTTPS    bool
	allowSelfSigned bool
}

// ValidatorOption configures the validator
type ValidatorOption func(*Validator)

// WithAllowlist sets the domain allowlist
func WithAllowlist(allowlist []string) ValidatorOption {
	return func(v *Validator) {
		v.allowlist = allowlist
	}
}

// WithBlocklist sets the domain blocklist
func WithBlocklist(blocklist []string) ValidatorOption {
	return func(v *Validator) {
		v.blocklist = blocklist
	}
}

// WithLocalhost allows localhost connections (for development)
func WithLocalhost(allow bool) ValidatorOption {
	return func(v *Validator) {
		v.allowLocalhost = allow
	}
}

// WithHTTPSRequirement enforces HTTPS for remote servers
func WithHTTPSRequirement(require bool) ValidatorOption {
	return func(v *Validator) {
		v.requireHTTPS = require
	}
}

// WithSelfSignedCerts allows self-signed certificates (insecure, dev only)
func WithSelfSignedCerts(allow bool) ValidatorOption {
	return func(v *Validator) {
		v.allowSelfSigned = allow
	}
}

// NewValidator creates a new MCP server security validator
func NewValidator(opts ...ValidatorOption) *Validator {
	v := &Validator{
		allowlist:       DefaultAllowlist(),
		blocklist:       DefaultBlocklist(),
		allowLocalhost:  false,
		requireHTTPS:    true,
		allowSelfSigned: false,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// Validate validates an MCP server URL and returns a complete assessment
func (v *Validator) Validate(urlStr string, transport string) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		URL:       urlStr,
		Transport: transport,
	}

	// Normalize URL
	normalized, err := v.normalizeURL(urlStr)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("URL normalization failed: %v", err))
		return result
	}
	result.NormalizedURL = normalized

	// Parse and validate URL
	parsed, err := url.Parse(normalized)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("URL parsing failed: %v", err))
		return result
	}

	// Security checks
	security := v.performSecurityChecks(parsed, transport)
	result.Security = security

	if !security.DomainAllowed {
		result.Warnings = append(result.Warnings, "Domain not in allowlist")
	}

	if security.DomainBlocked {
		result.Valid = false
		result.Errors = append(result.Errors, "Domain is blocked")
	}

	if !security.LocalhostAllowed && v.isLocalhost(parsed.Hostname()) && transport != "bundled" {
		result.Valid = false
		result.Errors = append(result.Errors, "Localhost connections not allowed (use allowLocalhost option for development)")
	}

	if v.requireHTTPS && !security.HTTPS && transport != "bundled" && transport != "stdio" {
		if transport == "http" || transport == "sse" {
			result.Warnings = append(result.Warnings, "HTTPS is required for remote servers for security")
		}
	}

	// Calculate risk score
	riskFactors := RiskFactors{
		HTTPS:          security.HTTPS,
		SelfSignedCert: security.SelfSignedCert,
		DomainAllowed:  security.DomainAllowed,
		DomainBlocked:  security.DomainBlocked,
		Transport:      transport,
	}
	result.RiskScore = CalculateRiskScore(riskFactors)

	if result.RiskScore.Rating == RiskRatingF {
		result.Valid = false
	}

	result.Warnings = append(result.Warnings, result.RiskScore.Warnings...)

	return result
}

// ValidateWithTools performs full validation including tool risk analysis
func (v *Validator) ValidateWithTools(urlStr string, transport string, tools []ToolInfo) ValidationResult {
	result := v.Validate(urlStr, transport)

	// Analyze tools
	highRisk, allRisks := AnalyzeTools(tools)

	// Update risk factors with tool information
	for _, risk := range allRisks {
		switch risk.Level {
		case ToolRiskWrite:
			result.RiskScore.RequiresApproval = true
		case ToolRiskExec:
			result.RiskScore.RequiresApproval = true
		case ToolRiskDangerous:
			result.RiskScore.RequiresApproval = true
			result.RiskScore.Rating = RiskRatingF
			result.Valid = false
		}
	}

	// Add warnings for high-risk tools
	for _, risk := range highRisk {
		result.Warnings = append(result.Warnings, fmt.Sprintf("High-risk tool '%s': %s", risk.Name, risk.Description))
	}

	// Add warnings for write tools (medium risk)
	for _, risk := range allRisks {
		if risk.Level == ToolRiskWrite {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Write tool '%s': %s", risk.Name, risk.Description))
		}
	}

	return result
}

// CheckDomain checks if a domain is allowed/blocked
func (v *Validator) CheckDomain(domain string) (allowed, blocked bool) {
	domain = strings.ToLower(domain)

	// Check blocklist first
	for _, blockedPattern := range v.blocklist {
		if matchesDomainPattern(domain, blockedPattern) {
			return false, true
		}
	}

	// Check allowlist (if not empty, domain must match)
	if len(v.allowlist) > 0 {
		for _, allowedPattern := range v.allowlist {
			if matchesDomainPattern(domain, allowedPattern) {
				return true, false
			}
		}
		return false, false
	}

	return true, false
}

// IsLocalhost checks if a hostname is localhost
func (v *Validator) isLocalhost(hostname string) bool {
	hostname = strings.ToLower(hostname)
	localhostPatterns := []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"::1",
		"[::]",
		"[::1]",
	}
	for _, pattern := range localhostPatterns {
		if hostname == pattern {
			return true
		}
	}
	return false
}

func (v *Validator) normalizeURL(urlStr string) (string, error) {
	urlStr = strings.TrimSpace(urlStr)

	// Handle bundled transport
	if urlStr == "bundled" || urlStr == "" {
		return "bundled://localhost", nil
	}

	// Ensure scheme is present
	if !strings.Contains(urlStr, "://") {
		urlStr = "https://" + urlStr
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Normalize
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimSuffix(parsed.Path, "/")

	return parsed.String(), nil
}

func (v *Validator) performSecurityChecks(parsed *url.URL, transport string) SecurityCheck {
	isBundled := transport == "bundled" || parsed.Scheme == "bundled"
	isStdio := transport == "stdio" || parsed.Scheme == "stdio"

	security := SecurityCheck{
		HTTPSRequired:    v.requireHTTPS && !isBundled && !isStdio,
		HTTPS:            parsed.Scheme == "https",
		LocalhostAllowed: v.allowLocalhost,
	}

	// Check domain allowlist/blocklist
	if !isBundled {
		security.DomainAllowed, security.DomainBlocked = v.CheckDomain(parsed.Hostname())
	} else {
		security.DomainAllowed = true
		security.DomainBlocked = false
	}

	// Check for self-signed certs (would require actual TLS connection)
	security.SelfSignedCert = false

	return security
}

// matchesDomainPattern checks if a domain matches a pattern
// Supports: exact match, substring, wildcard (*.example.com), and regex (regex:pattern)
func matchesDomainPattern(domain, pattern string) bool {
	domain = strings.ToLower(domain)
	pattern = strings.ToLower(pattern)

	// Exact match
	if domain == pattern {
		return true
	}

	// Substring match (for blocklist patterns like "malware")
	if strings.Contains(domain, pattern) {
		return true
	}

	// Wildcard match (*.example.com)
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[2:]
		return strings.HasSuffix(domain, suffix) && domain != suffix
	}

	// Regex match (regex:pattern)
	if strings.HasPrefix(pattern, "regex:") {
		regex := pattern[6:]
		re, err := regexp.Compile(regex)
		if err != nil {
			return false
		}
		return re.MatchString(domain)
	}

	return false
}

// DefaultAllowlist returns the default domain allowlist (empty = allow all)
func DefaultAllowlist() []string {
	return []string{
		// Well-known MCP server domains
		"*.github.com",
		"*.modelcontextprotocol.io",
		"*.anthropic.com",
		"*.openai.com",
		// Add more trusted domains as needed
	}
}

// DefaultBlocklist returns the default malicious/suspicious domain blocklist
func DefaultBlocklist() []string {
	return []string{
		// Common malicious patterns
		"malware",
		"phishing",
		"suspicious",
		// Add specific blocked domains here
	}
}

// CheckCertificate performs TLS certificate validation
func CheckCertificate(serverURL string) (valid bool, selfSigned bool, err error) {
	parsed, err := url.Parse(serverURL)
	if err != nil {
		return false, false, err
	}

	if parsed.Scheme != "https" {
		return true, false, nil // No cert check needed for non-HTTPS
	}

	// Create TLS config with proper verification
	conf := &tls.Config{
		InsecureSkipVerify: false,
	}

	conn, err := tls.Dial("tcp", parsed.Host+":443", conf)
	if err != nil {
		// Check if it's a self-signed cert error
		if strings.Contains(err.Error(), "certificate signed by unknown authority") {
			return false, true, nil
		}
		return false, false, err
	}
	defer conn.Close()

	// Verify certificate chain
	state := conn.ConnectionState()
	for _, cert := range state.PeerCertificates {
		// Check if self-signed (issuer == subject for root cert)
		if len(state.PeerCertificates) == 1 && cert.Issuer.String() == cert.Subject.String() {
			return true, true, nil
		}
	}

	return true, false, nil
}

func (v *Validator) ValidateMCPConfig(config MCPConfig) ValidationResult {
	transport := config.Transport
	if transport == "" {
		transport = "stdio"
	}

	urlStr := config.URL
	if urlStr == "" && len(config.Command) > 0 {
		urlStr = "stdio://" + config.Command[0]
	}

	return v.ValidateWithTools(urlStr, transport, config.Tools)
}

// MCPConfig represents an MCP server configuration to validate
type MCPConfig struct {
	Name      string
	Transport string
	URL       string
	Command   []string
	Tools     []ToolInfo
}
