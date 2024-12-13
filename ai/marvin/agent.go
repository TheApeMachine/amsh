package marvin

import (
	"context"

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
	errnie.Trace("%s", "role", role)

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
	errnie.Trace("%s", "tools", tools)
	agent.tools = append(agent.tools, tools...)
}

func (agent *Agent) AddProcesses(processes ...*data.Artifact) {
	errnie.Trace("%s", "processes", processes)

	for _, process := range processes {
		agent.processes[process.Peek("context")] = process
	}
}

func (agent *Agent) AddSidekick(key string, sidekick *Agent) {
	errnie.Trace("%s", "key", key, "sidekick", sidekick.ID)
	agent.sidekicks[key] = append(agent.sidekicks[key], sidekick)
}

func (agent *Agent) Read(p []byte) (n int, err error) {
	// Only try to unmarshal if we have data
	if len(p) > 0 {
		artifact := data.Empty()
		artifact.Unmarshal(p)
		errnie.Trace("%s", "artifact.Payload", artifact.Peek("payload"))
	}
	return agent.buffer.Read(p)
}

func (agent *Agent) Write(p []byte) (n int, err error) {
	// Only try to unmarshal if we have data
	if len(p) > 0 {
		artifact := data.Empty()
		artifact.Unmarshal(p)
		errnie.Trace("%s", "artifact.Payload", artifact.Peek("payload"))
	}
	return agent.buffer.Write(p)
}
