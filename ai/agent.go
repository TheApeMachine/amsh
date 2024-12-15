package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/process/persona"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/utils"
	"github.com/theapemachine/errnie"
)

// AgentState represents the current state of an agent
type AgentState string

const (
	StateIdle    AgentState = "idle"
	StateWorking AgentState = "working"
	StateWaiting AgentState = "waiting"
	StateDone    AgentState = "done"
)

// Agent represents an AI agent that can perform tasks and communicate with other agents
type Agent struct {
	ctx          context.Context
	key          string
	Name         string            `json:"name"`
	TeamName     string            `json:"team_name"`
	Role         string            `json:"role"`
	Buffer       *Buffer           `json:"agent_buffer"`
	Sidekicks    map[string]*Agent `json:"sidekicks"`
	scratchpad   []string
	provider     provider.Provider
	params       provider.GenerationParams
	workloads    []string
	Toolset      *Toolset
	iteration    int
	state        AgentState
	trainingPath string
}

// NewAgent creates a new agent with integrated reasoning and learning
func NewAgent(
	ctx context.Context,
	key,
	teamName,
	role,
	systemPrompt string,
	toolset *Toolset,
) *Agent {
	errnie.Info("creating agent %s in team %s with role %s", key, teamName, role)
	errnie.Log(systemPrompt)

	trainPath := filepath.Join(os.Getenv("HOME"), ".amsh", "train")
	os.MkdirAll(trainPath, 0755)

	if toolset != nil && len(toolset.tools) > 0 {
		systemPrompt = utils.JoinWith("\n\n",
			systemPrompt,
			strings.ReplaceAll(
				viper.GetViper().GetString("ai.setups."+key+".templates.tools"),
				"{{tools}}", toolset.Schemas(),
			),
		)
	}

	return &Agent{
		ctx:          ctx,
		key:          key,
		Name:         fmt.Sprintf("%s-%s", teamName, role),
		TeamName:     teamName,
		Role:         role,
		Buffer:       NewBuffer().AddMessage("system", systemPrompt),
		Sidekicks:    make(map[string]*Agent),
		scratchpad:   []string{},
		Toolset:      toolset,
		provider:     provider.NewBalancedProvider(),
		iteration:    0,
		state:        StateIdle,
		trainingPath: trainPath,
	}
}

func (agent *Agent) AddSidekick(sidekick string) *Agent {
	var systemPrompt string

	switch sidekick {
	case "optimizer":
		optimizer := persona.Optimizer{}
		systemPrompt = optimizer.SystemPrompt(agent.Buffer.String())
	}

	agent.Sidekicks[sidekick] = NewAgent(
		agent.ctx,
		agent.key,
		agent.Name,
		sidekick,
		systemPrompt,
		nil,
	)

	return agent
}

func (agent *Agent) AddWorkloads(workloads []string) *Agent {
	errnie.Info("adding workloads to agent %s", agent.Name)

	for _, workload := range workloads {
		schema := process.ProcessMap[workload]
		agent.workloads = append(agent.workloads, schema.GenerateSchema())
		errnie.Info("added workload %s to agent %s", workload, agent.Name)
	}

	return agent
}

func (agent *Agent) Execute(prompt string) <-chan provider.Event {
	errnie.Info("executing agent %s", agent.Name)
	errnie.Log(prompt)

	out := make(chan provider.Event)

	go func() {
		defer close(out)

		agent.Buffer.AddMessage("user", utils.JoinWith("\n",
			"<request>",
			prompt,
			"</request>",
		))

		if len(agent.workloads) == 0 {
			agent.ExecuteAgent(out)
		}

		agent.ExecuteWorkloads(agent.workloads, out)
		analyzer := NewAnalyzer()
		analyzer.PostMortemAnalysis(agent)
	}()

	return out
}

func (agent *Agent) ExecuteWorkloads(workloads []string, out chan<- provider.Event) {
	if len(workloads) == 0 {
		return
	}

	for _, workload := range agent.workloads {
		errnie.Info("executing workload for agent %s", agent.Name)
		errnie.Log(workload)

		agent.Buffer.AddMessage("assistant", utils.JoinWith("\n\n",
			"You should use the following schema when completing the current workload:",
			utils.JoinWith("\n",
				"<workload>",
				workload,
				"</workload>",
			),
		))

		agent.ExecuteAgent(out)
	}
}

func (agent *Agent) ExecuteAgent(out chan<- provider.Event) {
	agent.state = StateWorking
	var accumulator string

	for event := range agent.provider.Generate(
		context.Background(), agent.params,
	) {
		if event.Type == provider.EventToken {
			event.AgentID = agent.Name
			accumulator += event.Content
			out <- event
		}
	}
	for event := range agent.provider.Generate(
		context.Background(), agent.params,
	) {
		if event.Type == provider.EventToken {
			event.AgentID = agent.Name
			accumulator += event.Content
			out <- event
		}
	}

	// Execute tool calls
	agent.Buffer.AddMessage("assistant", accumulator)
	agent.Buffer.AddMessage("assistant", ExecuteToolCalls(agent, accumulator))

	errnie.Debug("agent %s iteration %d completed", agent.Name, agent.iteration)
	errnie.Log(accumulator)
	agent.state = StateIdle
}
