package ai

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/provider"
)

/*
Buffer is a simple buffer, that can be used to store messages.
*/
// Buffer manages the conversation history and message handling
type Buffer struct {
	systemPrompt string
	userPrompt   string
	messages     []provider.Message
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
func NewBuffer(systemPrompt, userPrompt string) *Buffer {
	return &Buffer{
		systemPrompt: systemPrompt,
		userPrompt:   userPrompt,
		messages:     make([]provider.Message, 0),
	}
}

// GetMessages returns all messages in the conversation
func (b *Buffer) GetMessages() []provider.Message {
	// Calculate initial capacity
	capacity := len(b.messages)
	if b.systemPrompt != "" {
		capacity++
	}
	if b.userPrompt != "" {
		capacity++
	}

	messages := make([]provider.Message, 0, capacity)

	// Add system prompt first if it exists
	if b.systemPrompt != "" {
		messages = append(messages, provider.Message{
			Role:    "system",
			Content: b.systemPrompt,
		})
	}

	// Add user prompt next if it exists
	if b.userPrompt != "" {
		messages = append(messages, provider.Message{
			Role:    "user",
			Content: b.userPrompt,
		})
	}

	// Add conversation history
	messages = append(messages, b.messages...)

	log.Info("Messages", "messages", messages)

	return messages
}

func (b *Buffer) AddMessage(role, content string) {
	b.messages = append(b.messages, provider.Message{Role: role, Content: content})
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

func (b *Buffer) GetSystemPrompt() string {
	return b.systemPrompt
}
