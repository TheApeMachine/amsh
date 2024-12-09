package interaction

import (
	"github.com/theapemachine/amsh/ai/marvin"
	"github.com/theapemachine/amsh/ai/provider"
)

type Discussion struct {
	agents map[string][]*marvin.Agent
}

func NewDiscussion() *Discussion {
	return &Discussion{
		agents: make(map[string][]*marvin.Agent),
	}
}

func (discussion *Discussion) AddAgent(role string, agent *marvin.Agent) {
	discussion.agents[role] = append(discussion.agents[role], agent)
}

func (discussion *Discussion) Generate() chan *provider.Event {
	out := make(chan *provider.Event)

	go func() {
		defer close(out)

		var accumulator *provider.Accumulator

	}()

	return out
}
