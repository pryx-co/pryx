package skills

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type RemoteInstallResult struct {
	Skill        Skill
	DownloadedAt time.Time
	SourceURL    string
}

func InstallFromURL(ctx context.Context, url string, opts Options) (*RemoteInstallResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download skill: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	fm, body, err := parseSkillFile(data)
	if err != nil {
		return nil, fmt.Errorf("parse skill: %w", err)
	}

	skillID := fm.Name
	if skillID == "" {
		return nil, fmt.Errorf("skill missing required 'name' field")
	}

	skillDir := filepath.Join(opts.ManagedRoot, skillID)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return nil, fmt.Errorf("create skill dir: %w", err)
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, data, 0644); err != nil {
		return nil, fmt.Errorf("save skill: %w", err)
	}

	skill := Skill{
		ID:          skillID,
		Source:      SourceRemote,
		Name:        fm.Name,
		Description: fm.Description,
		Path:        skillPath,
		Frontmatter: fm,
		Enabled:     true,
		Eligible:    true,
		bodyLoader: func() (string, error) {
			return body, nil
		},
	}

	return &RemoteInstallResult{
		Skill:        skill,
		DownloadedAt: time.Now(),
		SourceURL:    url,
	}, nil
}

func UninstallSkill(skillID string, opts Options) error {
	skillDir := filepath.Join(opts.ManagedRoot, skillID)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill not found: %s", skillID)
	}
	return os.RemoveAll(skillDir)
}

func ListInstalled(opts Options) ([]Skill, error) {
	entries, err := os.ReadDir(opts.ManagedRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return []Skill{}, nil
		}
		return nil, err
	}

	var skills []Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(opts.ManagedRoot, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(skillPath)
		if err != nil {
			continue
		}

		fm, _, err := parseSkillFile(data)
		if err != nil {
			continue
		}

		info, _ := entry.Info()
		skills = append(skills, Skill{
			ID:          fm.Name,
			Source:      SourceRemote,
			Name:        fm.Name,
			Description: fm.Description,
			Path:        skillPath,
			Frontmatter: fm,
			Enabled:     true,
			Eligible:    true,
			Metadata: map[string]interface{}{
				"installed_at": info.ModTime(),
			},
		})
	}

	return skills, nil
}
