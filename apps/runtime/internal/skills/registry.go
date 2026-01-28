package skills

import (
	"sort"
	"strings"
	"sync"
)

type Registry struct {
	mu     sync.RWMutex
	skills map[string]Skill
}

func NewRegistry() *Registry {
	return &Registry{
		skills: map[string]Skill{},
	}
}

func (r *Registry) Upsert(skill Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills[skill.ID] = skill
}

func (r *Registry) Get(id string) (Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.skills[id]
	return s, ok
}

func (r *Registry) List() []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Skill, 0, len(r.skills))
	for _, s := range r.skills {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func (r *Registry) MetadataSummary() string {
	skills := r.List()
	if len(skills) == 0 {
		return ""
	}
	var b strings.Builder
	for i, s := range skills {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("- ")
		b.WriteString(s.ID)
		if s.Frontmatter.Description != "" {
			b.WriteString(": ")
			b.WriteString(s.Frontmatter.Description)
		}
		req := s.Frontmatter.Metadata.Pryx.Requires
		if len(req.Bins) > 0 || len(req.Env) > 0 {
			b.WriteString(" (")
			if len(req.Bins) > 0 {
				b.WriteString("bins: ")
				b.WriteString(strings.Join(req.Bins, ", "))
				if len(req.Env) > 0 {
					b.WriteString("; ")
				}
			}
			if len(req.Env) > 0 {
				b.WriteString("env: ")
				b.WriteString(strings.Join(req.Env, ", "))
			}
			b.WriteString(")")
		}
	}
	return b.String()
}
