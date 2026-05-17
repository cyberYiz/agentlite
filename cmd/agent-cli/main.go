// Agent CLI — command-line interface for the Agent SDK.
//
// Usage:
//
//	agent-cli run "What is the capital of France?"
//	agent-cli chat
//	agent-cli task --agent-config=agents.yaml --task-config=tasks.yaml --task=deep_research --topic="AI"
//
// Environment variables:
//
//	OPENAI_API_KEY     (required)
//	OPENAI_MODEL       (default: gpt-4o-mini)
//	LOG_LEVEL          (default: info)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cyberYiz/agent-sdk/pkg/agent"
	"github.com/cyberYiz/agent-sdk/pkg/config"
	"github.com/cyberYiz/agent-sdk/pkg/llm/openai"
	"github.com/cyberYiz/agent-sdk/pkg/logging"
	"github.com/cyberYiz/agent-sdk/pkg/memory"
	"github.com/cyberYiz/agent-sdk/pkg/skills"
	"github.com/cyberYiz/agent-sdk/pkg/skills/builtin"
	"github.com/cyberYiz/agent-sdk/pkg/tools"
	"github.com/cyberYiz/agent-sdk/pkg/tools/calculator"
	"github.com/cyberYiz/agent-sdk/pkg/tools/websearch"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "run":
		runCommand()
	case "chat":
		chatCommand()
	case "task":
		taskCommand()
	case "list":
		listCommand()
	case "version":
		fmt.Println("agent-cli v0.1.0")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`Agent CLI v0.1.0

Usage:
  agent-cli run "your prompt"       Run a single query
  agent-cli chat                     Start interactive chat
  agent-cli task [options]           Execute a YAML-defined task
  agent-cli list [providers|tools|skills]  List available resources
  agent-cli version                  Show version

Environment:
  OPENAI_API_KEY     (required)  Your OpenAI API key
  OPENAI_MODEL                   Model name (default: gpt-4o-mini)
  LOG_LEVEL                     debug, info, warn, error (default: info)
`)
}

func runCommand() {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	model := fs.String("model", "", "Model to use")
	prompt := fs.String("prompt", "", "Prompt (alias for positional arg)")
	fs.Parse(os.Args[2:])

	query := *prompt
	if query == "" && fs.NArg() > 0 {
		query = strings.Join(fs.Args(), " ")
	}
	if query == "" {
		fmt.Fprintln(os.Stderr, "Error: no prompt provided")
		os.Exit(1)
	}

	cfg := config.Get()
	if cfg.LLM.OpenAI.APIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY not set")
		os.Exit(1)
	}

	if *model != "" {
		cfg.LLM.OpenAI.Model = *model
	}

	logger := logging.New()
	logger.Info("Starting agent...")

	ag := buildAgent(cfg)

	ctx := context.Background()
	resp, err := ag.RunDetailed(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(resp.Content)
	fmt.Println()
	logger.Info("Done. LLM calls: %d, Tools: %d, Time: %dms",
		resp.ExecutionSummary.LLMCalls,
		resp.ExecutionSummary.ToolCalls,
		resp.ExecutionSummary.ExecutionTime,
	)
}

func chatCommand() {
	fmt.Println("Chat mode not yet implemented. Use 'run' for single queries.")
}

func taskCommand() {
	fs := flag.NewFlagSet("task", flag.ExitOnError)
	agentConfig := fs.String("agent-config", "agents.yaml", "Agent YAML config file")
	taskConfig := fs.String("task-config", "tasks.yaml", "Task YAML config file")
	taskName := fs.String("task", "", "Task name to execute")
	topic := fs.String("topic", "", "Topic variable")
	fs.Parse(os.Args[2:])

	if *taskName == "" {
		fmt.Fprintln(os.Stderr, "Error: --task is required")
		os.Exit(1)
	}

	cfg := config.Get()
	if cfg.LLM.OpenAI.APIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY not set")
		os.Exit(1)
	}

	agentConfigs, err := agent.LoadAgentConfigsFromFile(*agentConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading agent config: %v\n", err)
		os.Exit(1)
	}

	taskConfigs, err := agent.LoadTaskConfigsFromFile(*taskConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading task config: %v\n", err)
		os.Exit(1)
	}

	task, ok := taskConfigs[*taskName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: task %q not found\n", *taskName)
		os.Exit(1)
	}

	ac, ok := agentConfigs[task.Agent]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: agent %q not found\n", task.Agent)
		os.Exit(1)
	}

	systemPrompt := agent.BuildSystemPromptFromConfig(ac)
	desc := strings.ReplaceAll(task.Description, "{topic}", *topic)
	desc = strings.ReplaceAll(desc, "{data}", *topic)

	fmt.Printf("Task: %s\n", *taskName)
	fmt.Printf("Agent: %s (%s)\n", task.Agent, ac.Role)
	fmt.Println("---")

	ag := buildAgentFromConfig(cfg, ac, systemPrompt)

	ctx := context.Background()
	resp, err := ag.RunDetailed(ctx, desc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(resp.Content)
}

func listCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: agent-cli list [providers|tools|skills]")
		return
	}

	switch os.Args[2] {
	case "providers":
		fmt.Println("Available LLM providers: openai")
	case "tools":
		fmt.Println("Built-in tools:")
		fmt.Println("  - calculator: Basic arithmetic")
		fmt.Println("  - web_search: Web search via DuckDuckGo")
	case "skills":
		fmt.Println("Built-in skills:")
		fmt.Println("  - code_review (code): Expert code review and analysis")
		fmt.Println("  - data_analysis (data): Data analysis and visualization")
		fmt.Println("  - web_research (web): Web research and information gathering")
		fmt.Println("  - creative_writing (content): Creative writing and content generation")
	}
}

func buildAgent(cfg *config.Config) *agent.Agent {
	llm := openai.NewClient(cfg.LLM.OpenAI.APIKey,
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

	ag, err := agent.NewAgent(
		agent.WithLLM(llm),
		agent.WithMemory(mem),
		agent.WithToolRegistry(tr),
		agent.WithSkillRegistry(sr),
		agent.WithSystemPrompt("You are a helpful AI assistant. Use tools when needed."),
		agent.WithName("AgentCLI"),
		agent.WithMaxIterations(config.GetMaxIterations()),
	)
	if err != nil {
		panic(err)
	}
	return ag
}

func buildAgentFromConfig(cfg *config.Config, ac agent.AgentConfig, systemPrompt string) *agent.Agent {
	llm := openai.NewClient(cfg.LLM.OpenAI.APIKey,
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

	for _, sn := range ac.Skills {
		if err := sr.Activate(sn); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot activate skill %s: %v\n", sn, err)
		}
	}

	maxIter := ac.MaxIterations
	if maxIter <= 0 {
		maxIter = config.GetMaxIterations()
	}

	ag, err := agent.NewAgent(
		agent.WithLLM(llm),
		agent.WithMemory(mem),
		agent.WithToolRegistry(tr),
		agent.WithSkillRegistry(sr),
		agent.WithSystemPrompt(systemPrompt),
		agent.WithName(ac.Name),
		agent.WithMaxIterations(maxIter),
	)
	if err != nil {
		panic(err)
	}
	return ag
}
