package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
	"golang.org/x/exp/rand"
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
	ctx         context.Context
	Name        string  `json:"name"`
	Role        string  `json:"role"`
	TeamMember  string  `json:"team_member"`
	TeamBuffer  *Buffer `json:"team_buffer"`
	AgentBuffer *Buffer `json:"agent_buffer"`
	provider    provider.Provider
	state       AgentState
	params      provider.GenerationParams
	toolset     *Toolset
	iteration   int
}

// NewAgent creates a new agent with integrated reasoning and learning
func NewAgent(
	ctx context.Context,
	key,
	role,
	teamMember string,
	systemPrompt string,
	teamBuffer *Buffer,
	toolset *Toolset,
) *Agent {
	errnie.Success("new agent created %s (%s) <%s>", key, role, teamMember)
	errnie.Log(
		"key: %s\nrole: %s\nteam member: %s\nsystem prompt: %s",
		key, role, teamMember, systemPrompt,
	)

	return &Agent{
		ctx:         ctx,
		Name:        fmt.Sprintf("%s-%s-%s", key, role, teamMember),
		Role:        role,
		TeamMember:  teamMember,
		TeamBuffer:  teamBuffer,
		AgentBuffer: NewBuffer().AddMessage("system", systemPrompt),
		toolset:     toolset,
		provider:    provider.NewBalancedProvider(),
		state:       StateIdle,
		iteration:   0,
	}
}

func (agent *Agent) WithToolset(toolset *Toolset) *Agent {
	agent.toolset = toolset
	return agent
}

func (agent *Agent) Execute(prompt string) <-chan provider.Event {
	errnie.Note("executing agent %s", agent.Name)

	if agent.TeamMember == "toolcaller" && agent.toolset == nil {
		errnie.Note("no tools for agent %s", agent.Name)
	}

	out := make(chan provider.Event)

	if agent.toolset != nil {
		agent.AgentBuffer.AddMessage("system", utils.JoinWith("\n",
			"<toolset>",
			agent.toolset.Schemas(),
			"</toolset>",
		))
	}

	agent.AgentBuffer.AddMessage("user", utils.JoinWith("\n\n",
		utils.JoinWith("\n", "<prompt>", prompt, "</prompt>"),
		utils.JoinWith("\n",
			fmt.Sprintf("<scratchpad (iteration: %d)>", agent.iteration),
			"(you can iterate as much as you need, just say Task Complete if you're done)",
			"</scratchpad>",
		),
	))

	go func() {
		defer close(out)

		for {
			var accumulator string

			for event := range agent.provider.Generate(
				context.Background(), agent.params, agent.AgentBuffer.GetMessages(),
			) {
				if event.Type == provider.EventToken {
					event.AgentID = agent.Name
					accumulator += event.Content
					out <- event
				}
			}

			// Execute tool calls
			agent.ExecuteToolCalls(accumulator)

			agent.AgentBuffer.AddMessage("assistant", utils.JoinWith("\n",
				fmt.Sprintf("<scratchpad (iteration: %d)>", agent.iteration),
				accumulator,
				"(you can iterate as much as you need, just say Task Complete if you're done)",
				fmt.Sprintf("</scratchpad (iteration: %d)>", agent.iteration),
			))

			agent.Tweak()

			if strings.Contains(strings.ToLower(accumulator), "task complete") {
				agent.TeamBuffer.AddMessage("assistant", utils.JoinWith("\n",
					fmt.Sprintf("<%s>", agent.Name),
					accumulator,
					fmt.Sprintf("</%s>", agent.Name),
				))

				break
			}

			errnie.Log(
				"agent %s iteration %d\n\n%s\n\n",
				agent.Name, agent.iteration, accumulator,
			)

			agent.iteration++
		}

		out <- provider.Event{Type: provider.EventDone}
	}()

	return out
}

func (agent *Agent) ExecuteToolCalls(accumulator string) {
	errnie.Success("executing tool calls for agent %s", agent.Name)
	// Extract all Markdown JSON blocks.
	pattern := regexp.MustCompile("(?s)```json\\s*([\\s\\S]*?)```")
	matches := pattern.FindAllStringSubmatch(accumulator, -1)

	// To get the tool that was used, we need to unmarshal the JSON string.
	for _, match := range matches {
		var data map[string]any
		if err := json.Unmarshal([]byte(match[1]), &data); err != nil {
			agent.TeamBuffer.AddMessage("assistant", "Error unmarshalling tool call: "+err.Error())
			continue
		}

		if toolValue, ok := data["tool_name"].(string); ok {
			errnie.Success("executing tool %s", toolValue)
			agent.toolset.Use(agent.ctx, toolValue, data)
		}
	}
}

func (agent *Agent) Tweak() provider.GenerationParams {
	agent.params.Interestingness = agent.MeasureInterestingness()
	if len(agent.params.InterestingnessHistory) > 5 {
		// If results getting boring, increase temperature
		if average(agent.params.InterestingnessHistory) < 0.5 {
			agent.params.Temperature *= 1.1
			agent.params.TopK += 10
		}

		// If results too wild, decrease temperature
		if average(agent.params.InterestingnessHistory) > 0.8 {
			agent.params.Temperature *= 0.9
			agent.params.TopK -= 5
		}

		// Keep a moving window
		agent.params.InterestingnessHistory = agent.params.InterestingnessHistory[1:]
	}

	return agent.params
}

func (agent *Agent) MeasureInterestingness() float64 {
	interestingness := measureInterestingness()
	agent.params.InterestingnessHistory = append(
		agent.params.InterestingnessHistory, interestingness,
	)

	return interestingness
}

func measureInterestingness() float64 {
	return rand.Float64()
}

func average(values []float64) float64 {
	return sum(values) / float64(len(values))
}

func sum(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total
}
