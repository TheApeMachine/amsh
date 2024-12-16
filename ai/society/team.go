package society

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/marvin"
	"github.com/theapemachine/amsh/data"
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
		lead:      marvin.NewAgent(ctx, "lead", name),
		members:   make(map[string]*marvin.Agent),
		ctx:       ctx,
	}
}

func (team *Team) Generate(prompt *data.Artifact) (any, error) {
	resultChan := team.pool.Schedule(uuid.New().String(), func() (any, error) {
		return team.lead.Generate(prompt)
	})

	// Wait for and return the result from the channel
	result := <-resultChan
	return result.Value, result.Error
}

func (team *Team) Add(role string, tools ...ai.Tool) {
	team.pool.Schedule(role, func() (any, error) {
		agent := marvin.NewAgent(team.ctx, role)
		agent.AddTools(tools...)
		team.members[role] = agent

		team.broadcast.Send(qpool.QuantumValue{
			Value: data.New(agent.Name, agent.Role, agent.Scope, []byte{}),
		})

		return agent, nil
	})
}
