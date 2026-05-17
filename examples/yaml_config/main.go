// Example: YAML Configuration
//
// Loads agent and task configurations from YAML files and runs them.
//
// Usage:
//
//	export OPENAI_API_KEY=sk-...
//	go run examples/yaml_config/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cyberYiz/agent-sdk/pkg/agent"
	"github.com/cyberYiz/agent-sdk/pkg/interfaces"
	"github.com/cyberYiz/agent-sdk/pkg/llm/openai"
	"github.com/cyberYiz/agent-sdk/pkg/memory"
	"github.com/cyberYiz/agent-sdk/pkg/skills"
	"github.com/cyberYiz/agent-sdk/pkg/skills/builtin"
	"github.com/cyberYiz/agent-sdk/pkg/tools"
	"github.com/cyberYiz/agent-sdk/pkg/tools/calculator"
	"github.com/cyberYiz/agent-sdk/pkg/tools/websearch"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY not set")
		os.Exit(1)
	}

	// Load configs
	agentConfigs, err := agent.LoadAgentConfigsFromFile("agents.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not load agents.yaml: %v\n", err)
		os.Exit(1)
	}

	taskConfigs, err := agent.LoadTaskConfigsFromFile("tasks.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not load tasks.yaml: %v\n", err)
		os.Exit(1)
	}

	// List loaded configs
	fmt.Println("Loaded agents:")
	for name := range agentConfigs {
		fmt.Printf("  - %s (%s)\n", name, agentConfigs[name].Role)
	}
	fmt.Println()
	fmt.Println("Loaded tasks:")
	for name := range taskConfigs {
		fmt.Printf("  - %s (agent: %s)\n", name, taskConfigs[name].Agent)
	}
	fmt.Println()

	// Pick a config and run
	cfg := agentConfigs["research_assistant"]
	systemPrompt := agent.BuildSystemPromptFromConfig(cfg)
	fmt.Printf("System prompt built (%d chars)\n\n", len(systemPrompt))

	// Create LLM, memory
	llm := openai.NewClient(apiKey, openai.WithModel("gpt-4o-mini"))
	mem := memory.NewConversationBuffer()

	// Map tool names to actual tools
	toolMap := map[string]interfaces.Tool{
		"web_search": websearch.New(),
		"calculator": calculator.New(),
	}

	// Map skill names to actual skills
	skillMap := map[string]interfaces.Skill{
		"code_review":      builtin.NewCodeReviewSkill(),
		"data_analysis":    builtin.NewDataAnalysisSkill(),
		"web_research":     builtin.NewWebResearchSkill(),
		"creative_writing": builtin.NewCreativeWritingSkill(),
	}

	// Build agent options from config
	var opts []agent.Option
	opts = append(opts, agent.WithLLM(llm))
	opts = append(opts, agent.WithMemory(mem))
	opts = append(opts, agent.WithSystemPrompt(systemPrompt))
	opts = append(opts, agent.WithName(cfg.Name))

	if cfg.MaxIterations > 0 {
		opts = append(opts, agent.WithMaxIterations(cfg.MaxIterations))
	}

	// Register and activate skills
	sr := skills.NewRegistry()
	for _, sn := range cfg.Skills {
		if s, ok := skillMap[sn]; ok {
			if err := sr.Register(s); err != nil {
				fmt.Printf("Warning: register skill %s: %v\n", sn, err)
				continue
			}
			if err := sr.Activate(sn); err != nil {
				fmt.Printf("Warning: activate %s: %v\n", sn, err)
			}
		}
	}
	opts = append(opts, agent.WithSkillRegistry(sr))

	// Register tools
	tr := tools.NewRegistry()
	for _, tn := range cfg.Tools {
		if t, ok := toolMap[tn]; ok {
			tr.Register(t)
		}
	}
	opts = append(opts, agent.WithToolRegistry(tr))

	ag, err := agent.NewAgent(opts...)
	if err != nil {
		panic(err)
	}

	// Run with a research query
	ctx := context.Background()
	query := "What are the key trends in AI for 2025?"

	task := taskConfigs["deep_research"]
	desc := strings.ReplaceAll(task.Description, "{topic}", query)

	fmt.Printf("Running task 'deep_research'...\n")
	fmt.Printf("Query: %s\n\n", query)

	resp, err := ag.RunDetailed(ctx, desc)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Response ===")
	fmt.Println(resp.Content)
	fmt.Println()
	fmt.Printf("LLM calls: %d, Tool calls: %d, Time: %dms\n",
		resp.ExecutionSummary.LLMCalls,
		resp.ExecutionSummary.ToolCalls,
		resp.ExecutionSummary.ExecutionTime,
	)
}
