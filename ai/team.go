package ai

import (
	"context"
	"fmt"

	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/process/layering"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Team struct {
	ctx       context.Context
	key       string
	name      string
	TeamLead  *Agent
	Agents    map[string]*Agent
	Sidekicks map[string]*Agent
	Buffer    *Buffer
	Process   process.Process
}

func NewTeam(ctx context.Context, key string) *Team {
	errnie.Info("team created %s", key)
	name := fmt.Sprintf("%s-%s", key, utils.NewName())

	team := &Team{
		ctx:    ctx,
		key:    key,
		name:   name,
		Buffer: NewBuffer(),
	}

	return team
}

func (team *Team) Execute(workload layering.Workload) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)
	}()

	return out
}
