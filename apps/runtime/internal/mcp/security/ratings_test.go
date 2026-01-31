package security

import (
	"testing"
)

func TestRiskRating(t *testing.T) {
	tests := []struct {
		name     string
		rating   RiskRating
		expected string
	}{
		{"A rating", RiskRatingA, "A"},
		{"B rating", RiskRatingB, "B"},
		{"C rating", RiskRatingC, "C"},
		{"D rating", RiskRatingD, "D"},
		{"F rating", RiskRatingF, "F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.rating) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.rating)
			}
		})
	}
}

func TestRiskScoreIsSafe(t *testing.T) {
	tests := []struct {
		name     string
		score    RiskScore
		expected bool
	}{
		{"Safe A rating", RiskScore{Rating: RiskRatingA, Score: 95}, true},
		{"Safe B rating", RiskScore{Rating: RiskRatingB, Score: 80}, true},
		{"Safe C rating", RiskScore{Rating: RiskRatingC, Score: 65}, true},
		{"Safe D rating", RiskScore{Rating: RiskRatingD, Score: 50}, true},
		{"Blocked F rating", RiskScore{Rating: RiskRatingF, Score: 20}, false},
		{"Low score D", RiskScore{Rating: RiskRatingD, Score: 30}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.score.IsSafe(); got != tt.expected {
				t.Errorf("IsSafe() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRiskScoreNeedsUserConfirmation(t *testing.T) {
	tests := []struct {
		name     string
		score    RiskScore
		expected bool
	}{
		{"A rating no confirm", RiskScore{Rating: RiskRatingA, RequiresApproval: false}, false},
		{"B rating no confirm", RiskScore{Rating: RiskRatingB, RequiresApproval: false}, false},
		{"C rating no confirm", RiskScore{Rating: RiskRatingC, RequiresApproval: false}, false},
		{"D rating needs confirm", RiskScore{Rating: RiskRatingD, RequiresApproval: false}, true},
		{"F rating needs confirm", RiskScore{Rating: RiskRatingF, RequiresApproval: false}, true},
		{"Approval required", RiskScore{Rating: RiskRatingB, RequiresApproval: true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.score.NeedsUserConfirmation(); got != tt.expected {
				t.Errorf("NeedsUserConfirmation() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCalculateRiskScore(t *testing.T) {
	tests := []struct {
		name           string
		factors        RiskFactors
		expectedRating RiskRating
		minScore       int
	}{
		{
			name: "Verified HTTPS server",
			factors: RiskFactors{
				HTTPS:          true,
				VerifiedAuthor: true,
				Curated:        true,
				Transport:      "bundled",
			},
			expectedRating: RiskRatingA,
			minScore:       90,
		},
		{
			name: "Community HTTPS server",
			factors: RiskFactors{
				HTTPS:          true,
				CommunityKnown: true,
				Transport:      "http",
			},
			expectedRating: RiskRatingB,
			minScore:       75,
		},
		{
			name: "Unknown HTTP server",
			factors: RiskFactors{
				HTTPS:         false,
				DomainAllowed: true,
				Transport:     "http",
			},
			expectedRating: RiskRatingD,
			minScore:       40,
		},
		{
			name: "Blocked domain",
			factors: RiskFactors{
				HTTPS:         true,
				DomainBlocked: true,
			},
			expectedRating: RiskRatingF,
			minScore:       0,
		},
		{
			name: "Malicious server",
			factors: RiskFactors{
				HTTPS:          true,
				KnownMalicious: true,
			},
			expectedRating: RiskRatingF,
			minScore:       0,
		},
		{
			name: "Server with exec tools",
			factors: RiskFactors{
				HTTPS:         true,
				DomainAllowed: true,
				HasExecTools:  true,
			},
			expectedRating: RiskRatingB,
			minScore:       75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateRiskScore(tt.factors)
			if score.Rating != tt.expectedRating {
				t.Errorf("Expected rating %s, got %s", tt.expectedRating, score.Rating)
			}
			if score.Score < tt.minScore {
				t.Errorf("Expected score >= %d, got %d", tt.minScore, score.Score)
			}
		})
	}
}

func TestClassifyToolRisk(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		description  string
		expectedRisk ToolRiskLevel
	}{
		{
			name:         "Read tool",
			toolName:     "read_file",
			description:  "Read the contents of a file",
			expectedRisk: ToolRiskRead,
		},
		{
			name:         "Write tool",
			toolName:     "write_file",
			description:  "Write contents to a file",
			expectedRisk: ToolRiskWrite,
		},
		{
			name:         "Execute tool",
			toolName:     "execute",
			description:  "Execute a shell command",
			expectedRisk: ToolRiskExec,
		},
		{
			name:         "Network tool",
			toolName:     "fetch",
			description:  "Fetch content from a URL",
			expectedRisk: ToolRiskNetwork,
		},
		{
			name:         "System tool",
			toolName:     "chmod",
			description:  "Change file permissions",
			expectedRisk: ToolRiskSystem,
		},
		{
			name:         "Dangerous tool",
			toolName:     "delete_file",
			description:  "Delete a file permanently",
			expectedRisk: ToolRiskDangerous,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClassifyToolRisk(tt.toolName, tt.description)
			if result.Level != tt.expectedRisk {
				t.Errorf("Expected risk %s, got %s", tt.expectedRisk, result.Level)
			}
		})
	}
}

func TestAnalyzeTools(t *testing.T) {
	tools := []ToolInfo{
		{Name: "read_file", Description: "Read file"},
		{Name: "write_file", Description: "Write file"},
		{Name: "execute", Description: "Execute command"},
	}

	highRisk, allRisks := AnalyzeTools(tools)

	if len(allRisks) != 3 {
		t.Errorf("Expected 3 risk assessments, got %d", len(allRisks))
	}

	if len(highRisk) != 1 {
		t.Errorf("Expected 1 high risk, got %d", len(highRisk))
	}

	if highRisk[0].Level != ToolRiskExec {
		t.Errorf("Expected exec to be high risk, got %s", highRisk[0].Level)
	}
}
