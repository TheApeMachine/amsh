package interaction

import (
	"github.com/theapemachine/amsh/ai/marvin"
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
