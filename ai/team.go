package ai

import (
	"context"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
)

type Team struct {
	ctx    context.Context
	Name   string            `json:"name"`
	Agents map[string]*Agent `json:"agents"`
	Buffer *Buffer           `json:"buffer"`
}

func NewTeam(ctx context.Context, ID, key string, proc process.Process) *Team {
	log.Info("team created", "id", ID, "key", key)

	return &Team{
		ctx:  ctx,
		Name: ID,
		Agents: map[string]*Agent{
			"reasoner": NewAgent(
				ctx, key, ID, "reasoner", proc.SystemPrompt(key), NewBuffer(), nil,
			),
			"toolcaller": NewAgent(
				ctx, key, ID, "toolcaller", ToolCallPrompt(key, ID), NewBuffer(),
				NewToolset(
					viper.GetViper().GetStringSlice(
						"ai.setups."+key+".processes."+ID+".tools",
					)...,
				),
			),
		},
		Buffer: NewBuffer().AddMessage("system", proc.SystemPrompt(key)),
	}
}

func ToolCallPrompt(key, ID string) string {
	return strings.ReplaceAll(viper.GetViper().GetString(
		"ai.setups."+key+".processes.toolcalls.prompt",
	), "{{schemas}}", NewToolset(
		viper.GetViper().GetStringSlice(
			"ai.setups."+key+".processes."+ID+".tools",
		)...,
	).Schemas())
}

func (team *Team) Execute(prompt string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		for _, agent := range team.Agents {
			for event := range agent.Execute(prompt) {
				event.TeamID = team.Name
				out <- event
			}
		}
	}()

	return out
}
