package interfaces

import "context"

// Agent orchestrates LLM calls, tool execution, and memory management.
type Agent interface {
	// Run executes a single prompt and returns the final response.
	Run(ctx context.Context, prompt string) (string, error)

	// RunDetailed executes a prompt and returns detailed results.
	RunDetailed(ctx context.Context, prompt string) (*AgentResponse, error)

	// Name returns the agent's name.
	Name() string

	// SystemPrompt returns the agent's system prompt.
	SystemPrompt() string

	// Tools returns the list of available tools.
	Tools() []Tool

	// GetMemory returns the agent's memory.
	GetMemory() Memory
}

// AgentResponse represents the result of an agent run.
type AgentResponse struct {
	Content          string
	AgentName        string
	Model            string
	Usage            *TokenUsage
	ExecutionSummary *ExecutionSummary
	Metadata         map[string]interface{}
}

// ExecutionSummary summarizes the agent's execution.
type ExecutionSummary struct {
	LLMCalls      int      `json:"llm_calls"`
	ToolCalls     int      `json:"tool_calls"`
	ExecutionTime int64    `json:"execution_time_ms"`
	UsedTools     []string `json:"used_tools"`
}
