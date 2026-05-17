// Package agentsdk is the top-level convenience package for the Agent SDK.
package agentsdk

import (
	"context"
	"fmt"

	"github.com/cyberYiz/agent-sdk/pkg/agent"
	"github.com/cyberYiz/agent-sdk/pkg/config"
	"github.com/cyberYiz/agent-sdk/pkg/interfaces"
	"github.com/cyberYiz/agent-sdk/pkg/llm/openai"
	"github.com/cyberYiz/agent-sdk/pkg/logging"
	"github.com/cyberYiz/agent-sdk/pkg/memory"
	"github.com/cyberYiz/agent-sdk/pkg/skills"
	"github.com/cyberYiz/agent-sdk/pkg/skills/builtin"
	"github.com/cyberYiz/agent-sdk/pkg/tools"
	"github.com/cyberYiz/agent-sdk/pkg/tools/calculator"
	"github.com/cyberYiz/agent-sdk/pkg/tools/websearch"
)

// DefaultAgent creates an agent with sensible defaults using environment config.
// It sets up OpenAI LLM, in-memory conversation buffer, and basic tools.
func DefaultAgent(ctx context.Context) (*agent.Agent, error) {
	cfg := config.Get()

	if cfg.LLM.OpenAI.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	// Create logger
	logger := logging.New()

	// Create LLM client
	llmClient := openai.NewClient(cfg.LLM.OpenAI.APIKey,
		openai.WithModel(cfg.LLM.OpenAI.Model),
	)

	// Create memory
	mem := memory.NewConversationBuffer()

	// Create tool registry with built-in tools
	tr := tools.NewRegistry()
	tr.Register(calculator.New())
	tr.Register(websearch.New())

	// Create skill registry with built-in skills
	sr := skills.NewRegistry()
	sr.Register(builtin.NewCodeReviewSkill())
	sr.Register(builtin.NewDataAnalysisSkill())
	sr.Register(builtin.NewWebResearchSkill())
	sr.Register(builtin.NewCreativeWritingSkill())

	ag, err := agent.NewAgent(
		agent.WithLLM(llmClient),
		agent.WithMemory(mem),
		agent.WithToolRegistry(tr),
		agent.WithSkillRegistry(sr),
		agent.WithSystemPrompt("You are a helpful AI assistant with access to tools and skills. Use tools when you need to search the web or perform calculations. Activate relevant skills when the task requires specialized expertise."),
		agent.WithName("AgentSDK"),
		agent.WithMaxIterations(config.GetMaxIterations()),
	)
	if err != nil {
		return nil, err
	}

	_ = logger
	return ag, nil
}

// QuickRun is the fastest way to get a response.
func QuickRun(ctx context.Context, prompt string) (string, error) {
	ag, err := DefaultAgent(ctx)
	if err != nil {
		return "", err
	}
	return ag.Run(ctx, prompt)
}

// LoadSkillsFromDir is a convenience function that loads all agentskills-format
// skills (SKILL.md files) from a directory into a new skill registry.
func LoadSkillsFromDir(dir string) (*skills.Registry, error) {
	sr := skills.NewRegistry()
	if _, err := sr.LoadDirectory(dir); err != nil {
		return nil, err
	}
	return sr, nil
}

// QuickAgentWithSkills creates an agent with skills loaded from a directory.
func QuickAgentWithSkills(ctx context.Context, skillsDir string) (*agent.Agent, error) {
	cfg := config.Get()
	if cfg.LLM.OpenAI.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	llmClient := openai.NewClient(cfg.LLM.OpenAI.APIKey,
		openai.WithModel(cfg.LLM.OpenAI.Model),
	)
	mem := memory.NewConversationBuffer()
	tr := tools.NewRegistry()
	tr.Register(calculator.New())
	tr.Register(websearch.New())

	sr := skills.NewRegistry()
	sr.Register(builtin.NewCodeReviewSkill())
	sr.Register(builtin.NewDataAnalysisSkill())
	sr.Register(builtin.NewWebResearchSkill())
	sr.Register(builtin.NewCreativeWritingSkill())

	// Load external skills from disk
	if skillsDir != "" {
		if _, err := sr.LoadDirectory(skillsDir); err != nil {
			return nil, fmt.Errorf("load skills from %q: %w", skillsDir, err)
		}
	}

	return agent.NewAgent(
		agent.WithLLM(llmClient),
		agent.WithMemory(mem),
		agent.WithToolRegistry(tr),
		agent.WithSkillRegistry(sr),
		agent.WithSystemPrompt("You are a helpful AI assistant with access to tools and skills. Use tools when you need to search the web or perform calculations. Activate relevant skills when the task requires specialized expertise."),
		agent.WithName("AgentSDK"),
		agent.WithMaxIterations(config.GetMaxIterations()),
	)
}

// Ensure interfaces are satisfied at compile time.
var (
	_ interfaces.Tool   = (*calculator.Tool)(nil)
	_ interfaces.Tool   = (*websearch.Tool)(nil)
	_ interfaces.Skill  = (*builtin.BaseSkill)(nil)
	_ interfaces.Skill  = (*skills.FileSkill)(nil)
	_ interfaces.Memory = (*memory.ConversationBuffer)(nil)
)
