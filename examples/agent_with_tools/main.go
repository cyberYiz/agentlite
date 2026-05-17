// Example: Agent with Tools
//
// Shows how to equip an agent with tools (calculator, web search).
// The agent will automatically use tools when the LLM decides to.
//
// Usage:
//
//	export OPENAI_API_KEY=sk-...
//	go run examples/agent_with_tools/main.go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/user/agent-sdk/pkg/agent"
	"github.com/user/agent-sdk/pkg/llm/openai"
	"github.com/user/agent-sdk/pkg/memory"
	"github.com/user/agent-sdk/pkg/tools"
	"github.com/user/agent-sdk/pkg/tools/calculator"
	"github.com/user/agent-sdk/pkg/tools/websearch"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY not set")
		os.Exit(1)
	}

	llm := openai.NewClient(apiKey, openai.WithModel("gpt-4o-mini"))
	mem := memory.NewConversationBuffer()

	// Build tool registry
	registry := tools.NewRegistry()
	registry.Register(calculator.New())
	registry.Register(websearch.New())

	ag, err := agent.NewAgent(
		agent.WithLLM(llm),
		agent.WithMemory(mem),
		agent.WithToolRegistry(registry),
		agent.WithSystemPrompt(
			"You are a helpful assistant. Use tools when needed. "+
				"Use calculator for math, web_search for current information.",
		),
		agent.WithName("ToolAgent"),
		agent.WithMaxIterations(10),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Run with a query that needs calculation
	fmt.Println("--- Math query ---")
	resp, err := ag.Run(ctx, "What is 156 * 42 + 789?")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
	fmt.Println()

	// Run a second query that needs search
	fmt.Println("--- Search query ---")
	resp, err = ag.Run(ctx, "Who won the most recent Super Bowl?")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}
