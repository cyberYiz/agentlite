// Package tools provides plug-and-play tools for agents.
package tools

import (
	"sync"

	"github.com/cyberYiz/agent-sdk/pkg/interfaces"
)

// Registry implements interfaces.ToolRegistry with thread-safe operations.
type Registry struct {
	tools map[string]interfaces.Tool
	mu    sync.RWMutex
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]interfaces.Tool),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(tool interfaces.Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (interfaces.Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// List returns all registered tools.
func (r *Registry) List() []interfaces.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]interfaces.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}
