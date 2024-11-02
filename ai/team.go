package ai

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
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

func NewTeam(ctx context.Context, key string, proc process.Process) *Team {
	log.Info("team created", "key", key)
	name := fmt.Sprintf("%s-%s", key, utils.NewName())

	team := &Team{
		ctx:  ctx,
		key:  key,
		name: name,
		Agents: map[string]*Agent{
			"teamlead": NewAgent(
				ctx, key, name, "teamlead",
				utils.JoinWith("\n\n",
					viper.GetViper().GetString("ai.setups."+key+".personas.teamlead.prompt"),
					proc.SystemPrompt(key),
				),
				NewToolset(
					viper.GetViper().GetStringSlice("ai.setups."+key+".personas.teamlead.tools")...,
				),
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

		for event := range team.Agents["teamlead"].Execute(input) {
			out <- event
		}
	}()

	return out
}
