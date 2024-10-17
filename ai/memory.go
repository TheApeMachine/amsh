package ai

import (
	"github.com/smallnest/ringbuffer"
	"github.com/theapemachine/amsh/ai/memory"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
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

func (m *Memory) Add(data *data.Artifact) error {
	_, err := m.ShortTerm.TryWrite(data.Marshal())
	return errnie.Error(err)
}

func (m *Memory) Query(query string) ([]map[string]interface{}, error) {
	return m.LongTerm.Query("graph", query)
}
