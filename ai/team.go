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
	ctx               context.Context
	Name              string            `json:"name"`
	Agents            map[string]*Agent `json:"agents"`
	Buffer            *Buffer           `json:"buffer"`
	conversationState string
}

func NewTeam(ctx context.Context, ID, key string, proc process.Process) *Team {
	log.Info("team created", "id", ID, "key", key)

	team := &Team{
		ctx:               ctx,
		Name:              ID,
		Agents:            make(map[string]*Agent),
		Buffer:            NewBuffer(),
		conversationState: "init",
	}

	team.Agents["reasoner"] = NewAgent(
		ctx, key, ID, "reasoner",
		proc.SystemPrompt(key),
		team.Buffer,
		nil,
	)

	team.Agents["toolcaller"] = NewAgent(
		ctx, key, ID, "toolcaller",
		ToolCallPrompt(key, ID),
		team.Buffer,
		NewToolset(
			viper.GetViper().GetStringSlice(
				"ai.setups."+key+".processes."+ID+".tools",
			)...,
		),
	)

	return team
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

		team.Buffer.AddMessage("user", prompt)

		for _, agent := range team.Agents {
			relevantContext := team.getRelevantContext(agent.Role)

			for event := range agent.Execute(relevantContext) {
				event.TeamID = team.Name
				out <- event
			}
		}
	}()

	return out
}

func (team *Team) getRelevantContext(role string) string {
	messages := team.Buffer.GetMessages()
	var relevant []string

	switch role {
	case "reasoner":
		for _, msg := range messages {
			if msg.Role == "user" ||
				(msg.Role == "assistant" && !strings.Contains(msg.Content, "tool")) {
				relevant = append(relevant, msg.Content)
			}
		}
	case "toolcaller":
		for _, msg := range messages {
			if msg.Role == "user" ||
				(msg.Role == "assistant" && strings.Contains(msg.Content, "tool")) {
				relevant = append(relevant, msg.Content)
			}
		}
	}

	return strings.Join(relevant, "\n\n")
}
