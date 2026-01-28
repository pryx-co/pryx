package skills

import "testing"

func TestParseSkillFile(t *testing.T) {
	data := []byte(`---
name: my-skill
description: What triggers this skill
metadata:
  pryx:
    emoji: "🔧"
    requires:
      bins: ["jq"]
      env: ["API_KEY"]
    install:
      - id: brew
        kind: brew
        formula: jq
        bins: ["jq"]
---
# My Skill

Instructions...`)

	fm, body, err := parseSkillFile(data)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if fm.Name != "my-skill" {
		t.Fatalf("expected name my-skill, got %q", fm.Name)
	}
	if fm.Metadata.Pryx.Emoji != "🔧" {
		t.Fatalf("expected emoji 🔧, got %q", fm.Metadata.Pryx.Emoji)
	}
	if len(fm.Metadata.Pryx.Requires.Bins) != 1 || fm.Metadata.Pryx.Requires.Bins[0] != "jq" {
		t.Fatalf("expected bins [jq], got %+v", fm.Metadata.Pryx.Requires.Bins)
	}
	if body == "" {
		t.Fatalf("expected non-empty body")
	}
}

func TestParseSkillFileMissingFrontmatter(t *testing.T) {
	_, _, err := parseSkillFile([]byte("# No frontmatter"))
	if err == nil {
		t.Fatalf("expected error")
	}
}
