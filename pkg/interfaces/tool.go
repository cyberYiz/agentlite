package interfaces

import "context"

// Tool represents a tool that can be used by an agent.
type Tool interface {
	// Name returns the unique name of the tool.
	Name() string

	// Description returns a human-readable description of what the tool does.
	Description() string

	// Parameters returns the JSON Schema for the tool's parameters.
	Parameters() map[string]interface{}

	// Execute runs the tool with the given JSON-encoded arguments.
	Execute(ctx context.Context, args string) (string, error)
}

// ToolWithDisplayName is an optional interface for tools to provide a display name.
type ToolWithDisplayName interface {
	DisplayName() string
}

// ToolRegistry is a thread-safe registry of tools.
type ToolRegistry interface {
	Register(tool Tool)
	Get(name string) (Tool, bool)
	List() []Tool
}
