package ai

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
)

/*
Team is a collection of Agents that work together to achieve a goal.
It compiles artifacts from the agent responses which can be used as inputs
for steps taken by other agents or teams, or as part of the final synthesis.
*/
type Team struct {
	Name      string            `json:"name"`
	Teamlead  *Agent            `json:"teamlead"`
	Agents    map[string]*Agent `json:"agents"`
	Artifacts map[string]string `json:"artifacts"`
	Reasoner  *Agent            `json:"reasoner"`
}

/*
NewTeam creates a new team with the given agents.
*/
func NewTeam(key, name string, agents map[string]*Agent) *Team {
	log.Info("NewTeam", "key", key, "name", name)

	return &Team{
		Name: name,
		Teamlead: NewAgent(
			fmt.Sprintf("%s-teamlead", key),
			"teamlead",
			strings.ReplaceAll(
				strings.ReplaceAll(
					viper.GetString(fmt.Sprintf("ai.setups.%s.teamlead.prompt", key)),
					"{{schemas}}",
					process.NewDiscussion().GenerateSchema(),
				),
				"{{team_members}}",
				getTeamMembers(agents),
			),
		),
		Agents:    agents,
		Artifacts: make(map[string]string),
		Reasoner: NewAgent(key, "reasoner", strings.ReplaceAll(
			viper.GetString(fmt.Sprintf("ai.setups.%s.reasoner.prompt", key)),
			"{{schemas}}",
			process.NewThinking().GenerateSchema(),
		)),
	}
}

func (team *Team) Execute(step process.Step) <-chan provider.Event {
	log.Info("executing team", "team", team.Name, "step_key", step.StepKey)
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		var accumulator string

		for event := range team.Reasoner.Execute(strings.Join([]string{
			"<context>",
			fmt.Sprintf("\tstep_key: %s", step.StepKey),
			fmt.Sprintf("\tinputs: %v", step.Inputs),
			fmt.Sprintf("\toutputs: %v", step.Outputs),
			"</context>",
			"",
		}, "\n")) {
			accumulator += event.Content
			out <- event
		}

		for event := range team.Teamlead.Execute(strings.Join([]string{
			"<context>",
			accumulator,
			"</context>",
			"",
			"<task>",
			fmt.Sprintf("\t%s", step.Prompt),
			"</task>",
		}, "\n")) {
			accumulator += event.Content
			out <- event
		}

		execution := process.NewExecution().Extract(accumulator)

		if execution == nil {
			errnie.Error(errors.New("failed to extract execution"))
			return
		}

		for _, action := range execution.Actions {
			for event := range team.Agents[action.TeamMember].Execute(action.Prompt) {
				out <- event
			}
		}
	}()

	return out
}

func getTeamMembers(agents map[string]*Agent) string {
	var members []string
	for _, agent := range agents {
		members = append(members, agent.Name)
	}
	return strings.Join(members, ", ")
}
