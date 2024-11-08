package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/process/layering"
	"github.com/theapemachine/amsh/ai/process/persona"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Team struct {
	ctx       context.Context
	key       string
	name      string
	TeamLead  *Agent
	Agents    map[string]*Agent
	Sidekicks map[string]*Agent
	Buffer    *Buffer
	Process   process.Process
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
		Buffer: NewBuffer(),
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

				agents := []*Agent{}
				for _, agent := range response.Agents {
					agents = append(agents, NewAgent(
						team.ctx,
						team.key,
						agent.Name,
						agent.Role,
						agent.SystemPrompt,
						nil,
					).AddSidekick("optimizer").AddWorkloads(agent.Workloads))
				}

				for event := range NewCompetition(team.ctx, team.key).Run(agents) {
					out <- event
				}
			}
		}
	}()

	return out
}
