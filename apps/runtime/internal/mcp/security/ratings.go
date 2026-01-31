// Package security provides MCP server security validation, risk classification,
// and sandboxing capabilities to prevent malicious code execution.
package security

import (
	"fmt"
	"strings"
)

// RiskRating represents the security risk level of an MCP server (A-F scale)
type RiskRating string

const (
	// RiskRatingA - Verified/official servers with strong security practices
	RiskRatingA RiskRating = "A"
	// RiskRatingB - Community servers with good reputation
	RiskRatingB RiskRating = "B"
	// RiskRatingC - Known but unverified servers
	RiskRatingC RiskRating = "C"
	// RiskRatingD - Unknown servers (warn user)
	RiskRatingD RiskRating = "D"
	// RiskRatingF - Blocked servers (malicious patterns detected)
	RiskRatingF RiskRating = "F"
)

// RiskScore represents a detailed security assessment
type RiskScore struct {
	Rating           RiskRating `json:"rating"`
	Score            int        `json:"score"` // 0-100, higher is better
	Category         string     `json:"category"`
	Verified         bool       `json:"verified"`
	Warnings         []string   `json:"warnings,omitempty"`
	RequiresApproval bool       `json:"requires_approval"`
}

// IsSafe returns true if the server is considered safe to connect
func (r RiskScore) IsSafe() bool {
	return r.Rating != RiskRatingF && r.Score >= 40
}

// NeedsUserConfirmation returns true if user must explicitly approve
func (r RiskScore) NeedsUserConfirmation() bool {
	return r.RequiresApproval || r.Rating == RiskRatingD || r.Rating == RiskRatingF
}

// String returns a human-readable representation
func (r RiskScore) String() string {
	status := "✓ Safe"
	if r.Rating == RiskRatingF {
		status = "✗ Blocked"
	} else if r.Rating == RiskRatingD {
		status = "⚠ Warning"
	} else if r.RequiresApproval {
		status = "⚠ Approval Required"
	}
	return fmt.Sprintf("[%s] %s (Score: %d/100) - %s", r.Rating, status, r.Score, r.Category)
}

// CalculateRiskScore computes a risk score based on various factors
func CalculateRiskScore(factors RiskFactors) RiskScore {
	score := 100
	var warnings []string
	requiresApproval := false

	// URL security deductions
	if !factors.HTTPS {
		score -= 30
		warnings = append(warnings, "Connection uses HTTP (insecure)")
		requiresApproval = true
	}

	if factors.SelfSignedCert {
		score -= 25
		warnings = append(warnings, "Self-signed certificate detected")
		requiresApproval = true
	}

	if factors.DomainBlocked {
		score = 0
		warnings = append(warnings, "Domain is in blocklist")
		return RiskScore{
			Rating:           RiskRatingF,
			Score:            0,
			Category:         "blocked",
			Verified:         false,
			Warnings:         warnings,
			RequiresApproval: true,
		}
	}

	if !factors.DomainAllowed {
		score -= 15
		warnings = append(warnings, "Domain not in allowlist")
	}

	// Transport security
	switch factors.Transport {
	case "bundled":
		// Bundled servers get bonus points
		score += 10
	case "stdio":
		// Local process, generally safe but depends on the command
		score -= 5
	case "http", "sse":
		// Network-based, requires HTTPS
		if !factors.HTTPS {
			score -= 20
		}
	}

	// Tool risk analysis
	if factors.HasWriteTools {
		score -= 10
		warnings = append(warnings, "Server has filesystem write capabilities")
		requiresApproval = true
	}

	if factors.HasExecTools {
		score -= 20
		warnings = append(warnings, "Server can execute arbitrary commands")
		requiresApproval = true
	}

	if factors.HasNetworkTools {
		score -= 5
		warnings = append(warnings, "Server can make network requests")
	}

	// Reputation factors
	if factors.KnownMalicious {
		score = 0
		warnings = append(warnings, "Server matches known malicious patterns")
		return RiskScore{
			Rating:           RiskRatingF,
			Score:            0,
			Category:         "malicious",
			Verified:         false,
			Warnings:         warnings,
			RequiresApproval: true,
		}
	}

	if factors.VerifiedAuthor {
		score += 15
	}

	if factors.Curated {
		score += 10
	}

	// Determine rating based on score
	var rating RiskRating
	switch {
	case score >= 90:
		rating = RiskRatingA
	case score >= 75:
		rating = RiskRatingB
	case score >= 60:
		rating = RiskRatingC
	case score >= 40:
		rating = RiskRatingD
	default:
		rating = RiskRatingF
		requiresApproval = true
	}

	category := "unknown"
	if factors.Curated {
		category = "curated"
	} else if factors.VerifiedAuthor {
		category = "verified"
	} else if factors.CommunityKnown {
		category = "community"
	}

	return RiskScore{
		Rating:           rating,
		Score:            max(0, min(100, score)),
		Category:         category,
		Verified:         factors.VerifiedAuthor || factors.Curated,
		Warnings:         warnings,
		RequiresApproval: requiresApproval || rating == RiskRatingD,
	}
}

// RiskFactors contains all the factors considered in risk scoring
type RiskFactors struct {
	// URL/Transport factors
	HTTPS          bool
	SelfSignedCert bool
	DomainAllowed  bool
	DomainBlocked  bool
	Transport      string

	// Tool capabilities
	HasWriteTools   bool
	HasReadTools    bool
	HasExecTools    bool
	HasNetworkTools bool

	// Reputation
	VerifiedAuthor bool
	Curated        bool
	CommunityKnown bool
	KnownMalicious bool
}

// ToolRiskLevel represents the risk level of a specific tool
type ToolRiskLevel string

const (
	ToolRiskRead      ToolRiskLevel = "read"      // Safe read-only operations
	ToolRiskWrite     ToolRiskLevel = "write"     // File/disk write operations
	ToolRiskExec      ToolRiskLevel = "exec"      // Command execution
	ToolRiskNetwork   ToolRiskLevel = "network"   // Network access
	ToolRiskSystem    ToolRiskLevel = "system"    // System-level operations
	ToolRiskDangerous ToolRiskLevel = "dangerous" // Potentially destructive
)

// ToolRiskAssessment contains risk information for a tool
type ToolRiskAssessment struct {
	Name        string        `json:"name"`
	Level       ToolRiskLevel `json:"level"`
	Description string        `json:"description"`
	Warnings    []string      `json:"warnings,omitempty"`
}

// ClassifyToolRisk classifies a tool name/description into risk levels
func ClassifyToolRisk(name, description string) ToolRiskAssessment {
	nameLower := strings.ToLower(name)
	descLower := strings.ToLower(description)
	combined := nameLower + " " + descLower

	assessment := ToolRiskAssessment{
		Name:        name,
		Description: description,
		Level:       ToolRiskRead, // Default to read
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"delete", "remove", "destroy", "wipe", "erase",
		"format", "drop", "truncate", "purge",
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(combined, pattern) {
			assessment.Level = ToolRiskDangerous
			assessment.Warnings = append(assessment.Warnings, "Tool may delete or destroy data")
			return assessment
		}
	}

	// Check for execution patterns
	execPatterns := []string{
		"exec", "execute", "run", "shell", "command", "cmd",
		"spawn", "fork", "system", "eval", "script",
	}
	for _, pattern := range execPatterns {
		if strings.Contains(combined, pattern) {
			assessment.Level = ToolRiskExec
			assessment.Warnings = append(assessment.Warnings, "Tool executes arbitrary commands")
			return assessment
		}
	}

	// Check for write patterns
	writePatterns := []string{
		"write", "create", "update", "modify", "edit", "save",
		"append", "touch", "mkdir", "rmdir", "rename", "move",
	}
	for _, pattern := range writePatterns {
		if strings.Contains(combined, pattern) {
			assessment.Level = ToolRiskWrite
			assessment.Warnings = append(assessment.Warnings, "Tool modifies files or data")
			return assessment
		}
	}

	// Check for system patterns
	systemPatterns := []string{
		"install", "uninstall", "configure", "setup",
		"permission", "chmod", "chown", "sudo",
	}
	for _, pattern := range systemPatterns {
		if strings.Contains(combined, pattern) {
			assessment.Level = ToolRiskSystem
			assessment.Warnings = append(assessment.Warnings, "Tool performs system-level operations")
			return assessment
		}
	}

	// Check for network patterns
	networkPatterns := []string{
		"fetch", "http", "request", "download", "upload",
		"connect", "socket", "api", "webhook", "curl", "wget",
	}
	for _, pattern := range networkPatterns {
		if strings.Contains(combined, pattern) {
			assessment.Level = ToolRiskNetwork
			assessment.Warnings = append(assessment.Warnings, "Tool makes network connections")
			return assessment
		}
	}

	return assessment
}

// AnalyzeTools analyzes a list of tools and returns risk summary
func AnalyzeTools(tools []ToolInfo) (highRisk []ToolRiskAssessment, allRisks []ToolRiskAssessment) {
	for _, tool := range tools {
		assessment := ClassifyToolRisk(tool.Name, tool.Description)
		allRisks = append(allRisks, assessment)

		switch assessment.Level {
		case ToolRiskDangerous, ToolRiskExec, ToolRiskSystem:
			highRisk = append(highRisk, assessment)
		}
	}
	return highRisk, allRisks
}

// ToolInfo represents a tool for risk analysis
type ToolInfo struct {
	Name        string
	Description string
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
