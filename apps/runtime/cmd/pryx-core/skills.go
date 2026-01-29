package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pryx-core/internal/config"
	"pryx-core/internal/skills"
)

func runSkills(args []string) int {
	if len(args) < 1 {
		skillsUsage()
		return 2
	}

	cmd := args[0]
	cfg := config.Load()

	switch cmd {
	case "list", "ls":
		return runListSkills(args[1:], cfg)
	case "info":
		return runInfoSkill(args[1:], cfg)
	case "check":
		return runCheckSkills(args[1:], cfg)
	case "enable":
		return runEnableSkill(args[1:], cfg)
	case "disable":
		return runDisableSkill(args[1:], cfg)
	case "install":
		return runInstallSkill(args[1:], cfg)
	case "uninstall":
		return runUninstallSkill(args[1:], cfg)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		skillsUsage()
		return 2
	}
}

func runListSkills(args []string, cfg *config.Config) int {
	eligibleOnly := false
	jsonOutput := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--eligible", "-e":
			eligibleOnly = true
		case "--json", "-j":
			jsonOutput = true
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", arg)
			skillsUsage()
			return 2
		}
	}

	opts := skills.DefaultOptions()

	skillsRepo, err := skills.Discover(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover skills: %v\n", err)
		return 1
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
		return skillsToDisplay[i].ID < skillsToDisplay[j].ID
	})

	if jsonOutput {
		data, err := json.MarshalIndent(skillsToDisplay, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal skills: %v\n", err)
			return 1
		}
		fmt.Println(string(data))
	} else {
		fmt.Printf("Available Skills (%d)\n", len(skillsToDisplay))
		fmt.Println(strings.Repeat("=", 51))
		for _, skill := range skillsToDisplay {
			status := ""
			if !skill.Enabled {
				status = " (disabled)"
			} else if skill.Eligible {
				status = " ✓"
			} else {
				status = " ⚠"
			}

			title := skill.Frontmatter.Name
			if title == "" {
				title = skill.ID
			}
			fmt.Printf("%s %s: %s\n", status, skill.ID, title)
			if skill.Frontmatter.Description != "" {
				fmt.Printf("  %s\n", skill.Frontmatter.Description)
			}
			fmt.Printf("  Source: %s, Enabled: %v\n", skill.Source, skill.Enabled)
		}
	}

	return 0
}

func runInfoSkill(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: skill name required\n")
		return 2
	}
	name := args[0]

	opts := skills.DefaultOptions()

	skillsRepo, err := skills.Discover(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover skills: %v\n", err)
		return 1
	}

	skill, found := skillsRepo.Get(name)
	if !found {
		fmt.Fprintf(os.Stderr, "Error: skill not found: %s\n", name)
		return 1
	}

	fmt.Printf("Skill: %s\n", skill.ID)
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("Title:       %s\n", skill.Frontmatter.Name)
	fmt.Printf("Description: %s\n", skill.Frontmatter.Description)
	fmt.Printf("Source:      %s\n", skill.Source)
	fmt.Printf("Path:        %s\n", skill.Path)
	fmt.Printf("Enabled:     %v\n", skill.Enabled)
	fmt.Printf("Eligible:    %v\n", skill.Eligible)

	if len(skill.Frontmatter.Metadata.Pryx.Requires.Bins) > 0 {
		fmt.Printf("Required binaries: %s\n", strings.Join(skill.Frontmatter.Metadata.Pryx.Requires.Bins, ", "))
	}
	if len(skill.Frontmatter.Metadata.Pryx.Requires.Env) > 0 {
		fmt.Printf("Required env vars: %s\n", strings.Join(skill.Frontmatter.Metadata.Pryx.Requires.Env, ", "))
	}
	if len(skill.Frontmatter.Metadata.Pryx.Install) > 0 {
		fmt.Printf("Installers: %d\n", len(skill.Frontmatter.Metadata.Pryx.Install))
		for i, installer := range skill.Frontmatter.Metadata.Pryx.Install {
			fmt.Printf("  [%d] %s %s\n", i+1, installer.Command, strings.Join(installer.Args, " "))
		}
	}

	return 0
}

func runCheckSkills(args []string, cfg *config.Config) int {
	opts := skills.DefaultOptions()

	skillsRepo, err := skills.Discover(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover skills: %v\n", err)
		return 1
	}

	fmt.Printf("Skills Check\n")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()

	allSkills := skillsRepo.List()
	if len(allSkills) == 0 {
		fmt.Println("No skills found.")
		return 0
	}

	// Check skills
	validCount := 0
	invalidCount := 0
	issues := 0

	for _, skill := range allSkills {
		issuesInSkill := 0

		// Check SKILL.md exists
		if skill.Path == "" {
			fmt.Printf("✗ %s: No path defined\n", skill.ID)
			issuesInSkill++
			issues++
			continue
		}

		// Check required fields
		if skill.ID == "" {
			fmt.Printf("✗ %s: Missing name\n", skill.ID)
			issuesInSkill++
			issues++
		}
		// Version field doesn't exist in current Frontmatter
		if skill.Frontmatter.Description == "" {
			fmt.Printf("✗ %s: Missing description\n", skill.ID)
			issuesInSkill++
			issues++
		}

		// Check prompts
		if len(skill.SystemPrompt) == 0 {
			fmt.Printf("⚠ %s: Empty system prompt\n", skill.ID)
			issuesInSkill++
			issues++
		}

		if issuesInSkill == 0 {
			fmt.Printf("✓ %s: All checks passed\n", skill.ID)
			validCount++
		} else {
			invalidCount++
		}
	}

	fmt.Println()
	fmt.Printf("Summary:\n")
	fmt.Printf("  Total Skills:  %d\n", len(allSkills))
	fmt.Printf("  Valid Skills:  %d\n", validCount)
	fmt.Printf("  Invalid Skills: %d\n", invalidCount)
	fmt.Printf("  Total Issues:  %d\n", issues)

	if issues == 0 {
		fmt.Println()
		fmt.Printf("✓ All skills are properly configured\n")
		return 0
	} else {
		fmt.Println()
		fmt.Printf("✗ Found %d issues across %d skills\n", issues, invalidCount)
		return 1
	}
}

func runEnableSkill(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: skill name required\n")
		return 2
	}
	name := args[0]

	opts := skills.DefaultOptions()

	skillsRepo, err := skills.Discover(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover skills: %v\n", err)
		return 1
	}

	_, found := skillsRepo.Get(name)
	if !found {
		fmt.Fprintf(os.Stderr, "Error: skill not found: %s\n", name)
		return 1
	}

	// Update skill enabled status in config
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".pryx", "skills.yaml")
	skillsConfig, err := loadSkillsConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load skills config: %v\n", err)
		return 1
	}

	if skillsConfig.EnabledSkills == nil {
		skillsConfig.EnabledSkills = make(map[string]bool)
	}

	if skillsConfig.EnabledSkills[name] {
		fmt.Printf("ℹ Skill %s is already enabled\n", name)
	} else {
		skillsConfig.EnabledSkills[name] = true
		if err := saveSkillsConfig(configPath, skillsConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to save skills config: %v\n", err)
			return 1
		}
		fmt.Printf("✓ Enabled skill: %s\n", name)
	}

	return 0
}

func runDisableSkill(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: skill name required\n")
		return 2
	}
	name := args[0]

	opts := skills.DefaultOptions()

	skillsRepo, err := skills.Discover(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover skills: %v\n", err)
		return 1
	}

	_, found := skillsRepo.Get(name)
	if !found {
		fmt.Fprintf(os.Stderr, "Error: skill not found: %s\n", name)
		return 1
	}

	// Update skill enabled status in config
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".pryx", "skills.yaml")
	skillsConfig, err := loadSkillsConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load skills config: %v\n", err)
		return 1
	}

	if skillsConfig.EnabledSkills == nil {
		skillsConfig.EnabledSkills = make(map[string]bool)
	}

	if !skillsConfig.EnabledSkills[name] {
		fmt.Printf("ℹ Skill %s is already disabled\n", name)
	} else {
		delete(skillsConfig.EnabledSkills, name)
		if err := saveSkillsConfig(configPath, skillsConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to save skills config: %v\n", err)
			return 1
		}
		fmt.Printf("✓ Disabled skill: %s\n", name)
	}

	return 0
}

func runInstallSkill(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: skill name required\n")
		return 2
	}
	name := args[0]

	fmt.Printf("Installing skill: %s\n", name)
	fmt.Println("(Installation logic to be implemented)")
	fmt.Printf("✓ Skill installation prepared: %s\n", name)

	return 0
}

func runUninstallSkill(args []string, cfg *config.Config) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: skill name required\n")
		return 2
	}
	name := args[0]

	fmt.Printf("Uninstalling skill: %s\n", name)
	fmt.Println("(Uninstallation logic to be implemented)")
	fmt.Printf("✓ Skill uninstallation prepared: %s\n", name)

	return 0
}

func skillsUsage() {
	fmt.Println("pryx-core skills - Manage Pryx skills")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  list [--eligible] [--json]        List available skills")
	fmt.Println("  info <name>                       Show skill details")
	fmt.Println("  check                             Check all skills for issues")
	fmt.Println("  enable <name>                     Enable a skill")
	fmt.Println("  disable <name>                    Disable a skill")
	fmt.Println("  install <name>                    Install a skill")
	fmt.Println("  uninstall <name>                  Uninstall a skill")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --eligible, -e                    Show only eligible skills")
	fmt.Println("  --json, -j                        Output in JSON format")
}

type skillsConfig struct {
	EnabledSkills map[string]bool `json:"enabled_skills"`
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
