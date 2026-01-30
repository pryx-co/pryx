package prompt

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestBuilder_NewBuilder(t *testing.T) {
	pryxDir := t.TempDir()
	defer func() {
		os.RemoveAll(pryxDir)
	}()

	builder := NewBuilder(pryxDir, ModeFull)

	if builder == nil {
		t.Fatal("Failed to create builder")
	}

	if builder.pryxDir != pryxDir {
		t.Errorf("Expected pryxDir to be %s, got %s", pryxDir, builder.pryxDir)
	}
}

func TestBuilder_NewBuilder_DefaultMode(t *testing.T) {
	pryxDir := t.TempDir()

	builder := NewBuilder(pryxDir, ModeFull)

	if builder.mode != ModeFull {
		t.Errorf("Expected default mode to be full, got %v", builder.mode)
	}
}

func TestBuilder_NewBuilder_CustomMode(t *testing.T) {
	pryxDir := t.TempDir()

	builder := NewBuilder(pryxDir, ModeMinimal)

	if builder.mode != ModeMinimal {
		t.Errorf("Expected mode to be minimal, got %v", builder.mode)
	}
}

func TestBuilder_SetMode(t *testing.T) {
	pryxDir := t.TempDir()

	builder := NewBuilder(pryxDir, ModeFull)
	builder.SetMode(ModeMinimal)

	if builder.GetMode() != ModeMinimal {
		t.Errorf("Expected mode to be minimal, got %v", builder.GetMode())
	}
}

func TestBuilder_BuildMinimal(t *testing.T) {
	pryxDir := t.TempDir()

	builder := NewBuilder(pryxDir, ModeMinimal)

	metadata := Metadata{
		AvailableTools:  []string{"tool1", "tool2"},
		AvailableSkills: []string{"skill1", "skill2"},
		CurrentTime:     time.Now(),
		Version:         "test-1.0",
		SessionID:       "test-session-123",
		Confidence:      ConfidenceHigh,
	}

	result, err := builder.Build(metadata)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	resultStr := result
	if !strings.Contains(resultStr, "You are Pryx") {
		t.Errorf("Expected 'You are Pryx' in minimal output, got: %s", resultStr)
	}

	if !strings.Contains(resultStr, "Available tools:") {
		t.Errorf("Expected 'Available tools:' in output, got: %s", resultStr)
	}
}

func TestBuilder_BuildFull(t *testing.T) {
	pryxDir := t.TempDir()

	builder := NewBuilder(pryxDir, ModeFull)

	metadata := Metadata{
		AvailableTools:  []string{"tool1", "tool2"},
		AvailableSkills: []string{"skill1", "skill2"},
		CurrentTime:     time.Now(),
		Version:         "test-2.0",
		SessionID:       "test-session-456",
		Confidence:      ConfidenceMedium,
	}

	result, err := builder.Build(metadata)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	resultStr := result

	expectedSections := []string{
		"=== RUNTIME CONTEXT ===",
		"=== AVAILABLE TOOLS ===",
		"=== AVAILABLE SKILLS ===",
		"=== CONSTRAINTS ===",
	}

	for _, section := range expectedSections {
		if !strings.Contains(resultStr, section) {
			t.Errorf("Expected section '%s' in full output, got: %s", section, resultStr)
		}
	}
}

func TestBuilder_BuildFull_WithTools(t *testing.T) {
	pryxDir := t.TempDir()

	builder := NewBuilder(pryxDir, ModeFull)

	metadata := Metadata{
		AvailableTools:  []string{"search", "fetch", "analyze"},
		AvailableSkills: []string{"code", "debug"},
		CurrentTime:     time.Now(),
		Version:         "test-3.0",
		SessionID:       "test-session-789",
		Confidence:      ConfidenceHigh,
	}

	result, err := builder.Build(metadata)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	resultStr := result

	if !strings.Contains(resultStr, "search") || !strings.Contains(resultStr, "fetch") || !strings.Contains(resultStr, "analyze") {
		t.Errorf("Expected tool names in output, got: %s", resultStr)
	}
}

func TestBuilder_Build_NoAvailableTools(t *testing.T) {
	pryxDir := t.TempDir()

	builder := NewBuilder(pryxDir, ModeFull)

	metadata := Metadata{
		AvailableTools:  []string{},
		AvailableSkills: []string{},
		CurrentTime:     time.Now(),
		Version:         "test-4.0",
		SessionID:       "test-session-111",
		Confidence:      ConfidenceMedium,
	}

	result, err := builder.Build(metadata)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	resultStr := result

	if strings.Contains(resultStr, "AVAILABLE TOOLS ===") {
		t.Errorf("Should not show tools section when no tools available, got: %s", resultStr)
	}
}

func TestBuilder_Build_ConfidenceHigh(t *testing.T) {
	pryxDir := t.TempDir()

	builder := NewBuilder(pryxDir, ModeFull)

	metadata := Metadata{
		AvailableTools:  []string{"tool1"},
		AvailableSkills: []string{},
		CurrentTime:     time.Now(),
		Version:         "test-5.0",
		SessionID:       "test-session-222",
		Confidence:      ConfidenceHigh,
	}

	result, err := builder.Build(metadata)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	resultStr := result

	if !strings.Contains(resultStr, "GUIDANCE: You have HIGH confidence") {
		t.Errorf("Expected HIGH confidence message, got: %s", resultStr)
	}
}

func TestBuilder_Build_ConfidenceMedium(t *testing.T) {
	pryxDir := t.TempDir()

	builder := NewBuilder(pryxDir, ModeFull)

	metadata := Metadata{
		AvailableTools:  []string{"tool1"},
		AvailableSkills: []string{},
		CurrentTime:     time.Now(),
		Version:         "test-6.0",
		SessionID:       "test-session-333",
		Confidence:      ConfidenceMedium,
	}

	result, err := builder.Build(metadata)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	resultStr := result

	if !strings.Contains(resultStr, "GUIDANCE: You have MEDIUM confidence") {
		t.Errorf("Expected MEDIUM confidence message, got: %s", resultStr)
	}
}

func TestBuilder_Build_ConfidenceLow(t *testing.T) {
	pryxDir := t.TempDir()

	builder := NewBuilder(pryxDir, ModeFull)

	metadata := Metadata{
		AvailableTools:  []string{"tool1"},
		AvailableSkills: []string{},
		CurrentTime:     time.Now(),
		Version:         "test-7.0",
		SessionID:       "test-session-444",
		Confidence:      ConfidenceLow,
	}

	result, err := builder.Build(metadata)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	resultStr := result

	if !strings.Contains(resultStr, "GUIDANCE: You have LOW confidence") {
		t.Errorf("Expected LOW confidence message, got: %s", resultStr)
	}
}
