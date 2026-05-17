package interfaces

// StreamEventType categorizes streaming events.
type StreamEventType string

const (
	StreamEventText     StreamEventType = "text"
	StreamEventToolCall StreamEventType = "tool_call"
	StreamEventToolResult StreamEventType = "tool_result"
	StreamEventError    StreamEventType = "error"
	StreamEventDone     StreamEventType = "done"
)

// StreamEvent represents a single event in an agent's streaming output.
type StreamEvent struct {
	Type      StreamEventType `json:"type"`
	Content   string          `json:"content,omitempty"`
	ToolName  string          `json:"tool_name,omitempty"`
	ToolArgs  string          `json:"tool_args,omitempty"`
	ToolID    string          `json:"tool_id,omitempty"`
	Error     error           `json:"-"`
	ErrorMsg  string          `json:"error,omitempty"`
	Usage     *TokenUsage     `json:"usage,omitempty"`
	Final     bool            `json:"final,omitempty"`
}
