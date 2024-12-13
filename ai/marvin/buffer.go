package marvin

import (
	"io"

	"github.com/charmbracelet/log"
	"github.com/pkoukk/tiktoken-go"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/errnie"
)

type Buffer struct {
	messages         []*data.Artifact
	maxContextTokens int
	pr               *io.PipeReader
	pw               *io.PipeWriter
}

func NewBuffer() *Buffer {
	errnie.Trace("%s", "Buffer.NewBuffer", "new")

	pr, pw := io.Pipe()
	return &Buffer{
		messages:         make([]*data.Artifact, 0),
		maxContextTokens: 128000,
		pr:               pr,
		pw:               pw,
	}
}

func (buffer *Buffer) Read(p []byte) (n int, err error) {
	// Only try to unmarshal if we have data
	if len(p) > 0 {
		artifact := data.Empty()
		artifact.Unmarshal(p)
		errnie.Trace("%s", "payload", artifact.Peek("payload"))
	}

	if buffer.pr == nil {
		return 0, io.EOF
	}

	buffer.truncate()

	// Marshal all messages into a single artifact
	artifact := data.New("buffer", "system", "context", nil)
	for _, msg := range buffer.messages {
		artifact.Append(msg.Peek("payload"))
	}

	// Write to pipe in goroutine to prevent blocking
	go func() {
		defer buffer.pw.Close()
		artifact.Marshal(p)
	}()

	return buffer.pr.Read(p)
}

func (buffer *Buffer) Write(p []byte) (n int, err error) {
	// Only try to unmarshal if we have data
	if len(p) == 0 {
		return 0, nil
	}

	artifact := data.Empty()
	artifact.Unmarshal(p)
	errnie.Trace("%s", "artifact.Payload", artifact.Peek("payload"))
	buffer.messages = append(buffer.messages, artifact)
	return len(p), nil
}

func (buffer *Buffer) Close() error {
	errnie.Trace("%s", "Buffer.Close", "close")

	buffer.messages = buffer.messages[:0]
	return nil
}

/*
Truncate the buffer to the maximum context tokens, making sure to always keep the
first two messages, which are the system prompt and the user message.
*/
func (buffer *Buffer) truncate() {
	errnie.Trace("%s", "Buffer.truncate", "truncate")

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
	errnie.Trace("%s", "msg", msg.Peek("role"))

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
