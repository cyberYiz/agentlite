// Package memory provides in-memory conversation storage.
package memory

import (
	"context"
	"sync"

	"github.com/cyberYiz/agentlite/pkg/interfaces"
)

// ConversationBuffer implements interfaces.Memory with an in-memory message store.
type ConversationBuffer struct {
	mu       sync.RWMutex
	messages []interfaces.Message
}

// NewConversationBuffer creates a new in-memory conversation buffer.
func NewConversationBuffer() *ConversationBuffer {
	return &ConversationBuffer{
		messages: make([]interfaces.Message, 0),
	}
}

// AddMessage appends a message to the buffer.
func (b *ConversationBuffer) AddMessage(ctx context.Context, msg interfaces.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = append(b.messages, msg)
	return nil
}

// GetMessages returns all messages, optionally filtered.
func (b *ConversationBuffer) GetMessages(ctx context.Context, opts ...interfaces.GetMessagesOption) ([]interfaces.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	b.mu.RLock()
	defer b.mu.RUnlock()

	options := &interfaces.GetMessagesOptions{}
	for _, o := range opts {
		o(options)
	}

	var result []interfaces.Message
	roleSet := make(map[interfaces.MessageRole]bool)
	for _, r := range options.Roles {
		roleSet[r] = true
	}

	for _, msg := range b.messages {
		if len(roleSet) > 0 && !roleSet[msg.Role] {
			continue
		}
		result = append(result, msg)
	}

	// Apply limit from end
	if options.Limit > 0 && len(result) > options.Limit {
		result = result[len(result)-options.Limit:]
	}

	return result, nil
}

// Clear removes all messages.
func (b *ConversationBuffer) Clear(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = make([]interfaces.Message, 0)
	return nil
}

// ConversationBufferStats returns total message count.
func (b *ConversationBuffer) Stats() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.messages)
}
