package system

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/types"
	"github.com/theapemachine/amsh/errnie"
)

/*
ProcessManager handles the lifecycle of a workload, mapping it across teams, agents, and processes.
It is ultimately controlled by an Agent called the Sequencer, which has been prompted to orchestrate
all the moving parts needed to make the system work.
*/
type ProcessManager struct {
	toolset *ai.Toolset
}

/*
NewProcessManager sets up the process manager, and the Agent that will act as the sequencer.
*/
func NewProcessManager(key string) *ProcessManager {
	return &ProcessManager{
		toolset: ai.NewToolset(),
	}
}

/*
Execute the process manager, using the incoming message as the initial prompt for the process.
*/
func (pm *ProcessManager) Execute(incoming string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		process := NewProcess()
		accumulator := ""

		for event := range ai.NewAgent(
			"orchestrator",
			strings.ReplaceAll(
				viper.GetString("ai.setups.marvin.orchestration"),
				"{{incoming_request}}",
				incoming,
			),
			map[string]types.Tool{},
		).Execute(incoming) {
			out <- event
			accumulator += event.Content
		}

		if err := process.Unmarshal(accumulator); err != nil {
			errnie.Error(err)

			out <- provider.Event{
				Type:    provider.EventError,
				Content: err.Error(),
			}

			return
		}

		// Set up the teams.
		teams := make(map[string]*ai.Team)

		for _, team := range process.Teams {
			agents := make(map[string]*ai.Agent)

			for _, agent := range team.Agents {
				tools := make(map[string]types.Tool)

				for _, tool := range agent.Tools {
					t, err := pm.toolset.GetTool(tool)
					if err != nil {
						errnie.Error(err)
						continue
					}

					tools[tool] = t
				}

				agents[agent.Name] = ai.NewAgent(agent.Name, agent.SystemPrompt, tools)
			}

			teams[team.Name] = ai.NewTeam(agents)
		}

		// Build an Execution, and supply the process and teams.
		for event := range NewExecution(process, teams).Execute() {
			out <- event
		}
	}()

	return out
}
