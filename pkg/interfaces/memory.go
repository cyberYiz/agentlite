package interfaces

import "context"

// MessageRole represents the role of a message sender.
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
	MessageRoleTool      MessageRole = "tool"
)

// Message represents a single message in a conversation.
type Message struct {
	Role       MessageRole            `json:"role"`
	Content    string                 `json:"content"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall             `json:"tool_calls,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a tool call made by the assistant.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Memory stores conversation history for an agent.
type Memory interface {
	// AddMessage appends a message to the conversation.
	AddMessage(ctx context.Context, msg Message) error

	// GetMessages retrieves messages, optionally filtered.
	GetMessages(ctx context.Context, opts ...GetMessagesOption) ([]Message, error)

	// Clear clears all messages.
	Clear(ctx context.Context) error
}

// GetMessagesOptions configures message retrieval.
type GetMessagesOptions struct {
	Limit int
	Roles []MessageRole
}

// GetMessagesOption is a functional option for GetMessages.
type GetMessagesOption func(*GetMessagesOptions)

// WithLimit sets the maximum number of messages to retrieve.
func WithLimit(limit int) GetMessagesOption {
	return func(o *GetMessagesOptions) {
		o.Limit = limit
	}
}

// WithRoles filters messages by role.
func WithRoles(roles ...MessageRole) GetMessagesOption {
	return func(o *GetMessagesOptions) {
		o.Roles = roles
	}
}

// ConversationIDKey is the context key for conversation IDs.
const ConversationIDKey = "conversation_id"
