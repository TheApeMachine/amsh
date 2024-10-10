package ai

import "strings"

type Memory struct {
	ShortTerm []string
}

func NewMemory() *Memory {
	return &Memory{
		ShortTerm: make([]string, 0),
	}
}

func (m *Memory) String() string {
	return strings.Join([]string{
		"[CURRENT CONTEXT]",
		strings.Join(m.ShortTerm, "\n\n"),
		"[/CURRENT CONTEXT]",
	}, "\n\n")
}

func (m *Memory) Add(item string) {
	m.ShortTerm = append(m.ShortTerm, item)
}
