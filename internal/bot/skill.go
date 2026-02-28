package bot

import (
	"fmt"
	"sort"
	"sync"
)

// Skill represents an intent recognition skill.
type Skill struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Intent      Intent              `json:"intent"`
	Priority    int                 `json:"priority"`    // 1-100, lower = higher priority
	Enabled     bool                `json:"enabled"`
	Builtin     bool                `json:"builtin"`
	Keywords    map[string][]string `json:"keywords"`              // lang -> keywords
	Synonyms    map[string]string   `json:"synonyms,omitempty"`    // variant -> canonical
	Examples    []string            `json:"examples,omitempty"`
	Stats       SkillStats          `json:"stats"`
}

// SkillStats tracks match statistics for a skill.
type SkillStats struct {
	TotalMatches int `json:"total_matches"`
	LocalMatches int `json:"local_matches"`
	LLMMatches   int `json:"llm_matches"`
}

// Validate checks that the skill has all required fields.
func (s *Skill) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if s.Description == "" {
		return fmt.Errorf("skill %q: description is required", s.Name)
	}
	if s.Intent == "" {
		return fmt.Errorf("skill %q: intent is required", s.Name)
	}
	if s.Priority < 1 || s.Priority > 100 {
		return fmt.Errorf("skill %q: priority must be 1-100, got %d", s.Name, s.Priority)
	}
	if len(s.Keywords) == 0 {
		return fmt.Errorf("skill %q: at least one language keyword set is required", s.Name)
	}
	for lang, kws := range s.Keywords {
		if len(kws) == 0 {
			return fmt.Errorf("skill %q: keyword list for %q is empty", s.Name, lang)
		}
	}
	return nil
}

// SkillRegistry manages registered skills.
type SkillRegistry struct {
	mu     sync.RWMutex
	skills map[string]*Skill
}

// NewSkillRegistry creates a new empty skill registry.
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make(map[string]*Skill),
	}
}

// Register adds a skill to the registry. Returns error if invalid or duplicate.
func (r *SkillRegistry) Register(s *Skill) error {
	if err := s.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.skills[s.Name]; exists {
		return fmt.Errorf("skill %q already registered", s.Name)
	}
	r.skills[s.Name] = s
	return nil
}

// Get returns a skill by name, or nil if not found.
func (r *SkillRegistry) Get(name string) *Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.skills[name]
}

// List returns all skills sorted by priority (ascending).
func (r *SkillRegistry) List() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*Skill, 0, len(r.skills))
	for _, s := range r.skills {
		list = append(list, s)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Priority < list[j].Priority
	})
	return list
}

// ListEnabled returns all enabled skills sorted by priority (ascending).
func (r *SkillRegistry) ListEnabled() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*Skill, 0, len(r.skills))
	for _, s := range r.skills {
		if s.Enabled {
			list = append(list, s)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Priority < list[j].Priority
	})
	return list
}

// SetEnabled enables or disables a skill by name.
func (r *SkillRegistry) SetEnabled(name string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.skills[name]
	if !ok {
		return fmt.Errorf("skill %q not found", name)
	}
	s.Enabled = enabled
	return nil
}
