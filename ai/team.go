package ai

import (
	"context"
	"sync"

	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/utils"
)

type Team struct {
	ctx    context.Context
	Name   string            `json:"name"`
	Agents map[string]*Agent `json:"agents"`
	Buffer *Buffer           `json:"buffer"`
	wg     *sync.WaitGroup
}

func NewTeam(ctx context.Context, key, systemPrompt string) *Team {
	return &Team{
		ctx:    ctx,
		Name:   utils.NewName(),
		Agents: make(map[string]*Agent),
		Buffer: NewBuffer().AddMessage("system", systemPrompt),
		wg:     &sync.WaitGroup{},
	}
}

func (team *Team) Execute(prompt string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		for {

		}
	}()

	return out
}
