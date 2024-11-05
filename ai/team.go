package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/process"
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

func NewTeam(ctx context.Context, key string, proc process.Process) *Team {
	log.Info("team created", "key", key)
	name := fmt.Sprintf("%s-%s", key, utils.NewName())

	team := &Team{
		ctx:  ctx,
		key:  key,
		name: name,
		TeamLead: NewAgent(
			ctx, key, name, "teamlead",
			utils.JoinWith("\n\n",
				viper.GetViper().GetString("ai.setups."+key+".personas.teamlead.prompt"),
				proc.SystemPrompt(key),
			),
			NewToolset(
				viper.GetViper().GetStringSlice("ai.setups."+key+".personas.teamlead.tools")...,
			),
		),
		Agents: make(map[string]*Agent),
		Sidekicks: map[string]*Agent{
			"memory": NewAgent(
				ctx,
				key,
				name,
				"memory",
				viper.GetViper().GetString("ai.setups."+key+".personas.memory.prompt"),
				nil,
			),
			"verifier": NewAgent(
				ctx,
				key,
				name,
				"verifier",
				viper.GetViper().GetString("ai.setups."+key+".personas.verifier.prompt"),
				nil,
			),
			"toolcaller": NewAgent(
				ctx,
				key,
				name,
				"toolcaller",
				viper.GetViper().GetString("ai.setups."+key+".personas.toolcaller.prompt"),
				nil,
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

		for {
			// Check if all agents are done
			done := true
			for _, agent := range team.Agents {
				if agent.state != StateDone {
					done = false
				}
			}

			if done {
				break
			}

			accumulator := ""
			for event := range team.TeamLead.Execute(input) {
				accumulator += event.Content
				out <- event
			}

			team.RunMemory(out, accumulator, "preprompt")

			for _, agent := range team.Agents {
				for event := range agent.Execute(accumulator) {
					accumulator += event.Content
					out <- event
				}
			}

			team.RunVerifier(out, accumulator)
		}

		errnie.Debug("team execution completed")
	}()

	return out
}

func (team *Team) RunVerifier(out chan<- provider.Event, accumulator string) {
	verifier := team.Sidekicks["verifier"]
	verifier.Buffer.Clear()
	verifier.Buffer.AddMessage("system", process.NewOversight().SystemPrompt(team.key))

	for event := range verifier.Execute(accumulator) {
		out <- event
	}
}

func (team *Team) RunMemory(out chan<- provider.Event, accumulator string, prompt string) string {
	memory := team.Sidekicks["memory"]
	memory.Buffer.Clear()
	memory.Buffer.AddMessage("system",
		strings.ReplaceAll(
			viper.GetViper().GetString("ai.setups."+team.key+".personas.memory.prompt"),
			"{{preorpost}}",
			viper.GetViper().GetString("ai.setups."+team.key+".personas.memory."+prompt),
		),
	)

	memory.Buffer.AddMessage("user", utils.JoinWith("\n",
		"<context>",
		accumulator,
		"</context>",
	))

	for event := range memory.Execute(accumulator) {
		out <- event
	}

	return accumulator
}
