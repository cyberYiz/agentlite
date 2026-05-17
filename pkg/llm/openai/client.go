// Package openai implements the LLM interface for OpenAI's API.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cyberYiz/agentlite/pkg/interfaces"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
	defaultModel   = "gpt-4o-mini"
)

// Client implements interfaces.LLM for OpenAI.
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewClient creates a new OpenAI LLM client.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		model:      defaultModel,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Option configures the OpenAI client.
type Option func(*Client)

// WithModel sets the model name.
func WithModel(model string) Option {
	return func(c *Client) {
		c.model = model
	}
}

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// Name returns the provider name.
func (c *Client) Name() string { return "openai" }

// SupportsStreaming returns true.
func (c *Client) SupportsStreaming() bool { return true }

// Generate implements LLM.Generate.
func (c *Client) Generate(ctx context.Context, prompt string, opts ...interfaces.GenerateOption) (string, error) {
	resp, err := c.GenerateDetailed(ctx, prompt, opts...)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// GenerateWithTools implements LLM.GenerateWithTools.
func (c *Client) GenerateWithTools(ctx context.Context, prompt string, tools []interfaces.Tool, opts ...interfaces.GenerateOption) (string, error) {
	resp, err := c.GenerateWithToolsDetailed(ctx, prompt, tools, opts...)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// GenerateDetailed implements LLM.GenerateDetailed.
func (c *Client) GenerateDetailed(ctx context.Context, prompt string, opts ...interfaces.GenerateOption) (*interfaces.LLMResponse, error) {
	return c.chat(ctx, prompt, nil, opts...)
}

// GenerateWithToolsDetailed implements LLM.GenerateWithToolsDetailed.
func (c *Client) GenerateWithToolsDetailed(ctx context.Context, prompt string, tools []interfaces.Tool, opts ...interfaces.GenerateOption) (*interfaces.LLMResponse, error) {
	return c.chat(ctx, prompt, tools, opts...)
}

func (c *Client) chat(ctx context.Context, prompt string, tools []interfaces.Tool, opts ...interfaces.GenerateOption) (*interfaces.LLMResponse, error) {
	// Fast-path context check before any work
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("llm request cancelled before start: %w", err)
	}

	options := &interfaces.GenerateOptions{}
	for _, o := range opts {
		o(options)
	}

	req := chatRequest{
		Model:    c.model,
		Messages: []chatMessage{},
	}

	if options.Temperature > 0 {
		req.Temperature = &options.Temperature
	}
	if options.TopP > 0 {
		req.TopP = &options.TopP
	}
	if options.MaxTokens > 0 {
		req.MaxTokens = options.MaxTokens
	}
	if options.ResponseFormat != nil {
		req.ResponseFormat = &responseFormat{
			Type:       options.ResponseFormat.Type,
			JSONSchema: options.ResponseFormat.JSONSchema,
		}
	}

	// Add system message if present
	if options.SystemMessage != "" {
		req.Messages = append(req.Messages, chatMessage{
			Role:    "system",
			Content: options.SystemMessage,
		})
	}

	// Add user prompt
	req.Messages = append(req.Messages, chatMessage{
		Role:    "user",
		Content: prompt,
	})

	// Convert tools to OpenAI format
	if len(tools) > 0 {
		req.Tools = make([]toolDef, len(tools))
		for i, t := range tools {
			req.Tools[i] = toolDef{
				Type: "function",
				Function: toolFunction{
					Name:        t.Name(),
					Description: t.Description(),
					Parameters:  t.Parameters(),
				},
			}
		}
		req.ToolChoice = "auto"
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := chatResp.Choices[0]
	result := &interfaces.LLMResponse{
		Model:      chatResp.Model,
		StopReason: choice.FinishReason,
	}

	if chatResp.Usage != nil {
		result.Usage = &interfaces.TokenUsage{
			InputTokens:  chatResp.Usage.PromptTokens,
			OutputTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:  chatResp.Usage.TotalTokens,
		}
	}

	// Handle tool calls
	if len(choice.Message.ToolCalls) > 0 {
		for _, tc := range choice.Message.ToolCalls {
			result.ToolCalls = append(result.ToolCalls, interfaces.ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			})
		}
		result.Content = "" // tool calls mean no text content
	} else {
		result.Content = choice.Message.Content
	}

	return result, nil
}

// --- OpenAI API types ---

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    *float64        `json:"temperature,omitempty"`
	TopP           *float64        `json:"top_p,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Tools          []toolDef       `json:"tools,omitempty"`
	ToolChoice     string          `json:"tool_choice,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
	Stream         bool            `json:"stream,omitempty"`
}

type chatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function toolCallFunc `json:"function"`
}

type toolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type toolDef struct {
	Type     string       `json:"type"`
	Function toolFunction `json:"function"`
}

type toolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

type responseFormat struct {
	Type       string                 `json:"type"`
	JSONSchema map[string]interface{} `json:"json_schema,omitempty"`
}

type chatResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
	Usage   *usage   `json:"usage,omitempty"`
}

type choice struct {
	Index        int         `json:"index"`
	Message      chatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
