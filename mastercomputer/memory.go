package mastercomputer

import "github.com/theapemachine/amsh/utils"

type Memory struct {
	ID        string
	ShortTerm []string
}

func NewMemory() *Memory {
	return &Memory{
		ID:        utils.NewID(),
		ShortTerm: make([]string, 0),
	}
}
