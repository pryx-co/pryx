package security

// SecurityService provides a unified interface for MCP server security
type SecurityService struct {
	validator *Validator
	generator *WarningGenerator
	auditor   *SecurityAuditLogger
}

// NewSecurityService creates a new security service with default settings
func NewSecurityService() *SecurityService {
	return &SecurityService{
		validator: NewValidator(),
		generator: NewWarningGenerator(false),
		auditor:   NewSecurityAuditLogger(true),
	}
}

// NewSecurityServiceWithOptions creates a security service with custom options
func NewSecurityServiceWithOptions(
	validatorOpts []ValidatorOption,
	strictMode bool,
	enableAudit bool,
) *SecurityService {
	return &SecurityService{
		validator: NewValidator(validatorOpts...),
		generator: NewWarningGenerator(strictMode),
		auditor:   NewSecurityAuditLogger(enableAudit),
	}
}

// ValidateServer performs security validation on an MCP server
func (s *SecurityService) ValidateServer(url, transport string, tools []ToolInfo) (ValidationResult, WarningResult) {
	validation := s.validator.ValidateWithTools(url, transport, tools)
	warnings := s.generator.GenerateWarnings(validation)
	return validation, warnings
}

// ValidateServerConfig validates a complete MCP server configuration
func (s *SecurityService) ValidateServerConfig(config MCPConfig) (ValidationResult, WarningResult) {
	validation := s.validator.ValidateMCPConfig(config)
	warnings := s.generator.GenerateWarnings(validation)
	return validation, warnings
}

// CheckURL performs basic URL validation
func (s *SecurityService) CheckURL(url, transport string) ValidationResult {
	return s.validator.Validate(url, transport)
}

// GetSandboxForRisk returns an appropriate sandbox for a risk rating
func (s *SecurityService) GetSandboxForRisk(risk RiskScore) *Sandbox {
	return SandboxFromRisk(risk)
}

// ShouldAllowConnection determines if a connection should be allowed
func (s *SecurityService) ShouldAllowConnection(warnings WarningResult) bool {
	return !ShouldBlockConnection(warnings)
}

// GetSecurityStatus returns a summary of security status
func (s *SecurityService) GetSecurityStatus(result ValidationResult) map[string]interface{} {
	return map[string]interface{}{
		"valid":             result.Valid,
		"risk_rating":       result.RiskScore.Rating,
		"risk_score":        result.RiskScore.Score,
		"requires_approval": result.RiskScore.NeedsUserConfirmation(),
		"https":             result.Security.HTTPS,
		"domain_allowed":    result.Security.DomainAllowed,
		"warnings_count":    len(result.Warnings),
	}
}

// Export exports the security package types for use by other packages
var (
	// Risk ratings
	RatingA = RiskRatingA
	RatingB = RiskRatingB
	RatingC = RiskRatingC
	RatingD = RiskRatingD
	RatingF = RiskRatingF

	// Tool risk levels
	RiskRead      = ToolRiskRead
	RiskWrite     = ToolRiskWrite
	RiskExec      = ToolRiskExec
	RiskNetwork   = ToolRiskNetwork
	RiskSystem    = ToolRiskSystem
	RiskDangerous = ToolRiskDangerous
)
