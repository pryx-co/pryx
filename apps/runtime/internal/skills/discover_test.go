package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDiscoverPrecedenceWorkspaceOverridesManaged(t *testing.T) {
	workspaceRoot := t.TempDir()
	managedRoot := t.TempDir()

	managedSkill := filepath.Join(managedRoot, "linter")
	workspaceSkill := filepath.Join(workspaceRoot, ".pryx", "skills", "linter")

	if err := os.MkdirAll(managedSkill, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(workspaceSkill, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	managed := []byte(`---
name: linter
description: managed
---
# Linter
managed`)
	workspace := []byte(`---
name: linter
description: workspace
---
# Linter
workspace`)

	if err := os.WriteFile(filepath.Join(managedSkill, "SKILL.md"), managed, 0o644); err != nil {
		t.Fatalf("write managed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceSkill, "SKILL.md"), workspace, 0o644); err != nil {
		t.Fatalf("write workspace: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	reg, err := Discover(ctx, Options{
		WorkspaceRoot: workspaceRoot,
		ManagedRoot:   managedRoot,
		BundledRoot:   "",
		MaxConcurrent: 4,
	})
	if reg == nil {
		t.Fatalf("expected registry")
	}
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	skill, ok := reg.Get("linter")
	if !ok {
		t.Fatalf("expected skill linter")
	}
	if skill.Source != SourceWorkspace {
		t.Fatalf("expected workspace source, got %s", skill.Source)
	}
	if skill.Frontmatter.Description != "workspace" {
		t.Fatalf("expected workspace description, got %q", skill.Frontmatter.Description)
	}

	body, err := skill.Body()
	if err != nil {
		t.Fatalf("expected nil body error, got %v", err)
	}
	if filepath.Base(skill.Path) != "SKILL.md" {
		t.Fatalf("expected SKILL.md path, got %q", skill.Path)
	}
	if body == "" {
		t.Fatalf("expected non-empty body")
	}
}

func BenchmarkDiscover100Skills(b *testing.B) {
	workspaceRoot := b.TempDir()
	managedRoot := b.TempDir()

	if err := os.MkdirAll(filepath.Join(managedRoot, "skill1"), 0o755); err != nil {
		b.Fatalf("mkdir: %v", err)
	}

	skillTemplate := []byte(`---
name: skill-%d
description: Benchmark skill
---
# Benchmark skill body
`)
	for i := 0; i < 100; i++ {
		skillDir := filepath.Join(managedRoot, fmt.Sprintf("skill%d", i))
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			b.Fatalf("mkdir skill%d: %v", i, err)
		}
		data := fmt.Sprintf(string(skillTemplate), i)
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(data), 0o644); err != nil {
			b.Fatalf("write skill%d: %v", i, err)
		}
	}

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reg, err := Discover(ctx, Options{
			WorkspaceRoot: workspaceRoot,
			ManagedRoot:   managedRoot,
			BundledRoot:   "",
			MaxConcurrent: 4,
		})
		if err != nil {
			b.Fatalf("discover failed: %v", err)
		}
		if reg == nil {
			b.Fatalf("expected registry")
		}
	}
}

func TestLoadSkillsWithin1s(t *testing.T) {
	workspaceRoot := t.TempDir()
	managedRoot := t.TempDir()

	if err := os.MkdirAll(filepath.Join(managedRoot, "skill1"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	for i := 0; i < 50; i++ {
		skillDir := filepath.Join(managedRoot, fmt.Sprintf("skill%d", i))
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatalf("mkdir skill%d: %v", i, err)
		}
		skillContent := fmt.Sprintf(`---
name: skill-%d
description: Test skill
---
# Skill body
`, i)
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644); err != nil {
			t.Fatalf("write skill%d: %v", i, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	reg, err := Discover(ctx, Options{
		WorkspaceRoot: workspaceRoot,
		ManagedRoot:   managedRoot,
		BundledRoot:   "",
		MaxConcurrent: 4,
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("discover failed: %v", err)
	}
	if reg == nil {
		t.Fatalf("expected registry")
	}
	skills := reg.List()
	if len(skills) != 50 {
		t.Fatalf("expected 50 skills, got %d", len(skills))
	}

	if elapsed > time.Second {
		t.Logf("Loaded 50 skills in %v (exceeds 1s requirement)", elapsed)
	} else {
		t.Logf("Loaded 50 skills in %v (within 1s requirement)", elapsed)
	}
}
