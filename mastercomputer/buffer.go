package mastercomputer

import (
	"sync"

	"github.com/openai/openai-go"
)

// ConversationBuffer is a scoped context that holds the messages for a specific worker.
type ConversationBuffer struct {
	messages []openai.ChatCompletionMessageParamUnion
	tag      string
	mux      sync.Mutex
}

// NewConversationBuffer creates a new buffer for the given worker.
func NewConversationBuffer(tag string) *ConversationBuffer {
	return &ConversationBuffer{
		messages: []openai.ChatCompletionMessageParamUnion{},
		tag:      tag,
	}
}

// AddMessage adds a message to the worker's conversation buffer with tagging.
func (buffer *ConversationBuffer) AddMessage(message openai.ChatCompletionMessageParamUnion) {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()
	buffer.messages = append(buffer.messages, message)
}

// GetScopedMessages returns a copy of all the messages from this buffer.
func (buffer *ConversationBuffer) GetScopedMessages() []openai.ChatCompletionMessageParamUnion {
	buffer.mux.Lock()
	defer buffer.mux.Unlock()
	return append([]openai.ChatCompletionMessageParamUnion{}, buffer.messages...)
}
