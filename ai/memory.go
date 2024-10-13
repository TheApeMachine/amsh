package ai

import (
	"io"

	"github.com/smallnest/ringbuffer"
	"github.com/theapemachine/amsh/ai/memory"
	"github.com/theapemachine/amsh/data"
)

type Memory struct {
	ShortTerm *ringbuffer.RingBuffer
	LongTerm  *memory.LongTerm
}

func NewMemory() *Memory {
	return &Memory{
		ShortTerm: ringbuffer.New(10),
	}
}

func (memory *Memory) Read(p []byte) (n int, err error) {
	return memory.ShortTerm.Read(p)
}

func (memory *Memory) Write(p []byte) (n int, err error) {
	artifact := data.Empty
	artifact = artifact.Unmarshal(p)

	if artifact.Peek("scope") == "short-term" {
		io.Copy(memory.ShortTerm, artifact)
	} else {
		io.Copy(memory.LongTerm, artifact)
	}

	return len(p), nil
}
