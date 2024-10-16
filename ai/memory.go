package ai

import (
	"errors"
	"io"

	"github.com/smallnest/ringbuffer"
	"github.com/theapemachine/amsh/ai/memory"
	"github.com/theapemachine/amsh/data"
)

// Memory represents the memory of a worker, including short-term and long-term memory.
type Memory struct {
	ShortTerm *ringbuffer.RingBuffer
	LongTerm  *memory.LongTerm
}

// NewMemory creates a new Memory instance for a worker.
func NewMemory(agentID string) *Memory {
	return &Memory{
		ShortTerm: ringbuffer.New(1024),        // Adjust the size as needed
		LongTerm:  memory.NewLongTerm(agentID), // Pass the agent ID for identification
	}
}

// Read reads data from short-term memory.
func (m *Memory) Read(p []byte) (n int, err error) {
	return m.ShortTerm.Read(p)
}

// Write writes data to memory, deciding whether it goes to short-term or long-term memory.
func (m *Memory) Write(p []byte) (n int, err error) {
	artifact := data.Unmarshal(p)
	if artifact == nil {
		return 0, errors.New("failed to unmarshal artifact")
	}

	scope := artifact.Peek("scope")
	switch scope {
	case "short-term":
		_, err = m.ShortTerm.Write(artifact.Marshal())
		if err != nil {
			return 0, err
		}
		return len(p), nil
	case "vector":
		_, err = io.Copy(m.LongTerm, artifact)
		if err != nil {
			return 0, err
		}
		return len(p), nil
	case "graph":
		_, err = io.Copy(m.LongTerm, artifact)
		if err != nil {
			return 0, err
		}
		return len(p), nil
	default:
		return 0, errors.New("invalid memory scope")
	}
}

// Close closes the long-term memory resources.
func (m *Memory) Close() error {
	if m.LongTerm != nil {
		return m.LongTerm.Close()
	}
	return nil
}
