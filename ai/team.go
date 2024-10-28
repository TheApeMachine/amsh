package ai

import (
	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/provider"
)

/*
Team is a collection of Agents that work together to achieve a goal.
It compiles artifacts from the agent responses which can be used as inputs
for steps taken by other agents or teams, or as part of the final synthesis.
*/
type Team struct {
	Name      string            `json:"name"`
	Agents    map[string]*Agent `json:"agents"`
	Artifacts map[string]string `json:"artifacts"`
}

/*
NewTeam creates a new team with the given agents.
*/
func NewTeam(agents map[string]*Agent) *Team {
	return &Team{
		Agents:    agents,
		Artifacts: make(map[string]string),
	}
}

func (team *Team) Execute(agent, step, prompt string) <-chan provider.Event {
	log.Info("executing team", "agent", agent, "step", step, "prompt", prompt)
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		var accumulator string

		for event := range team.Agents[agent].Execute(prompt) {
			accumulator += event.Content
			out <- event
		}

		team.Artifacts[step] = accumulator
	}()

	return out
}
