package ai

import (
	"fmt"
	"strings"
)

type Memory struct {
	ShortTerm []string
}

func NewMemory() *Memory {
	return &Memory{
		ShortTerm: make([]string, 0),
	}
}

func (m *Memory) ToString() string {
	builder := strings.Builder{}

	builder.WriteString("[SHORT TERM MEMORY]\n")
	for _, item := range m.ShortTerm {
		builder.WriteString(fmt.Sprintf("  - %s\n", item))
	}
	builder.WriteString("[/SHORT TERM MEMORY]")

	return builder.String()
}
