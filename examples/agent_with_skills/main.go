// Example: Agent with Skills
//
// Demonstrates the Skills system: register skills, activate one,
// and see how it augments the agent's behavior.
//
// Usage:
//
//	export OPENAI_API_KEY=sk-...
//	go run examples/agent_with_skills/main.go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/user/agent-sdk/pkg/agent"
	"github.com/user/agent-sdk/pkg/llm/openai"
	"github.com/user/agent-sdk/pkg/memory"
	"github.com/user/agent-sdk/pkg/skills"
	"github.com/user/agent-sdk/pkg/skills/builtin"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY not set")
		os.Exit(1)
	}

	llm := openai.NewClient(apiKey, openai.WithModel("gpt-4o-mini"))
	mem := memory.NewConversationBuffer()

	// Create skill registry and register skills
	sr := skills.NewRegistry()
	sr.Register(builtin.NewCodeReviewSkill())
	sr.Register(builtin.NewDataAnalysisSkill())
	sr.Register(builtin.NewWebResearchSkill())
	sr.Register(builtin.NewCreativeWritingSkill())

	// Activate the web research skill
	if err := sr.Activate("web_research"); err != nil {
		panic(err)
	}

	ag, err := agent.NewAgent(
		agent.WithLLM(llm),
		agent.WithMemory(mem),
		agent.WithSkillRegistry(sr),
		agent.WithSystemPrompt("You are a helpful assistant."),
		agent.WithName("SkillAgent"),
		agent.WithMaxIterations(10),
	)
	if err != nil {
		panic(err)
	}

	// Show the combined system prompt
	fmt.Println("=== Combined System Prompt ===")
	fmt.Println(ag.SystemPrompt())
	fmt.Println("==============================")
	fmt.Println()

	ctx := context.Background()
	resp, err := ag.Run(ctx,
		"Research the current state of quantum computing and provide a 3-bullet summary.",
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("=== Agent Response ===")
	fmt.Println(resp)
}
