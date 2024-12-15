package marvin

import (
	"io"

	"github.com/charmbracelet/log"
	"github.com/pkoukk/tiktoken-go"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/errnie"
)

type Buffer struct {
	messages         []*data.Artifact
	maxContextTokens int
	pr               *io.PipeReader
	pw               *io.PipeWriter
	provider         provider.Provider
}

func NewBuffer() *Buffer {
	pr, pw := io.Pipe()
	return &Buffer{
		messages:         make([]*data.Artifact, 0),
		maxContextTokens: 128000,
		pr:               pr,
		pw:               pw,
		provider:         provider.NewBalancedProvider(),
	}
}

func (buffer *Buffer) Read(p []byte) (n int, err error) {
	// Read directly from provider first
	n, err = buffer.provider.Read(p)
	if err != nil {
		return n, err
	}

	// Only try to unmarshal if we successfully read data
	if n > 0 {
		artifact := data.Empty()
		if err := artifact.Unmarshal(p[:n]); err != nil {
			errnie.Error(err)
			// Continue even if unmarshal fails - the raw data will still be returned
		}
	}

	return n, nil
}

func (buffer *Buffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	artifact := data.Empty()
	if err := artifact.Unmarshal(p); err != nil {
		errnie.Error(err)
		return 0, err
	}

	// Store in messages
	buffer.messages = append(buffer.messages, artifact)

	// Forward to provider
	return buffer.provider.Write(p)
}

func (buffer *Buffer) Close() error {
	buffer.messages = buffer.messages[:0]
	return nil
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
