package ai

import "fmt"

/*
Buffer is a simple buffer, that can be used to store messages.
*/
// Buffer manages the conversation history and message handling
type Buffer struct {
	systemPrompt string
	userPrompt   string
	messages     []Message
}

type Message struct {
	Role    string
	Content string
}

/*
NewBuffer creates a new buffer.
*/
func NewBuffer(systemPrompt, userPrompt string) *Buffer {
	return &Buffer{
		systemPrompt: systemPrompt,
		userPrompt:   userPrompt,
		messages:     make([]Message, 0),
	}
}

func (b *Buffer) GetMessages() []Message {
	// Start with system and user prompts
	messages := []Message{
		{Role: "system", Content: b.systemPrompt},
		{Role: "user", Content: b.userPrompt},
	}

	// Add conversation history
	messages = append(messages, b.messages...)

	return messages
}

func (b *Buffer) AddMessage(role, content string) {
	b.messages = append(b.messages, Message{Role: role, Content: content})
}

func (b *Buffer) AddToolResult(name, result string) {
	b.messages = append(b.messages, Message{
		Role:    "tool",
		Content: fmt.Sprintf("Tool %s returned: %s", name, result),
	})
}

func (b *Buffer) Clear() {
	b.messages = b.messages[:0]
}
