package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pryx-core/internal/skills"
)

type skillsConfig struct {
	EnabledSkills map[string]bool `json:"enabled_skills"`
}

func loadSkillsConfig(path string) (*skillsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &skillsConfig{EnabledSkills: map[string]bool{}}, nil
		}
		return nil, err
	}

	var cfg skillsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.EnabledSkills == nil {
		cfg.EnabledSkills = map[string]bool{}
	}
	return &cfg, nil
}

func saveSkillsConfig(path string, cfg *skillsConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".pryx", "skills.yaml")
	}
	return filepath.Join(home, ".pryx", "skills.yaml")
}

func runList(args []string) int {
	eligibleOnly := false
	jsonOutput := false

	for _, arg := range args {
		switch arg {
		case "--eligible", "-e":
			eligibleOnly = true
		case "--json", "-j":
			jsonOutput = true
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", arg)
			usage()
			return 2
		}
	}

	opts := skills.DefaultOptions()
	reg, err := skills.Discover(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover skills: %v\n", err)
		return 1
	}

	all := reg.List()
	var outSkills []skills.Skill
	for _, s := range all {
		if eligibleOnly && !s.Eligible {
			continue
		}
		outSkills = append(outSkills, s)
	}

	sort.Slice(outSkills, func(i, j int) bool { return outSkills[i].ID < outSkills[j].ID })

	if jsonOutput {
		data, err := json.MarshalIndent(outSkills, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal skills: %v\n", err)
			return 1
		}
		fmt.Println(string(data))
		return 0
	}

	fmt.Printf("Available Skills (%d)\n", len(outSkills))
	fmt.Println(strings.Repeat("=", 51))
	for _, s := range outSkills {
		title := s.Frontmatter.Name
		if title == "" {
			title = s.ID
		}
		fmt.Printf("%s: %s\n", s.ID, title)
		if s.Frontmatter.Description != "" {
			fmt.Printf("  %s\n", s.Frontmatter.Description)
		}
		fmt.Printf("  Source: %s\n", s.Source)
	}
	return 0
}

func runInfo(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: skill name required\n")
		return 2
	}
	name := args[0]

	opts := skills.DefaultOptions()
	reg, err := skills.Discover(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover skills: %v\n", err)
		return 1
	}

	skill, ok := reg.Get(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: skill not found: %s\n", name)
		return 1
	}

	fmt.Printf("Skill: %s\n", skill.ID)
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("Title:       %s\n", skill.Frontmatter.Name)
	fmt.Printf("Description: %s\n", skill.Frontmatter.Description)
	fmt.Printf("Source:      %s\n", skill.Source)
	fmt.Printf("Path:        %s\n", skill.Path)
	return 0
}

func runCheck() int {
	opts := skills.DefaultOptions()
	reg, err := skills.Discover(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover skills: %v\n", err)
		return 1
	}
	all := reg.List()

	fmt.Printf("Skills Check\n")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()

	issues := 0
	for _, s := range all {
		if s.ID == "" {
			issues++
			fmt.Printf("✗ <unknown>: missing id\n")
			continue
		}
		if s.Path == "" {
			issues++
			fmt.Printf("✗ %s: missing path\n", s.ID)
			continue
		}
		if s.Frontmatter.Description == "" {
			issues++
			fmt.Printf("✗ %s: missing description\n", s.ID)
			continue
		}
		fmt.Printf("✓ %s\n", s.ID)
	}

	if issues > 0 {
		fmt.Printf("\n✗ Found %d issues\n", issues)
		return 1
	}
	fmt.Printf("\n✓ All skills OK\n")
	return 0
}

func runEnableDisable(name string, enable bool) int {
	opts := skills.DefaultOptions()
	reg, err := skills.Discover(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to discover skills: %v\n", err)
		return 1
	}
	if _, ok := reg.Get(name); !ok {
		fmt.Fprintf(os.Stderr, "Error: skill not found: %s\n", name)
		return 1
	}

	path := configPath()
	cfg, err := loadSkillsConfig(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load skills config: %v\n", err)
		return 1
	}

	if enable {
		cfg.EnabledSkills[name] = true
	} else {
		delete(cfg.EnabledSkills, name)
	}
	if err := saveSkillsConfig(path, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save skills config: %v\n", err)
		return 1
	}

	if enable {
		fmt.Printf("✓ Enabled skill: %s\n", name)
	} else {
		fmt.Printf("✓ Disabled skill: %s\n", name)
	}
	return 0
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "list":
		os.Exit(runList(args))
	case "info":
		os.Exit(runInfo(args))
	case "check":
		os.Exit(runCheck())
	case "enable":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: skill name required\n")
			os.Exit(2)
		}
		os.Exit(runEnableDisable(args[0], true))
	case "disable":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: skill name required\n")
			os.Exit(2)
		}
		os.Exit(runEnableDisable(args[0], false))
	case "install":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: skill name required\n")
			os.Exit(2)
		}
		fmt.Printf("Installing skill: %s\n", args[0])
		os.Exit(0)
	case "uninstall":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: skill name required\n")
			os.Exit(2)
		}
		fmt.Printf("Uninstalling skill: %s\n", args[0])
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
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
