package mastercomputer

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/pkoukk/tiktoken-go"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/utils"
)

type Buffer struct {
	messages         []provider.Message
	maxContextTokens int
}

func NewBuffer() *Buffer {
	return &Buffer{
		messages:         make([]provider.Message, 0),
		maxContextTokens: 10000,
	}
}

func (buffer *Buffer) Poke(message provider.Message) *Buffer {
	buffer.messages = append(buffer.messages, message)
	return buffer
}

func (buffer *Buffer) Peek() []provider.Message {
	return buffer.messages
}

func (buffer *Buffer) Clear() *Buffer {
	buffer.messages = buffer.messages[:0]
	return buffer
}

func (buffer *Buffer) String() string {
	out := []string{
		"<buffer>",
	}

	for _, msg := range buffer.messages {
		out = append(out, fmt.Sprintf("\t<%s>\n\t\t%s\n\t</%s>", msg.Role, msg.Content, msg.Role))
	}

	return utils.JoinWith("\n\n", out...) + "\n</buffer>"
}

/*
Truncate the buffer to the maximum context tokens, making sure to always keep the
first two messages, which are the system prompt and the user message.
*/
func (buffer *Buffer) Truncate() []provider.Message {
	// Always include first two messages (system prompt and user message)
	if len(buffer.messages) < 2 {
		return buffer.messages
	}

	maxTokens := buffer.maxContextTokens - 500 // Reserve tokens for response
	totalTokens := 0
	var truncatedMessages []provider.Message

	// Add first two messages
	truncatedMessages = append(truncatedMessages, buffer.messages[0], buffer.messages[1])
	totalTokens += buffer.estimateTokens(buffer.messages[0])
	totalTokens += buffer.estimateTokens(buffer.messages[1])

	// Start from the most recent message for the rest
	for i := len(buffer.messages) - 1; i >= 2; i-- {
		msg := buffer.messages[i]
		messageTokens := buffer.estimateTokens(msg)
		if totalTokens+messageTokens <= maxTokens {
			truncatedMessages = append([]provider.Message{msg}, truncatedMessages[2:]...)
			truncatedMessages = append(buffer.messages[0:2], truncatedMessages...)
			totalTokens += messageTokens
		} else {
			break
		}
	}

	return truncatedMessages
}

func (buffer *Buffer) estimateTokens(msg provider.Message) int { // Use tiktoken-go to estimate tokens
	encoding, err := tiktoken.EncodingForModel("gpt-4o-mini")
	if err != nil {
		log.Error("Error getting encoding", "error", err)
		return 0
	}

	tokensPerMessage := 4 // As per OpenAI's token estimation guidelines

	numTokens := tokensPerMessage
	numTokens += len(encoding.Encode(msg.Content, nil, nil))
	if msg.Role == "user" || msg.Role == "assistant" || msg.Role == "system" || msg.Role == "function" {
		numTokens += len(encoding.Encode(msg.Role, nil, nil))
	}

	return numTokens
}
