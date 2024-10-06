package ai

import (
	"context"

	"github.com/theapemachine/amsh/ai/tools"
)

type Team struct {
	ID     string
	Name   string
	Lead   *Agent
	Agents []*Agent
	active bool
}

func NewTeam(id string, name string, agents ...*Agent) *Team {
	return &Team{
		ID:   id,
		Name: name,
		Lead: NewAgent(
			context.Background(),
			NewConn(),
			name+"-lead",
			"teamlead",
			[]tools.Tool{},
		),
		Agents: agents,
		active: false,
	}
}

func (team *Team) AddAgents(agents ...*Agent) {
	team.Agents = append(team.Agents, agents...)
}

func (team *Team) Generate(system string, user string) chan *Chunk {
	out := make(chan *Chunk)

	for _, agent := range team.Agents {
		agent.Generate(context.Background(), system, user)
	}

	return out
}

func (team *Team) Start() {
	for _, agent := range team.Agents {
		agent.Start()
	}
}

func (team *Team) Stop() {
	for _, agent := range team.Agents {
		agent.Stop()
	}
}
