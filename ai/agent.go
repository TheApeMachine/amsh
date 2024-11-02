package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/JesusIslam/tldr"
	"github.com/spf13/viper"
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
	ctx        context.Context
	key        string
	Name       string            `json:"name"`
	TeamName   string            `json:"team_name"`
	Role       string            `json:"role"`
	Buffer     *Buffer           `json:"agent_buffer"`
	Agents     map[string]*Agent `json:"agents"`
	scratchpad []string
	provider   provider.Provider
	params     provider.GenerationParams
	toolset    *Toolset
	iteration  int
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
	systemPrompt = utils.JoinWith("\n",
		systemPrompt,
		"",
		"<your details>",
		"name: "+fmt.Sprintf("%s-%s", teamName, role),
		"role: "+role,
		"team: "+teamName,
		"</your details>",
	)
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
		ctx:        ctx,
		key:        key,
		Name:       fmt.Sprintf("%s-%s", teamName, role),
		TeamName:   teamName,
		Role:       role,
		Buffer:     NewBuffer().AddMessage("system", systemPrompt),
		scratchpad: []string{},
		toolset:    toolset,
		provider:   provider.NewBalancedProvider(),
		iteration:  0,
	}
}

func (agent *Agent) Execute(prompt string) <-chan provider.Event {
	errnie.Note("executing agent %s", agent.Name)

	out := make(chan provider.Event)
	agent.Buffer.AddMessage("user", utils.JoinWith("\n",
		"<user prompt>",
		prompt,
		"</user prompt>",
		viper.GetViper().GetString("ai.setups."+agent.key+".templates.guidelines"),
	))

	go func() {
		defer close(out)

		for {
			var accumulator string

			errnie.Log("agent %s iteration %d\n\n%s\n\n", agent.Name, agent.iteration, agent.Buffer.GetMessages())

			for event := range agent.provider.Generate(
				context.Background(), agent.params, agent.Buffer.GetMessages(),
			) {
				if event.Type == provider.EventToken {
					event.AgentID = agent.Name
					accumulator += event.Content
					out <- event
				}
			}

			// Execute tool calls
			agent.Buffer.AddMessage("assistant", accumulator)
			agent.Buffer.AddMessage("assistant", agent.ExecuteToolCalls(accumulator))

			agent.Tweak()

			if strings.Contains(strings.ReplaceAll(strings.ToLower(accumulator), "_", ""), "task complete") {
				break
			}

			errnie.Log(
				"agent %s iteration %d\n\n%s\n\n", agent.Name, agent.iteration, agent.Buffer.GetMessages(),
			)

			agent.iteration++
		}

		out <- provider.Event{Type: provider.EventDone}
	}()

	return out
}

func (agent *Agent) ExecuteToolCalls(accumulator string) string {
	errnie.Success("executing tool calls for agent %s", agent.Name)
	// Extract all Markdown JSON blocks.
	pattern := regexp.MustCompile("(?s)```json\\s*([\\s\\S]*?)```")
	matches := pattern.FindAllStringSubmatch(accumulator, -1)

	// To get the tool that was used, we need to unmarshal the JSON string.
	for _, match := range matches {
		var data map[string]any
		if err := json.Unmarshal([]byte(match[1]), &data); err != nil {
			return "error unmarshalling tool call: " + err.Error()
		}

		if toolValue, ok := data["tool_name"].(string); ok {
			errnie.Success("executing tool %s", toolValue)

			switch toolValue {
			case "recruit":
				agent.Agents[data["role"].(string)] = NewAgent(
					agent.ctx,
					agent.Name,
					agent.TeamName,
					data["role"].(string),
					utils.JoinWith("\n",
						data["system_prompt"].(string),
						"{{tools}}",
					),
					agent.toolset,
				)

				return "new agent created"
			case "inspect":
				out := []string{}
				for _, agnt := range agent.Agents {
					bag := tldr.New()
					result, _ := bag.Summarize(agnt.Buffer.String(), 10)

					out = append(out, strings.Join([]string{
						"NAME: " + agnt.Name,
						"ROLE: " + agnt.Role,
						"\nSUMMARY: \n\n" + strings.Join(result, "\n"),
					}, "\n"))
				}

				return strings.Join(out, "\n---\n")
			default:
				return agent.toolset.Use(agent.ctx, toolValue, data)
			}
		}
	}

	return "all tool calls executed"
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
