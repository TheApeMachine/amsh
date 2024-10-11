package ai

import (
	"github.com/smallnest/ringbuffer"
	"github.com/theapemachine/amsh/errnie"
)

type Memory struct {
	err       error
	ShortTerm *ringbuffer.RingBuffer
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
	if _, memory.err = memory.ShortTerm.Write(p); memory.err != nil {
		errnie.Error(memory.err)
	}

	return len(p), nil
}
