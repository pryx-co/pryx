package security

import (
	"fmt"
	"strings"
)

// UserWarning represents a warning to display to the user before connecting
type UserWarning struct {
	Severity WarningSeverity `json:"severity"`
	Title    string          `json:"title"`
	Message  string          `json:"message"`
	Action   WarningAction   `json:"action"`
	Details  []string        `json:"details,omitempty"`
}

// WarningSeverity indicates how serious the warning is
type WarningSeverity string

const (
	SeverityInfo     WarningSeverity = "info"
	SeverityWarning  WarningSeverity = "warning"
	SeverityCritical WarningSeverity = "critical"
)

// WarningAction specifies what action the user should take
type WarningAction string

const (
	ActionNone    WarningAction = "none"    // No action needed
	ActionReview  WarningAction = "review"  // User should review details
	ActionConfirm WarningAction = "confirm" // User must explicitly confirm
	ActionDeny    WarningAction = "deny"    // Connection should be denied
)

// WarningResult contains all warnings for a server connection
type WarningResult struct {
	CanConnect      bool          `json:"can_connect"`
	RequiresConfirm bool          `json:"requires_confirm"`
	Warnings        []UserWarning `json:"warnings,omitempty"`
	RiskRating      RiskRating    `json:"risk_rating"`
}

// WarningGenerator generates user warnings based on security validation
type WarningGenerator struct {
	strictMode bool
}

// NewWarningGenerator creates a new warning generator
func NewWarningGenerator(strictMode bool) *WarningGenerator {
	return &WarningGenerator{strictMode: strictMode}
}

// GenerateWarnings creates user-facing warnings from validation results
func (g *WarningGenerator) GenerateWarnings(result ValidationResult) WarningResult {
	warningResult := WarningResult{
		CanConnect:      result.Valid,
		RequiresConfirm: result.RiskScore.NeedsUserConfirmation(),
		RiskRating:      result.RiskScore.Rating,
	}

	// Add warnings based on risk rating
	switch result.RiskScore.Rating {
	case RiskRatingF:
		warningResult.CanConnect = false
		warningResult.Warnings = append(warningResult.Warnings, UserWarning{
			Severity: SeverityCritical,
			Title:    "Connection Blocked",
			Message:  "This MCP server has been blocked due to security concerns.",
			Action:   ActionDeny,
			Details:  result.Errors,
		})

	case RiskRatingD:
		warningResult.Warnings = append(warningResult.Warnings, UserWarning{
			Severity: SeverityCritical,
			Title:    "Unverified Server",
			Message:  "This is an unverified MCP server. Use with extreme caution.",
			Action:   ActionConfirm,
			Details: append([]string{
				"The server is not in our curated list or allowlist",
				"Its security practices are unknown",
				"It may have access to sensitive data or system resources",
			}, result.Warnings...),
		})

	case RiskRatingC:
		warningResult.Warnings = append(warningResult.Warnings, UserWarning{
			Severity: SeverityWarning,
			Title:    "Limited Verification",
			Message:  "This server has limited security verification.",
			Action:   ActionReview,
			Details:  result.Warnings,
		})

	case RiskRatingB:
		if len(result.Warnings) > 0 {
			warningResult.Warnings = append(warningResult.Warnings, UserWarning{
				Severity: SeverityWarning,
				Title:    "Community Server",
				Message:  "This is a community-maintained server with good reputation.",
				Action:   ActionReview,
				Details:  result.Warnings,
			})
		}

	case RiskRatingA:
		if len(result.Warnings) > 0 {
			warningResult.Warnings = append(warningResult.Warnings, UserWarning{
				Severity: SeverityInfo,
				Title:    "Verified Server",
				Message:  "This is a verified MCP server.",
				Action:   ActionNone,
				Details:  result.Warnings,
			})
		}
	}

	// Add transport-specific warnings
	if result.Transport == "http" && !result.Security.HTTPS {
		warningResult.Warnings = append(warningResult.Warnings, UserWarning{
			Severity: SeverityCritical,
			Title:    "Insecure Connection",
			Message:  "This server uses HTTP instead of HTTPS. Data may be intercepted.",
			Action:   ActionConfirm,
			Details: []string{
				"HTTP connections are not encrypted",
				"Sensitive data may be exposed to network attackers",
				"Consider using an HTTPS-enabled server instead",
			},
		})
	}

	if result.Security.SelfSignedCert {
		warningResult.Warnings = append(warningResult.Warnings, UserWarning{
			Severity: SeverityWarning,
			Title:    "Self-Signed Certificate",
			Message:  "This server uses a self-signed SSL certificate.",
			Action:   ActionConfirm,
			Details: []string{
				"Self-signed certificates cannot be verified by standard authorities",
				"The server's identity cannot be guaranteed",
				"Only proceed if you trust this specific server",
			},
		})
	}

	// Add tool permission warnings
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "exec") || strings.Contains(warning, "command") {
			warningResult.Warnings = append(warningResult.Warnings, UserWarning{
				Severity: SeverityCritical,
				Title:    "Command Execution",
				Message:  "This server can execute arbitrary commands on your system.",
				Action:   ActionConfirm,
				Details: []string{
					"The server has access to execute shell commands",
					"This could potentially harm your system",
					"Review all commands carefully before allowing execution",
				},
			})
			warningResult.RequiresConfirm = true
		}
	}

	// In strict mode, lower-rated servers require confirmation
	if g.strictMode && result.RiskScore.Rating <= RiskRatingC {
		warningResult.RequiresConfirm = true
	}

	return warningResult
}

// FormatWarning formats a warning for display
func FormatWarning(warning UserWarning) string {
	var severityIcon string
	switch warning.Severity {
	case SeverityInfo:
		severityIcon = "ℹ️"
	case SeverityWarning:
		severityIcon = "⚠️"
	case SeverityCritical:
		severityIcon = "🚫"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s\n", severityIcon, warning.Title))
	b.WriteString(fmt.Sprintf("   %s\n", warning.Message))

	if len(warning.Details) > 0 {
		b.WriteString("\n   Details:\n")
		for _, detail := range warning.Details {
			b.WriteString(fmt.Sprintf("   • %s\n", detail))
		}
	}

	b.WriteString(fmt.Sprintf("\n   Action Required: %s\n", warning.Action))

	return b.String()
}

// FormatWarnings formats all warnings for display
func FormatWarnings(result WarningResult) string {
	if len(result.Warnings) == 0 {
		return "✓ No security warnings"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Security Assessment: %s\n", result.RiskRating))
	b.WriteString(strings.Repeat("=", 50) + "\n\n")

	for i, warning := range result.Warnings {
		b.WriteString(fmt.Sprintf("[%d/%d] ", i+1, len(result.Warnings)))
		b.WriteString(FormatWarning(warning))
		b.WriteString("\n")
	}

	if result.RequiresConfirm {
		b.WriteString("⚠️  You must explicitly confirm to proceed with this connection.\n")
	}

	return b.String()
}

// ShouldBlockConnection returns true if connection should be blocked
func ShouldBlockConnection(result WarningResult) bool {
	if !result.CanConnect {
		return true
	}

	for _, warning := range result.Warnings {
		if warning.Action == ActionDeny {
			return true
		}
	}

	return false
}

// ConfirmationDialog represents a user confirmation dialog
type ConfirmationDialog struct {
	Title      string   `json:"title"`
	Message    string   `json:"message"`
	RiskRating string   `json:"risk_rating"`
	Warnings   []string `json:"warnings"`
	Options    []string `json:"options"`
	Default    int      `json:"default"`
}

// BuildConfirmationDialog creates a confirmation dialog for risky connections
func BuildConfirmationDialog(result WarningResult) ConfirmationDialog {
	dialog := ConfirmationDialog{
		Title:      "Security Warning",
		RiskRating: string(result.RiskRating),
		Options:    []string{"Cancel", "Allow Once", "Allow Always"},
		Default:    0,
	}

	switch result.RiskRating {
	case RiskRatingF:
		dialog.Title = "Connection Blocked"
		dialog.Message = "This MCP server has been blocked for security reasons and cannot be connected."
		dialog.Options = []string{"Cancel"}

	case RiskRatingD:
		dialog.Title = "Unverified Server Warning"
		dialog.Message = "You are about to connect to an unverified MCP server. This server could potentially access your files, execute commands, or perform other sensitive operations."

	case RiskRatingC:
		dialog.Title = "Limited Verification Warning"
		dialog.Message = "This MCP server has limited security verification. Please review the warnings below before proceeding."

	default:
		dialog.Title = "Security Notice"
		dialog.Message = "Please review the following security information before connecting."
	}

	for _, warning := range result.Warnings {
		dialog.Warnings = append(dialog.Warnings, warning.Message)
		if len(warning.Details) > 0 {
			dialog.Warnings = append(dialog.Warnings, warning.Details...)
		}
	}

	return dialog
}

// AuditLogEntry represents a security audit log entry
type AuditLogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Event       string                 `json:"event"`
	ServerURL   string                 `json:"server_url"`
	RiskRating  string                 `json:"risk_rating"`
	Action      string                 `json:"action"`
	UserConfirm bool                   `json:"user_confirm"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityAuditLogger handles security audit logging
type SecurityAuditLogger struct {
	enabled bool
}

// NewSecurityAuditLogger creates a new audit logger
func NewSecurityAuditLogger(enabled bool) *SecurityAuditLogger {
	return &SecurityAuditLogger{enabled: enabled}
}

// LogConnection logs an MCP connection attempt
func (l *SecurityAuditLogger) LogConnection(serverURL string, result ValidationResult, userConfirmed bool) AuditLogEntry {
	action := "allowed"
	if !result.Valid {
		action = "blocked"
	} else if result.RiskScore.NeedsUserConfirmation() {
		action = "allowed_with_confirmation"
	}

	entry := AuditLogEntry{
		Timestamp:   "", // Set by caller with actual time
		Event:       "mcp_connection",
		ServerURL:   serverURL,
		RiskRating:  string(result.RiskScore.Rating),
		Action:      action,
		UserConfirm: userConfirmed,
		Metadata: map[string]interface{}{
			"transport":      result.Transport,
			"https":          result.Security.HTTPS,
			"domain_allowed": result.Security.DomainAllowed,
			"warnings":       len(result.Warnings),
		},
	}

	return entry
}

// LogToolExecution logs a tool execution event
func (l *SecurityAuditLogger) LogToolExecution(serverURL, toolName string, risk ToolRiskLevel) AuditLogEntry {
	return AuditLogEntry{
		Timestamp: "",
		Event:     "tool_execution",
		ServerURL: serverURL,
		Action:    "executed",
		Metadata: map[string]interface{}{
			"tool":       toolName,
			"risk_level": string(risk),
		},
	}
}
