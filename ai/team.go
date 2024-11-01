package ai

import (
	"context"
	"fmt"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/utils"
)

type Team struct {
	ctx     context.Context
	key     string
	name    string
	Agents  map[string]*Agent
	Buffer  *Buffer
	Process process.Process
}

func NewTeam(ctx context.Context, key string, proc process.Process, wg *sync.WaitGroup) *Team {
	log.Info("team created", "key", key)
	name := fmt.Sprintf("%s-%s", key, utils.NewName())

	team := &Team{
		ctx:  ctx,
		key:  key,
		name: name,
		Agents: map[string]*Agent{
			"teamlead": NewAgent(
				ctx, key, name, "teamlead",
				proc.SystemPrompt(key),
				NewToolset(),
			),
		},
		Buffer: NewBuffer(),
	}

	return team
}

func (team *Team) Execute(input string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		team.Agents["teamlead"].Execute(input)
	}()

	return out
}
