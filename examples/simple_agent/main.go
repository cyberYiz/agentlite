// Example: Simple Agent
//
// Demonstrates the most basic agent usage: create an agent with a system
// prompt and run a single query.
//
// Usage:
//
//	export OPENAI_API_KEY=sk-...
//	go run examples/simple_agent/main.go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cyberYiz/agent-sdk/pkg/agent"
	"github.com/cyberYiz/agent-sdk/pkg/llm/openai"
	"github.com/cyberYiz/agent-sdk/pkg/memory"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY not set")
		os.Exit(1)
	}

	// Create LLM client
	llm := openai.NewClient(apiKey, openai.WithModel("gpt-4o-mini"))

	// Create agent
	ag, err := agent.NewAgent(
		agent.WithLLM(llm),
		agent.WithMemory(memory.NewConversationBuffer()),
		agent.WithSystemPrompt("You are a helpful AI assistant. Answer concisely."),
		agent.WithName("SimpleAgent"),
	)
	if err != nil {
		panic(err)
	}

	// Run
	ctx := context.Background()
	response, err := ag.Run(ctx, "What is the capital of France?")
	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
