package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()
	assert.NotNil(t, reg)
	assert.NotNil(t, reg.skills)
	assert.Empty(t, reg.skills)
}

func TestRegistry_Upsert(t *testing.T) {
	reg := NewRegistry()

	skill := Skill{
		ID: "test-skill",
		Frontmatter: Frontmatter{
			Name:        "Test Skill",
			Description: "A test skill",
			Metadata: SkillMetadata{
				Pryx: PryxMetadata{
					Requires: Requirements{
						Bins: []string{"git"},
						Env:  []string{"HOME"},
					},
				},
			},
		},
	}

	reg.Upsert(skill)

	// Verify skill was added
	found, ok := reg.Get("test-skill")
	assert.True(t, ok)
	assert.Equal(t, skill.ID, found.ID)
	assert.Equal(t, skill.Frontmatter.Name, found.Frontmatter.Name)
}

func TestRegistry_Upsert_UpdateExisting(t *testing.T) {
	reg := NewRegistry()

	// Insert initial skill
	skill1 := Skill{
		ID:          "test-skill",
		Frontmatter: Frontmatter{Name: "Original Name"},
	}
	reg.Upsert(skill1)

	// Update the skill
	skill2 := Skill{
		ID:          "test-skill",
		Frontmatter: Frontmatter{Name: "Updated Name"},
	}
	reg.Upsert(skill2)

	// Verify it was updated
	found, ok := reg.Get("test-skill")
	assert.True(t, ok)
	assert.Equal(t, "Updated Name", found.Frontmatter.Name)
}

func TestRegistry_Get(t *testing.T) {
	reg := NewRegistry()

	// Try to get non-existent skill
	_, ok := reg.Get("nonexistent")
	assert.False(t, ok)

	// Add and retrieve skill
	skill := Skill{ID: "exists"}
	reg.Upsert(skill)

	found, ok := reg.Get("exists")
	assert.True(t, ok)
	assert.Equal(t, "exists", found.ID)
}

func TestRegistry_Get_Concurrent(t *testing.T) {
	reg := NewRegistry()

	// Add a skill
	reg.Upsert(Skill{ID: "concurrent-test"})

	// Concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, ok := reg.Get("concurrent-test")
			done <- ok
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		require.True(t, <-done)
	}
}

func TestRegistry_List(t *testing.T) {
	reg := NewRegistry()

	// Empty registry
	list := reg.List()
	assert.Empty(t, list)

	// Add skills
	skills := []Skill{
		{ID: "skill-c"},
		{ID: "skill-a"},
		{ID: "skill-b"},
	}

	for _, s := range skills {
		reg.Upsert(s)
	}

	// List should be sorted by ID
	list = reg.List()
	require.Len(t, list, 3)
	assert.Equal(t, "skill-a", list[0].ID)
	assert.Equal(t, "skill-b", list[1].ID)
	assert.Equal(t, "skill-c", list[2].ID)
}

func TestRegistry_List_Concurrent(t *testing.T) {
	reg := NewRegistry()

	// Add initial skills
	for i := 0; i < 5; i++ {
		reg.Upsert(Skill{ID: string(rune('a' + i))})
	}

	// Concurrent reads and writes
	done := make(chan bool, 20)

	// Readers
	for i := 0; i < 10; i++ {
		go func() {
			_ = reg.List()
			done <- true
		}()
	}

	// Writers
	for i := 0; i < 10; i++ {
		go func(index int) {
			reg.Upsert(Skill{ID: string(rune('z' - index))})
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 20; i++ {
		<-done
	}

	// Registry should still be functional
	list := reg.List()
	assert.NotNil(t, list)
}

func TestRegistry_MetadataSummary(t *testing.T) {
	tests := []struct {
		name     string
		skills   []Skill
		expected string
	}{
		{
			name:     "empty registry",
			skills:   []Skill{},
			expected: "",
		},
		{
			name: "single skill",
			skills: []Skill{
				{
					ID: "git",
					Frontmatter: Frontmatter{
						Description: "Git operations",
					},
				},
			},
			expected: "- git: Git operations",
		},
		{
			name: "skill with requirements",
			skills: []Skill{
				{
					ID: "docker",
					Frontmatter: Frontmatter{
						Description: "Docker operations",
						Metadata: SkillMetadata{
							Pryx: PryxMetadata{
								Requires: Requirements{
									Bins: []string{"docker", "docker-compose"},
								},
							},
						},
					},
				},
			},
			expected: "- docker: Docker operations (bins: docker, docker-compose)",
		},
		{
			name: "skill with env requirements",
			skills: []Skill{
				{
					ID: "deploy",
					Frontmatter: Frontmatter{
						Description: "Deployment skill",
						Metadata: SkillMetadata{
							Pryx: PryxMetadata{
								Requires: Requirements{
									Env: []string{"AWS_ACCESS_KEY", "AWS_SECRET_KEY"},
								},
							},
						},
					},
				},
			},
			expected: "- deploy: Deployment skill (env: AWS_ACCESS_KEY, AWS_SECRET_KEY)",
		},
		{
			name: "skill with both requirements",
			skills: []Skill{
				{
					ID: "k8s",
					Frontmatter: Frontmatter{
						Description: "Kubernetes operations",
						Metadata: SkillMetadata{
							Pryx: PryxMetadata{
								Requires: Requirements{
									Bins: []string{"kubectl"},
									Env:  []string{"KUBECONFIG"},
								},
							},
						},
					},
				},
			},
			expected: "- k8s: Kubernetes operations (bins: kubectl; env: KUBECONFIG)",
		},
		{
			name: "multiple skills",
			skills: []Skill{
				{ID: "alpha", Frontmatter: Frontmatter{Description: "First"}},
				{ID: "beta", Frontmatter: Frontmatter{Description: "Second"}},
			},
			expected: "- alpha: First\n- beta: Second",
		},
		{
			name: "skill without description",
			skills: []Skill{
				{ID: "minimal"},
			},
			expected: "- minimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry()
			for _, s := range tt.skills {
				reg.Upsert(s)
			}

			result := reg.MetadataSummary()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegistry_MetadataSummary_NoDescription(t *testing.T) {
	reg := NewRegistry()
	reg.Upsert(Skill{
		ID: "no-desc-skill",
		Frontmatter: Frontmatter{
			Metadata: SkillMetadata{
				Pryx: PryxMetadata{
					Requires: Requirements{
						Bins: []string{"tool"},
					},
				},
			},
		},
	})

	result := reg.MetadataSummary()
	assert.Equal(t, "- no-desc-skill (bins: tool)", result)
}

func BenchmarkRegistry_Upsert(b *testing.B) {
	reg := NewRegistry()
	skill := Skill{ID: "benchmark-skill"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		skill.ID = string(rune(i))
		reg.Upsert(skill)
	}
}

func BenchmarkRegistry_Get(b *testing.B) {
	reg := NewRegistry()
	reg.Upsert(Skill{ID: "test-skill"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reg.Get("test-skill")
	}
}

func BenchmarkRegistry_List(b *testing.B) {
	reg := NewRegistry()
	for i := 0; i < 100; i++ {
		reg.Upsert(Skill{ID: string(rune(i))})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reg.List()
	}
}
