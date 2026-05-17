// Package agent implements the core agent orchestration logic.
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cyberYiz/agentlite/pkg/interfaces"
	"github.com/cyberYiz/agentlite/pkg/skills"
)

// Agent orchestrates LLM calls, tool execution, memory, and skills.
type Agent struct {
	llm          interfaces.LLM
	memory       interfaces.Memory
	toolRegistry interfaces.ToolRegistry
	skillReg     *skills.Registry

	name         string
	systemPrompt string
	maxIter      int
	metadata     map[string]interface{}
}

// Option is a functional option for configuring an Agent.
type Option func(*Agent)

// WithLLM sets the LLM provider.
func WithLLM(llm interfaces.LLM) Option {
	return func(a *Agent) {
		a.llm = llm
	}
}

// WithMemory sets the memory store.
func WithMemory(mem interfaces.Memory) Option {
	return func(a *Agent) {
		a.memory = mem
	}
}

// WithTools registers individual tools.
func WithTools(tools ...interfaces.Tool) Option {
	return func(a *Agent) {
		for _, t := range tools {
			a.toolRegistry.Register(t)
		}
	}
}

// WithToolRegistry sets a pre-built tool registry.
func WithToolRegistry(reg interfaces.ToolRegistry) Option {
	return func(a *Agent) {
		a.toolRegistry = reg
	}
}

// WithSkillRegistry sets the skill registry.
func WithSkillRegistry(reg *skills.Registry) Option {
	return func(a *Agent) {
		a.skillReg = reg
	}
}

// WithSystemPrompt sets the base system prompt.
func WithSystemPrompt(prompt string) Option {
	return func(a *Agent) {
		a.systemPrompt = prompt
	}
}

// WithName sets the agent's name.
func WithName(name string) Option {
	return func(a *Agent) {
		a.name = name
	}
}

// WithMaxIterations sets the maximum number of tool-calling iterations.
func WithMaxIterations(n int) Option {
	return func(a *Agent) {
		a.maxIter = n
	}
}

// WithMetadata sets custom metadata on the agent.
func WithMetadata(m map[string]interface{}) Option {
	return func(a *Agent) {
		a.metadata = m
	}
}

// NewAgent creates a new Agent with the given options.
func NewAgent(opts ...Option) (*Agent, error) {
	a := &Agent{
		toolRegistry: newToolRegistry(),
		skillReg:     skills.NewRegistry(),
		name:         "agent",
		maxIter:      15,
		metadata:     make(map[string]interface{}),
	}

	for _, o := range opts {
		o(a)
	}

	if a.llm == nil {
		return nil, fmt.Errorf("agent: LLM is required")
	}
	if a.memory == nil {
		return nil, fmt.Errorf("agent: Memory is required")
	}

	return a, nil
}

// Name returns the agent's name.
func (a *Agent) Name() string { return a.name }

// SystemPrompt returns the combined system prompt (base + active skills).
func (a *Agent) SystemPrompt() string {
	base := a.systemPrompt
	if aug := a.skillReg.GetCombinedSystemPrompt(); aug != "" {
		base = strings.TrimSpace(base) + "\n\n" + aug
	}
	return base
}

// Tools returns all available tools (from registry + active skills).
func (a *Agent) Tools() []interfaces.Tool {
	tools := a.toolRegistry.List()
	tools = append(tools, a.skillReg.GetAllActiveTools()...)
	return tools
}

// GetMemory returns the agent's memory.
func (a *Agent) GetMemory() interfaces.Memory { return a.memory }

// SkillRegistry returns the skill registry.
func (a *Agent) SkillRegistry() *skills.Registry { return a.skillReg }

// ToolRegistry returns the tool registry.
func (a *Agent) ToolRegistry() interfaces.ToolRegistry { return a.toolRegistry }

// Run executes a prompt and returns the final response.
func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
	resp, err := a.RunDetailed(ctx, prompt)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// RunDetailed executes a prompt and returns detailed results including
// execution summary, token usage, and tool calls.
func (a *Agent) RunDetailed(ctx context.Context, prompt string) (*interfaces.AgentResponse, error) {
	// Fast-path: check context before any work
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("agent context cancelled before run: %w", err)
	}

	startTime := time.Now()

	// Run pre-run hooks from active skills
	for _, s := range a.skillReg.GetActiveSkills() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("agent context cancelled during pre-run: %w", err)
		}
		var err error
		prompt, err = s.PreRun(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("skill %s pre-run: %w", s.Name(), err)
		}
	}

	// Add user message to memory
	if err := a.memory.AddMessage(ctx, interfaces.Message{
		Role:    interfaces.MessageRoleUser,
		Content: prompt,
	}); err != nil {
		return nil, fmt.Errorf("add user message: %w", err)
	}

	tools := a.Tools()
	systemPrompt := a.SystemPrompt()

	var llmCalls int
	var toolCalls int
	usedTools := make(map[string]bool)
	var totalUsage *interfaces.TokenUsage

	// Tool-calling loop
	for iteration := 0; iteration < a.maxIter; iteration++ {
		// Check context before each iteration (between LLM calls or tool executions)
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("agent context cancelled at iteration %d: %w", iteration, err)
		}

		llmCalls++

		// Build messages from memory
		msgs, err := a.memory.GetMessages(ctx)
		if err != nil {
			return nil, fmt.Errorf("get messages: %w", err)
		}

		// Build the prompt with chat history for the LLM
		// We use system prompt + memory messages
		var opts []interfaces.GenerateOption
		opts = append(opts, interfaces.WithSystemMessage(systemPrompt))
		opts = append(opts, interfaces.WithMaxIterations(a.maxIter-iteration))

		// If we have conversation history, include it in the prompt
		fullPrompt := prompt
		if len(msgs) > 0 {
			fullPrompt = buildChatPrompt(msgs)
		}

		var resp *interfaces.LLMResponse
		if len(tools) > 0 {
			resp, err = a.llm.GenerateWithToolsDetailed(ctx, fullPrompt, tools, opts...)
		} else {
			resp, err = a.llm.GenerateDetailed(ctx, fullPrompt, opts...)
		}
		if err != nil {
			return nil, fmt.Errorf("llm call: %w", err)
		}

		// Accumulate usage
		if resp.Usage != nil {
			if totalUsage == nil {
				totalUsage = resp.Usage
			} else {
				totalUsage.InputTokens += resp.Usage.InputTokens
				totalUsage.OutputTokens += resp.Usage.OutputTokens
				totalUsage.TotalTokens += resp.Usage.TotalTokens
			}
		}

		// Check for tool calls
		if len(resp.ToolCalls) > 0 {
			// Add assistant message with tool calls to memory
			msg := interfaces.Message{
				Role:      interfaces.MessageRoleAssistant,
				Content:   resp.Content,
				ToolCalls: make([]interfaces.ToolCall, len(resp.ToolCalls)),
			}
			for i, tc := range resp.ToolCalls {
				msg.ToolCalls[i] = tc
				usedTools[tc.Name] = true
			}
			if err := a.memory.AddMessage(ctx, msg); err != nil {
				return nil, fmt.Errorf("add assistant message: %w", err)
			}

			// Execute each tool call
			for _, tc := range resp.ToolCalls {
				// Check context between tool executions
				if err := ctx.Err(); err != nil {
					return nil, fmt.Errorf("agent context cancelled during tool execution (%s): %w", tc.Name, err)
				}
				toolCalls++
				toolResult, execErr := a.executeTool(ctx, tc)
				toolMsg := interfaces.Message{
					Role:       interfaces.MessageRoleTool,
					Content:    toolResult,
					ToolCallID: tc.ID,
				}
				if execErr != nil {
					toolMsg.Content = fmt.Sprintf("Error: %v", execErr)
				}
				if err := a.memory.AddMessage(ctx, toolMsg); err != nil {
					return nil, fmt.Errorf("add tool result: %w", err)
				}
			}

			// Continue loop - LLM will process tool results
			continue
		}

		// No tool calls — this is the final response
		content := resp.Content
		if content == "" {
			content = "I processed your request but have no additional output."
		}

		// Add assistant response to memory
		if err := a.memory.AddMessage(ctx, interfaces.Message{
			Role:    interfaces.MessageRoleAssistant,
			Content: content,
		}); err != nil {
			return nil, fmt.Errorf("add assistant message: %w", err)
		}

		// Run post-run hooks
		for _, s := range a.skillReg.GetActiveSkills() {
			// Context cancellation during post-run is non-fatal; we still return the content
			if ctx.Err() != nil {
				break
			}
			var postErr error
			content, postErr = s.PostRun(ctx, content)
			if postErr != nil {
				// Log but don't fail
				_ = postErr
			}
		}

		// Build used tools list
		var usedToolsList []string
		for t := range usedTools {
			usedToolsList = append(usedToolsList, t)
		}

		return &interfaces.AgentResponse{
			Content:   content,
			AgentName: a.name,
			Model:     resp.Model,
			Usage:     totalUsage,
			ExecutionSummary: &interfaces.ExecutionSummary{
				LLMCalls:      llmCalls,
				ToolCalls:     toolCalls,
				ExecutionTime: time.Since(startTime).Milliseconds(),
				UsedTools:     usedToolsList,
			},
		}, nil
	}

	return nil, fmt.Errorf("exceeded max iterations (%d) without final response", a.maxIter)
}

func (a *Agent) executeTool(ctx context.Context, tc interfaces.ToolCall) (string, error) {
	// Fast-path context check before tool lookup + execution
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context cancelled before executing tool %q: %w", tc.Name, err)
	}

	t, ok := a.toolRegistry.Get(tc.Name)
	if !ok {
		// Try skills tools
		for _, s := range a.skillReg.GetActiveSkills() {
			for _, st := range s.Tools() {
				if st.Name() == tc.Name {
					t = st
					break
				}
			}
			if t != nil {
				break
			}
		}
	}

	if t == nil {
		return "", fmt.Errorf("tool %q not found", tc.Name)
	}

	return t.Execute(ctx, tc.Arguments)
}

// buildChatPrompt converts memory messages to a formatted prompt string.
func buildChatPrompt(msgs []interfaces.Message) string {
	var sb strings.Builder
	for _, m := range msgs {
		switch m.Role {
		case interfaces.MessageRoleUser:
			sb.WriteString(fmt.Sprintf("User: %s\n", m.Content))
		case interfaces.MessageRoleAssistant:
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					sb.WriteString(fmt.Sprintf("Assistant: [Tool Call: %s(%s)]\n", tc.Name, tc.Arguments))
				}
			} else {
				sb.WriteString(fmt.Sprintf("Assistant: %s\n", m.Content))
			}
		case interfaces.MessageRoleTool:
			sb.WriteString(fmt.Sprintf("Tool Result (id=%s): %s\n", m.ToolCallID, m.Content))
		case interfaces.MessageRoleSystem:
			sb.WriteString(fmt.Sprintf("System: %s\n", m.Content))
		}
	}
	return sb.String()
}

// newToolRegistry creates a basic in-memory tool registry.
func newToolRegistry() interfaces.ToolRegistry {
	return &simpleToolRegistry{
		tools: make(map[string]interfaces.Tool),
	}
}

// simpleToolRegistry is a basic implementation of ToolRegistry.
type simpleToolRegistry struct {
	tools map[string]interfaces.Tool
}

func (r *simpleToolRegistry) Register(tool interfaces.Tool) {
	r.tools[tool.Name()] = tool
}
func (r *simpleToolRegistry) Get(name string) (interfaces.Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}
func (r *simpleToolRegistry) List() []interfaces.Tool {
	out := make([]interfaces.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}
