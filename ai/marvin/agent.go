package marvin

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/errnie"
)

/*
Agent is a type that can communicate with an AI provider and execute operations.
*/
type Agent struct {
	ID        string
	ctx       context.Context
	buffer    *Buffer
	processes map[string]*data.Artifact
	sidekicks map[string][]*Agent
	prompt    *Prompt
	role      string
	tools     []ai.Tool
}

func NewAgent(ctx context.Context, role string) *Agent {
	return &Agent{
		ID:        uuid.New().String(),
		ctx:       ctx,
		buffer:    NewBuffer(),
		processes: make(map[string]*data.Artifact),
		sidekicks: make(map[string][]*Agent),
		prompt:    NewPrompt(role),
		role:      role,
		tools:     make([]ai.Tool, 0),
	}
}

func (agent *Agent) AddTools(tools ...ai.Tool) {
	agent.tools = append(agent.tools, tools...)
}

func (agent *Agent) AddProcesses(processes ...*data.Artifact) {
	for _, process := range processes {
		agent.processes[process.Peek("context")] = process
	}
}

func (agent *Agent) AddSidekick(key string, sidekick *Agent) {
	agent.sidekicks[key] = append(agent.sidekicks[key], sidekick)
}

func (agent *Agent) Read(p []byte) (n int, err error) {
	if n = errnie.SafeMust(func() (int, error) {
		return agent.buffer.Read(p)
	}); n == 0 {
		return 0, io.EOF
	}

	artifact := data.Empty()
	errnie.Error(artifact.Unmarshal(p[:n]))

	return
}

func (agent *Agent) Write(p []byte) (n int, err error) {
	n = errnie.SafeMust(func() (int, error) {
		return agent.buffer.Write(p)
	})

	return
}
