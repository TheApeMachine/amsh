package marvin

import (
	"github.com/charmbracelet/log"
	"github.com/pkoukk/tiktoken-go"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/errnie"
)

type Buffer struct {
	messages         []*data.Artifact
	maxContextTokens int
}

func NewBuffer() *Buffer {
	return &Buffer{
		messages:         make([]*data.Artifact, 0),
		maxContextTokens: 128000,
	}
}

func (buffer *Buffer) Peek() []*data.Artifact {
	errnie.Trace("Peek", "messages_count_before_truncate", len(buffer.messages))
	buffer.truncate()
	errnie.Trace("Peek", "messages_count_after_truncate", len(buffer.messages))
	return buffer.messages
}

func (buffer *Buffer) Poke(artifact *data.Artifact) *Buffer {
	errnie.Trace("Poke", "artifact_payload", artifact.Peek("payload"))
	buffer.messages = append(buffer.messages, artifact)
	errnie.Trace("Poke", "messages_count", len(buffer.messages))
	return buffer
}

/*
Truncate the buffer to the maximum context tokens, making sure to always keep the
first two messages, which are the system prompt and the user message.
*/
func (buffer *Buffer) truncate() {
	// Always include first two messages (system prompt and user message)
	if len(buffer.messages) < 2 {
		return
	}

	maxTokens := buffer.maxContextTokens - 500 // Reserve tokens for response
	totalTokens := 0
	var truncatedMessages []*data.Artifact

	// Add first two messages
	truncatedMessages = append(truncatedMessages, buffer.messages[0], buffer.messages[1])
	totalTokens += buffer.estimateTokens(buffer.messages[0])
	totalTokens += buffer.estimateTokens(buffer.messages[1])

	// Start from the most recent message for the rest
	for i := len(buffer.messages) - 1; i >= 2; i-- {
		msg := buffer.messages[i]
		messageTokens := buffer.estimateTokens(msg)
		if totalTokens+messageTokens <= maxTokens {
			truncatedMessages = append([]*data.Artifact{msg}, truncatedMessages[2:]...)
			truncatedMessages = append(buffer.messages[0:2], truncatedMessages...)
			totalTokens += messageTokens
		} else {
			break
		}
	}

	buffer.messages = truncatedMessages
	errnie.Trace("Buffer.truncate", "message_count", len(buffer.messages))
}

func (buffer *Buffer) estimateTokens(msg *data.Artifact) int { // Use tiktoken-go to estimate tokens
	encoding, err := tiktoken.EncodingForModel("gpt-4o-mini")
	if err != nil {
		log.Error("Error getting encoding", "error", err)
		return 0
	}

	tokensPerMessage := 4 // As per OpenAI's token estimation guidelines

	numTokens := tokensPerMessage
	numTokens += len(encoding.Encode(msg.Peek("payload"), nil, nil))
	if msg.Peek("role") == "user" || msg.Peek("role") == "assistant" || msg.Peek("role") == "system" || msg.Peek("role") == "tool" {
		numTokens += len(encoding.Encode(msg.Peek("role"), nil, nil))
	}

	return numTokens
}
