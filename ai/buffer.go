package ai

import (
	"fmt"

	"github.com/theapemachine/amsh/ai/provider"
)

/*
Buffer is a simple buffer, that can be used to store messages.
*/
// Buffer manages the conversation history and message handling
type Buffer struct {
	messages []provider.Message
}

// Message represents a conversation message
type Message struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

/*
NewBuffer creates a new buffer.
*/
func NewBuffer() *Buffer {
	return &Buffer{
		messages: make([]provider.Message, 0),
	}
}

// GetMessages returns all messages in the conversation
func (b *Buffer) GetMessages() []provider.Message {
	messages := make([]provider.Message, 0)
	messages = append(messages, b.messages...)

	return messages
}

func (b *Buffer) AddMessage(role, content string) *Buffer {
	b.messages = append(b.messages, provider.Message{Role: role, Content: content})
	return b
}

func (b *Buffer) AddToolResult(name, result string) {
	b.messages = append(b.messages, provider.Message{
		Role:    "tool",
		Content: fmt.Sprintf("Tool %s returned: %s", name, result),
	})
}

func (b *Buffer) Clear() {
	b.messages = b.messages[:0]
}
