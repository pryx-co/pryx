package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Mode string

const (
	ModeFull    Mode = "full"
	ModeMinimal Mode = "minimal"
	ModeNone    Mode = "none"
)

type Builder struct {
	pryxDir string
	mode    Mode
}

func NewBuilder(pryxDir string, mode Mode) *Builder {
	if mode == "" {
		mode = ModeFull
	}
	return &Builder{
		pryxDir: pryxDir,
		mode:    mode,
	}
}

func (b *Builder) Build(metadata Metadata) (string, error) {
	switch b.mode {
	case ModeNone:
		return "", nil
	case ModeMinimal:
		return b.buildMinimal(metadata)
	case ModeFull:
		return b.buildFull(metadata)
	default:
		return b.buildFull(metadata)
	}
}

func (b *Builder) buildMinimal(metadata Metadata) (string, error) {
	parts := []string{
		"You are Pryx, a helpful AI assistant.",
		"",
		fmt.Sprintf("Current date: %s", metadata.CurrentTime.Format("2006-01-02")),
		fmt.Sprintf("Available tools: %d", len(metadata.AvailableTools)),
	}

	if len(metadata.AvailableTools) > 0 {
		parts = append(parts, "")
		parts = append(parts, "Available tools:")
		for _, tool := range metadata.AvailableTools {
			parts = append(parts, fmt.Sprintf("- %s", tool))
		}
	}

	return strings.Join(parts, "\n"), nil
}

func (b *Builder) buildFull(metadata Metadata) (string, error) {
	var parts []string

	soulContent, err := b.loadFile("SOUL.md")
	if err == nil && soulContent != "" {
		parts = append(parts, "=== PERSONA ===")
		parts = append(parts, soulContent)
		parts = append(parts, "")
	}

	agentsContent, err := b.loadFile("AGENTS.md")
	if err == nil && agentsContent != "" {
		parts = append(parts, "=== OPERATING INSTRUCTIONS ===")
		parts = append(parts, agentsContent)
		parts = append(parts, "")
	}

	parts = append(parts, "=== RUNTIME CONTEXT ===")
	parts = append(parts, fmt.Sprintf("Current date/time: %s", metadata.CurrentTime.Format(time.RFC3339)))
	parts = append(parts, fmt.Sprintf("Pryx version: %s", metadata.Version))
	parts = append(parts, fmt.Sprintf("Session ID: %s", metadata.SessionID))
	parts = append(parts, "")

	if len(metadata.AvailableTools) > 0 {
		parts = append(parts, "=== AVAILABLE TOOLS ===")
		for _, tool := range metadata.AvailableTools {
			parts = append(parts, fmt.Sprintf("- %s", tool))
		}
		parts = append(parts, "")
	}

	if len(metadata.AvailableSkills) > 0 {
		parts = append(parts, "=== AVAILABLE SKILLS ===")
		for _, skill := range metadata.AvailableSkills {
			parts = append(parts, fmt.Sprintf("- %s", skill))
		}
		parts = append(parts, "")
	}

	parts = append(parts, "=== CONSTRAINTS ===")
	parts = append(parts, getDefaultConstraints())

	return strings.Join(parts, "\n"), nil
}

func (b *Builder) loadFile(filename string) (string, error) {
	path := filepath.Join(b.pryxDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (b *Builder) SetMode(mode Mode) {
	b.mode = mode
}

func (b *Builder) GetMode() Mode {
	return b.mode
}

type Metadata struct {
	CurrentTime     time.Time
	Version         string
	SessionID       string
	AvailableTools  []string
	AvailableSkills []string
}

func DefaultPryxDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pryx")
}

func (b *Builder) EnsureTemplates() error {
	templates := map[string]string{
		"AGENTS.md": getDefaultAgentsTemplate(),
		"SOUL.md":   getDefaultSoulTemplate(),
	}

	for filename, content := range templates {
		path := filepath.Join(b.pryxDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to create %s: %w", filename, err)
			}
		}
	}

	return nil
}
