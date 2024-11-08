package ai

import (
	"context"
	"fmt"
	"sync"

	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
)

/*
Competition is an executor that runs a competition between a set of agents.
A independent agent will determine the best response that will be promoted
as the final response.
*/
type Competition struct {
	ctx context.Context
	key string
}

func NewCompetition(ctx context.Context, key string) *Competition {
	errnie.Info("creating competition %s", key)

	return &Competition{
		ctx: ctx,
		key: key,
	}
}

func (competition *Competition) Run(agents []*Agent) chan provider.Event {
	errnie.Info("running competition %s", competition.key)

	output := make(chan provider.Event)

	decider := NewAgent(
		competition.ctx,
		competition.key,
		"executor",
		"decider",
		`You are an expert at determining the best response from a set of options.
		You will be given a set of responses and asked to choose the best one.
		Compare the responses based on accuracy, completeness, and clarity.
		Return only the better response without any additional commentary.`,
		nil,
	)

	go func() {
		defer close(output)

		// Create a slice to keep track of agent names in order
		agentNames := make([]string, len(agents))
		accumulators := make(map[string]string)

		var wg sync.WaitGroup
		wg.Add(len(agents))

		// Initialize agent names slice while starting goroutines
		for i, agent := range agents {
			agentNames[i] = agent.Name
			go func(agent *Agent) {
				defer wg.Done()

				for event := range agent.Execute("Please provide your response.") {
					if event.Type == provider.EventToken {
						accumulators[agent.Name] += event.Content
						output <- event
					}
				}
			}(agent)
		}

		wg.Wait()

		// Tournament rounds until we have a winner
		for len(agentNames) > 1 {
			nextRoundNames := make([]string, 0, (len(agentNames)+1)/2)
			nextRoundResponses := make(map[string]string)
			roundWg := sync.WaitGroup{}

			// Process pairs concurrently
			for i := 0; i < len(agentNames); i += 2 {
				// If we have an odd number, promote the last response
				if i+1 >= len(agentNames) {
					nextRoundNames = append(nextRoundNames, agentNames[i])
					nextRoundResponses[agentNames[i]] = accumulators[agentNames[i]]
					break
				}

				roundWg.Add(1)
				go func(idx int) {
					defer roundWg.Done()

					prompt := fmt.Sprintf("Compare these two responses and return the better one:\n\nResponse 1:\n%s\n\nResponse 2:\n%s",
						accumulators[agentNames[idx]],
						accumulators[agentNames[idx+1]],
					)

					var decision string
					for event := range decider.Execute(prompt) {
						if event.Type == provider.EventToken {
							decision += event.Content
							output <- provider.Event{
								Type:    provider.EventToken,
								Content: event.Content,
								AgentID: "decider",
							}
						}
					}

					// Generate a unique name for this round's winner
					winnerName := fmt.Sprintf("round_winner_%d", idx/2)
					nextRoundNames = append(nextRoundNames, winnerName)
					nextRoundResponses[winnerName] = decision
				}(i)
			}

			roundWg.Wait()
			agentNames = nextRoundNames
			accumulators = nextRoundResponses
		}

		// Send the final winner event
		if len(agentNames) > 0 {
			output <- provider.Event{
				Type:    provider.EventToken,
				Content: "\n\nFINAL WINNER:\n" + accumulators[agentNames[0]],
				AgentID: "competition",
			}
		}
	}()

	return output
}
