package interfaces

import "context"

// Skill represents a composable, reusable capability that can be
// attached to an agent. Skills can provide:
//   - System prompt augmentation (instructions, domain knowledge)
//   - Specialized tools
//   - Sub-agents
//   - Pre/post processing hooks
type Skill interface {
	// Name returns the unique name of the skill.
	Name() string

	// Description returns a human-readable description.
	Description() string

	// Category returns the skill category (e.g., "code", "data", "web").
	Category() string

	// SystemPromptAugment returns additional system prompt content
	// that gets appended when the skill is active.
	SystemPromptAugment() string

	// Tools returns any specialized tools this skill provides.
	Tools() []Tool

	// Init is called when the skill is loaded into an agent.
	// It can be used for one-time setup.
	Init(ctx context.Context) error

	// PreRun is called before each agent run with the user prompt.
	// It can modify or enrich the prompt.
	PreRun(ctx context.Context, prompt string) (string, error)

	// PostRun is called after each agent run with the final response.
	PostRun(ctx context.Context, response string) (string, error)
}

// SkillRegistry manages the lifecycle of loaded skills.
type SkillRegistry interface {
	// Register adds a skill to the registry.
	Register(skill Skill) error

	// Get returns a skill by name.
	Get(name string) (Skill, bool)

	// List returns all registered skills.
	List() []Skill

	// ListByCategory returns skills in a given category.
	ListByCategory(category string) []Skill

	// ActiveSkills returns the set of currently active skill names.
	ActiveSkills() []string

	// Activate marks a skill as active for the agent.
	Activate(name string) error

	// Deactivate marks a skill as inactive.
	Deactivate(name string) error

	// GetCombinedSystemPrompt returns the combined system prompt
	// augmentation from all active skills.
	GetCombinedSystemPrompt() string
}
