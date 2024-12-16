package marvin

import (
	"context"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/utils"
)

type Agent struct {
	Name      string
	Role      string
	Scope     string
	ctx       context.Context
	buffer    *Buffer
	processes map[string]*data.Artifact
	sidekicks map[string][]*Agent
	tools     []ai.Tool
	provider  *provider.Provider
}

func NewAgent(ctx context.Context, role, scope string) *Agent {
	return &Agent{
		Name:      utils.NewName(),
		Role:      role,
		Scope:     scope,
		ctx:       ctx,
		buffer:    NewBuffer(),
		processes: make(map[string]*data.Artifact),
		sidekicks: make(map[string][]*Agent),
		tools:     make([]ai.Tool, 0),
		provider:  provider.NewBalancedProvider(),
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

func (agent *Agent) GetCapabilities() []string {
	capabilities := make([]string, 0)

	for _, tool := range agent.tools {
		capabilities = append(capabilities, tool.Name())
	}

	return capabilities
}

func (agent *Agent) Generate(prompt *data.Artifact) chan *data.Artifact {
	return nil
}
