package society

import (
	"context"
	"time"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/marvin"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/qpool"
)

type Team struct {
	pool      *qpool.Q
	broadcast *qpool.BroadcastGroup
	lead      *marvin.Agent
	members   map[string]*marvin.Agent
	ctx       context.Context
}

func NewTeam(ctx context.Context, name string) *Team {
	pool := qpool.NewQ(ctx, 2, 5, &qpool.Config{
		SchedulingTimeout: time.Second * 5,
	})

	return &Team{
		pool:      pool,
		broadcast: pool.CreateBroadcastGroup(name, time.Hour),
		lead:      marvin.NewAgent(ctx, "lead", name, data.New("test", "system", "prompt", []byte("You are a helpful assistant."))),
		members:   make(map[string]*marvin.Agent),
		ctx:       ctx,
	}
}

func (team *Team) Generate(prompt *data.Artifact) <-chan *data.Artifact {
	return twoface.NewAccumulator(
		"team",
		"lead",
		team.lead.Name,
		prompt,
	).Yield(func(artifacts []*data.Artifact, out chan<- *data.Artifact) {
		for artifact := range team.lead.Generate(prompt) {
			out <- artifact
		}
	}).Generate()
}

func (team *Team) Add(role string, tools ...ai.Tool) {
	team.pool.Schedule(role, func() (any, error) {
		agent := marvin.NewAgent(team.ctx, role, role, data.New("test", "system", "prompt", []byte("You are a helpful assistant.")))
		agent.AddTools(tools...)
		team.members[role] = agent

		team.broadcast.Send(qpool.QuantumValue{
			Value: data.New(agent.Name, agent.Role, agent.Scope, []byte{}),
		})

		return agent, nil
	})
}
