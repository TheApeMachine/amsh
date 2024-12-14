package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/process/layering"
	"github.com/theapemachine/amsh/ai/process/persona"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/utils"
	"github.com/theapemachine/errnie"
)

type Team struct {
	ctx      context.Context
	key      string
	name     string
	TeamLead *Agent
	Process  process.Process
}

func NewTeam(ctx context.Context, key string) *Team {
	errnie.Info("team created %s", key)
	name := fmt.Sprintf("%s-%s", key, utils.NewName())

	team := &Team{
		ctx:  ctx,
		key:  key,
		name: name,
		TeamLead: NewAgent(
			ctx,
			key,
			name,
			"teamlead",
			persona.SystemPrompt("teamlead"),
			nil,
		),
	}

	return team
}

func (team *Team) Execute(workload layering.Workload) <-chan provider.Event {
	errnie.Info("executing team %s", team.name)
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		// Stringify the workload
		workloads := errnie.SafeMust(func() ([]byte, error) {
			return json.Marshal(workload)
		})

		accumulator := ""

		for event := range team.TeamLead.Execute(utils.JoinWith("\n\n",
			"Consider the following workloads, and recruit the appropriate agents to complete them.",
			utils.JoinWith("\n",
				"<workloads>",
				string(workloads),
				"</workloads>",
			),
		)) {
			accumulator += event.Content
			out <- event
		}

		extracted := utils.ExtractCodeBlocks(accumulator)

		for _, blocks := range extracted {
			for _, block := range blocks {

				response := persona.Teamlead{}
				errnie.MustVoid(json.Unmarshal([]byte(block), &response))

				agents := map[string]*Agent{}
				for _, agent := range response.Agents {
					agents[agent.Name] = NewAgent(
						team.ctx,
						team.key,
						agent.Name,
						agent.Role,
						agent.SystemPrompt,
						nil,
					).AddSidekick("optimizer").AddWorkloads(agent.Workloads)
				}

				for _, interaction := range response.Interactions {
					if interaction.ProcessInParallel {
						team.parallelInteraction(interaction, agents, out)
					} else {
						team.sequentialInteraction(interaction, agents, out)
					}
				}
			}
		}
	}()

	return out
}

func (team *Team) parallelInteraction(
	interaction persona.Interaction,
	agents map[string]*Agent,
	out chan<- provider.Event,
) map[string]string {
	var wg sync.WaitGroup
	wg.Add(len(interaction.Agents))
	accumulator := map[string]string{}

	for _, agent := range interaction.Agents {
		go func(agent string, wg *sync.WaitGroup) {
			defer wg.Done()

			for event := range agents[agent].Execute(interaction.Prompt) {
				accumulator[event.AgentID] += event.Content
				out <- event
			}
		}(agent, &wg)
	}

	wg.Wait()
	return accumulator
}

func (team *Team) sequentialInteraction(
	interaction persona.Interaction,
	agents map[string]*Agent,
	out chan<- provider.Event,
) map[string]string {
	accumulator := map[string]string{}

	for _, agent := range interaction.Agents {
		for event := range agents[agent].Execute(interaction.Prompt) {
			accumulator[event.AgentID] += event.Content
			out <- event
		}
	}

	return accumulator
}
