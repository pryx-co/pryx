package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"pryx-core/internal/skills"
	"pryx-core/internal/config"
)

func listSkills(cfg *config.Config, eligibleOnly bool, jsonOutput bool) error {
	opts := skills.DefaultOptions{
		WorkspaceRoot: cfg.GetWorkspaceRoot(),
		ManagedRoot:   cfg.GetSkillsDir(),
	}

	skillsRepo, err := skills.Discover(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	var skillsToDisplay []skills.Skill
	for _, skill := range skillsRepo.List() {
		if eligibleOnly {
			if !skill.Eligible {
				continue
			}
		}
		skillsToDisplay = append(skillsToDisplay, skill)
	}

	// Sort by name
	sort.Slice(skillsToDisplay, func(i, j int) bool {
		return skillsToDisplay[i].Name < skillsToDisplay[j].Name
	})

	if jsonOutput {
		data, err := json.MarshalIndent(skillsToDisplay, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal skills: %w", err)
		}
		fmt.Println(string(data))
	} else {
		fmt.Printf("Available Skills (%d)\n", len(skillsToDisplay))
		fmt.Println("=" + strings.Repeat("-", 50))
		for _, skill := range skillsToDisplay {
			status := ""
			if !skill.Enabled {
				status = " (disabled)"
			} else if skill.Eligible {
				status = " ✓"
			} else {
				status = " ⚠"
			}

			fmt.Printf("%s %s: %s\n", status, skill.Name, skill.Title)
			fmt.Printf("  Description: %s\n", skill.Description)
			fmt.Printf("  Version: %s\n", skill.Version)
			fmt.Printf("  Enabled: %v, Eligible: %v\n", skill.Enabled, skill.Eligible)
		}
	}

	return nil
}

func infoSkill(cfg *config.Config, name string) error {
	opts := skills.DefaultOptions{
		WorkspaceRoot: cfg.GetWorkspaceRoot(),
		ManagedRoot:   cfg.GetSkillsDir(),
	}

	skillsRepo, err := skills.Discover(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	skill, err := skillsRepo.Get(name)
	if err != nil {
		return fmt.Errorf("skill not found: %w", err)
	}

	fmt.Printf("Skill: %s\n", skill.Name)
	fmt.Printf("  Title: %s\n", skill.Title)
	fmt.Printf("  Description: %s\n", skill.Description)
	fmt.Printf("  Version: %s\n", skill.Version)
	fmt.Printf("  Author: %s\n", skill.Author)
	fmt.Printf("  Path: %s\n", skill.Path)
	fmt.Printf("  System Prompt: %d tokens\n", len(skill.SystemPrompt))
	fmt.Printf("  User Prompt: %d tokens\n", len(skill.UserPrompt))
	fmt.Printf("  Enabled: %v\n", skill.Enabled)
	fmt.Printf("  Eligible: %v\n", skill.Eligible)
	fmt.Printf("  Metadata: %v\n", skill.Metadata)

	return nil
}

func checkSkills(cfg *config.Config) error {
	opts := skills.DefaultOptions{
		WorkspaceRoot: cfg.GetWorkspaceRoot(),
		ManagedRoot:   cfg.GetSkillsDir(),
	}

	skillsRepo, err := skills.Discover(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	fmt.Printf("Skills Check\n")
	fmt.Printf("=" + strings.Repeat("-", 40) + "\n\n")

	allSkills := skillsRepo.List()
	if len(allSkills) == 0 {
		fmt.Println("No skills found.")
		return nil
	}

	// Check skills
	validCount := 0
	invalidCount := 0
	issues := 0

	for _, skill := range allSkills {
		issuesInSkill := 0

		// Check SKILL.md exists
		if skill.Path == "" {
			fmt.Printf("✗ %s: No path defined\n", skill.Name)
			issuesInSkill++
			issues++
			continue
		}

		// Check SKILL.md readable
		info, err := skill.Parse()
		if err != nil {
			fmt.Printf("✗ %s: Failed to parse SKILL.md: %v\n", skill.Name, err)
			issuesInSkill++
			issues++
			continue
		}

		if info == nil {
			fmt.Printf("✗ %s: SKILL.md parsing returned nil\n", skill.Name)
			issuesInSkill++
			issues++
			continue
		}

		// Check required fields
		if info.Name == "" {
			fmt.Printf("✗ %s: Missing name\n", skill.Name)
			issuesInSkill++
			issues++
		}
		if info.Version == "" {
			fmt.Printf("✗ %s: Missing version\n", skill.Name)
			issuesInSkill++
			issues++
		}
		if info.Description == "" {
			fmt.Printf("✗ %s: Missing description\n", skill.Name)
			issuesInSkill++
			issues++
		}

		// Check prompts
		if len(info.SystemPrompt) == 0 {
			fmt.Printf("⚠ %s: Empty system prompt\n", skill.Name)
			issuesInSkill++
			issues++
		}

		// Check metadata
		if len(info.Metadata) == 0 {
			fmt.Printf("⚠ %s: No metadata\n", skill.Name)
			issuesInSkill++
			issues++
		}

		if issuesInSkill == 0 {
			fmt.Printf("✓ %s: All checks passed\n", skill.Name)
			validCount++
		} else {
			invalidCount++
		}
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total Skills: %d\n", len(allSkills))
	fmt.Printf("  Valid Skills: %d\n", validCount)
	fmt.Printf("  Invalid Skills: %d\n", invalidCount)
	fmt.Printf("  Total Issues: %d\n", issues)

	if issues == 0 {
		fmt.Printf("\n✓ All skills are properly configured\n")
		return nil
	} else {
		fmt.Printf("\n✗ Found %d issues across %d skills\n", issues, invalidCount)
		return fmt.Errorf("found %d issues", issues)
	}
}

func enableSkill(cfg *config.Config, name string) error {
	opts := skills.DefaultOptions{
		WorkspaceRoot: cfg.GetWorkspaceRoot(),
		ManagedRoot:   cfg.GetSkillsDir(),
	}

	skillsRepo, err := skills.Discover(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	skill, err := skillsRepo.Get(name)
	if err != nil {
		return fmt.Errorf("skill not found: %w", err)
	}

	if !skill.Eligible {
		return fmt.Errorf("skill %s is not eligible", name)
	}

	// Enable skill
	if !skill.Enabled {
		// Update skill enabled status in config
		configPath := cfg.GetSkillsConfigPath()
		skillsConfig, err := loadSkillsConfig(configPath)
		if err == nil {
			if skillsConfig.EnabledSkills == nil {
				skillsConfig.EnabledSkills = make(map[string]bool)
			}
			skillsConfig.EnabledSkills[name] = true
			_ = saveSkillsConfig(configPath, skillsConfig)
		}

		fmt.Printf("✓ Enabled skill: %s\n", name)
	} else {
		fmt.Printf("ℹ Skill %s is already enabled\n", name)
	}

	return nil
}

func disableSkill(cfg *config.Config, name string) error {
	opts := skills.DefaultOptions{
		WorkspaceRoot: cfg.GetWorkspaceRoot(),
		ManagedRoot:   cfg.GetSkillsDir(),
	}

	skillsRepo, err := skills.Discover(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	skill, err := skillsRepo.Get(name)
	if err != nil {
		return fmt.Errorf("skill not found: %w", err)
	}

	// Disable skill
	if skill.Enabled {
		// Update skill enabled status in config
		configPath := cfg.GetSkillsConfigPath()
		skillsConfig, err := loadSkillsConfig(configPath)
		if err == nil {
			if skillsConfig.EnabledSkills == nil {
				skillsConfig.EnabledSkills = make(map[string]bool)
			}
			delete(skillsConfig.EnabledSkills, name)
			_ = saveSkillsConfig(configPath, skillsConfig)
		}

		fmt.Printf("✓ Disabled skill: %s\n", name)
	} else {
		fmt.Printf("ℹ Skill %s is already disabled\n", name)
	}

	return nil
}

func installSkill(cfg *config.Config, name string) error {
	opts := skills.DefaultOptions{
		WorkspaceRoot: cfg.GetWorkspaceRoot(),
		ManagedRoot:   cfg.GetSkillsDir(),
	}

	skillsRepo, err := skills.Discover(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	// Check if skill is available for installation
	skill, err := skillsRepo.Get(name)
	if err != nil {
		// Try to install skill if it exists
		installer, err := skills.GetInstaller(name, opts)
		if err != nil {
			return fmt.Errorf("skill installer not found: %w", err)
		}

		// Run installation
		fmt.Printf("Installing skill: %s\n", name)
		fmt.Printf("  Installer: %s\n", installer.Command)

		// Execute installer
		cmd := exec.Command(installer.Command, installer.Args...)
		cmd.Dir = installer.Directory
		cmd.Env = os.Environ()

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("installation failed: %w\nOutput: %s", err, string(output))
		}

		fmt.Printf("✓ Skill installed successfully\n")
		fmt.Printf("  Output: %s\n", string(output))

		return nil
	}

	return fmt.Errorf("skill not found or not available for installation")
}

func loadSkillsConfig(path string) (*skillsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &skillsConfig{EnabledSkills: make(map[string]bool)}, nil
		}
		return nil, err
	}

	var config skillsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveSkillsConfig(path string, config *skillsConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

type skillsConfig struct {
	EnabledSkills map[string]bool `json:"enabled_skills"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	cfg := config.Load()
	opts := skills.DefaultOptions{
		WorkspaceRoot: cfg.GetWorkspaceRoot(),
		ManagedRoot:   cfg.GetSkillsDir(),
	}

	switch cmd {
	case "list":
		eligibleOnly := false
		jsonOutput := false

		for i, arg := range os.Args[2:] {
			switch arg {
			case "--eligible", "-e":
				eligibleOnly = true
			case "--json", "-j":
				jsonOutput = true
			}
		}

		if err := listSkills(cfg, eligibleOnly, jsonOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "info":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: skill name required\n")
			os.Exit(1)
		}
		name := os.Args[2]

		if err := infoSkill(cfg, name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "check":
		if err := checkSkills(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "enable":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: skill name required\n")
			os.Exit(1)
		}
		name := os.Args[2]

		if err := enableSkill(cfg, name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "disable":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: skill name required\n")
			os.Exit(1)
		}
		name := os.Args[2]

		if err := disableSkill(cfg, name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "install":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: skill name required\n")
			os.Exit(1)
		}
		name := os.Args[2]

		if err := installSkill(cfg, name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		usage()
		os.Exit(1)
	}

func usage() {
	fmt.Println("pryx-core skills - Manage Pryx skills")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  list [--eligible] [--json]        List available skills")
	fmt.Println("  info <name>                         Show skill details")
	fmt.Println("  check                                Check all skills for issues")
	fmt.Println("  enable <name>                       Enable a skill")
	fmt.Println("  disable <name>                      Disable a skill")
	fmt.Println("  install <name>                       Install a skill")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --eligible, -e                        Show only eligible skills")
	fmt.Println("  --json, -j                             Output in JSON format")
}
