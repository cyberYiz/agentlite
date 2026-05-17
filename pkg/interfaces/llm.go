// Package interfaces defines the core abstractions for the agent SDK.
package interfaces

import "context"

// LLM represents a large language model provider.
type LLM interface {
	// Generate generates text based on the provided prompt.
	Generate(ctx context.Context, prompt string, options ...GenerateOption) (string, error)

	// GenerateWithTools generates text and can use tools.
	GenerateWithTools(ctx context.Context, prompt string, tools []Tool, options ...GenerateOption) (string, error)

	// GenerateDetailed generates text and returns detailed response information including token usage.
	GenerateDetailed(ctx context.Context, prompt string, options ...GenerateOption) (*LLMResponse, error)

	// GenerateWithToolsDetailed generates text with tools and returns detailed response information.
	GenerateWithToolsDetailed(ctx context.Context, prompt string, tools []Tool, options ...GenerateOption) (*LLMResponse, error)

	// Name returns the name of the LLM provider.
	Name() string

	// SupportsStreaming returns true if this LLM supports streaming.
	SupportsStreaming() bool
}

// GenerateOption represents an option for text generation.
type GenerateOption func(options *GenerateOptions)

// GenerateOptions contains configuration for text generation.
type GenerateOptions struct {
	SystemMessage    string
	Temperature      float64
	TopP             float64
	FrequencyPenalty float64
	PresencePenalty  float64
	StopSequences    []string
	MaxTokens        int
	ResponseFormat   *ResponseFormat
	MaxIterations    int
	Memory           Memory
	StreamCh         chan<- StreamEvent
}

// ResponseFormat specifies the expected response format.
type ResponseFormat struct {
	Type       string                 `json:"type"`        // "json_object" or "text"
	JSONSchema map[string]interface{} `json:"json_schema"` // JSON schema for structured output
}

// LLMResponse represents the detailed response from an LLM generation request.
type LLMResponse struct {
	Content    string
	Usage      *TokenUsage
	Model      string
	StopReason string
	ToolCalls  []ToolCall
	Metadata   map[string]interface{}
}

// TokenUsage represents token usage information.
type TokenUsage struct {
	InputTokens    int `json:"input_tokens"`
	OutputTokens   int `json:"output_tokens"`
	TotalTokens    int `json:"total_tokens"`
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
}

// WithSystemMessage sets the system message.
func WithSystemMessage(msg string) GenerateOption {
	return func(o *GenerateOptions) {
		o.SystemMessage = msg
	}
}

// WithTemperature sets the temperature.
func WithTemperature(t float64) GenerateOption {
	return func(o *GenerateOptions) {
		o.Temperature = t
	}
}

// WithTopP sets the top_p.
func WithTopP(p float64) GenerateOption {
	return func(o *GenerateOptions) {
		o.TopP = p
	}
}

// WithMaxTokens sets the max tokens.
func WithMaxTokens(n int) GenerateOption {
	return func(o *GenerateOptions) {
		o.MaxTokens = n
	}
}

// WithResponseFormat sets the response format.
func WithResponseFormat(f ResponseFormat) GenerateOption {
	return func(o *GenerateOptions) {
		o.ResponseFormat = &f
	}
}

// WithMaxIterations sets the max tool-calling iterations.
func WithMaxIterations(n int) GenerateOption {
	return func(o *GenerateOptions) {
		o.MaxIterations = n
	}
}

// WithStreamChannel sets the channel for streaming events.
func WithStreamChannel(ch chan<- StreamEvent) GenerateOption {
	return func(o *GenerateOptions) {
		o.StreamCh = ch
	}
}
