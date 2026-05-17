// Package skills implements the Skills system for agents.
package skills

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/user/agent-sdk/pkg/interfaces"
)

// Registry implements interfaces.SkillRegistry.
type Registry struct {
	mu       sync.RWMutex
	skills   map[string]interfaces.Skill
	active   map[string]bool // set of active skill names
}

// NewRegistry creates a new skill registry.
func NewRegistry() *Registry {
	return &Registry{
		skills: make(map[string]interfaces.Skill),
		active: make(map[string]bool),
	}
}

// Register adds a skill to the registry.
func (r *Registry) Register(skill interfaces.Skill) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if skill.Name() == "" {
		return fmt.Errorf("skill name cannot be empty")
	}
	if _, exists := r.skills[skill.Name()]; exists {
		return fmt.Errorf("skill %q already registered", skill.Name())
	}
	r.skills[skill.Name()] = skill
	return nil
}

// Get returns a skill by name.
func (r *Registry) Get(name string) (interfaces.Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.skills[name]
	return s, ok
}

// List returns all registered skills.
func (r *Registry) List() []interfaces.Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]interfaces.Skill, 0, len(r.skills))
	for _, s := range r.skills {
		out = append(out, s)
	}
	return out
}

// ListByCategory returns skills filtered by category.
func (r *Registry) ListByCategory(category string) []interfaces.Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []interfaces.Skill
	for _, s := range r.skills {
		if s.Category() == category {
			out = append(out, s)
		}
	}
	return out
}

// ActiveSkills returns the names of currently active skills.
func (r *Registry) ActiveSkills() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.active))
	for name := range r.active {
		out = append(out, name)
	}
	return out
}

// Activate enables a skill for the agent.
func (r *Registry) Activate(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	skill, ok := r.skills[name]
	if !ok {
		return fmt.Errorf("skill %q not found", name)
	}
	if err := skill.Init(context.Background()); err != nil {
		return fmt.Errorf("init skill %q: %w", name, err)
	}
	r.active[name] = true
	return nil
}

// Deactivate disables a skill.
func (r *Registry) Deactivate(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.skills[name]; !ok {
		return fmt.Errorf("skill %q not found", name)
	}
	delete(r.active, name)
	return nil
}

// GetCombinedSystemPrompt returns the concatenated system prompt
// augmentations from all active skills.
func (r *Registry) GetCombinedSystemPrompt() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var parts []string
	for name := range r.active {
		if s, ok := r.skills[name]; ok {
			if aug := s.SystemPromptAugment(); aug != "" {
				parts = append(parts, fmt.Sprintf("## Skill: %s\n%s", s.Name(), aug))
			}
		}
	}
	return strings.Join(parts, "\n\n")
}

// GetActiveSkills returns the actual Skill objects for all active skills.
func (r *Registry) GetActiveSkills() []interfaces.Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []interfaces.Skill
	for name := range r.active {
		if s, ok := r.skills[name]; ok {
			out = append(out, s)
		}
	}
	return out
}

// GetAllActiveTools collects tools from all active skills.
func (r *Registry) GetAllActiveTools() []interfaces.Tool {
	var tools []interfaces.Tool
	for _, s := range r.GetActiveSkills() {
		tools = append(tools, s.Tools()...)
	}
	return tools
}

// LoadDirectory scans a directory for Agent Skills (SKILL.md files) and
// registers all found skills into this registry. Uses the agentskills.io format.
// Returns the number of skills loaded.
func (r *Registry) LoadDirectory(root string, opts ...FileSkillOption) (int, error) {
	skills, err := LoadSkillsFromDirectory(root, opts...)
	if err != nil {
		return 0, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for name, fs := range skills {
		if _, exists := r.skills[name]; exists {
			continue // don't overwrite existing
		}
		r.skills[name] = fs
		count++
	}
	return count, nil
}

// LoadFileSkill loads a single skill from a directory path and registers it.
func (r *Registry) LoadFileSkill(dirPath string, opts ...FileSkillOption) error {
	fs, err := LoadSkillFromPath(dirPath, opts...)
	if err != nil {
		return err
	}
	return r.Register(fs)
}
