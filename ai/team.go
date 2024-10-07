package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/errnie"
)

/*
Team represents a team of AI agents with a lead agent and a set of agents.
It provides methods to start and stop the team, as well as to generate responses based on a given chunk.
*/
type Team struct {
	ctx         context.Context
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Lead        *Agent   `json:"lead"`
	Agents      []*Agent `json:"agents"`
	Prompt      *Prompt  `json:"prompt"`
	Response    string   `json:"response"`
	history     []Chunk
	fullHistory string
	active      bool
}

/*
NewTeam creates a new Team with a unique ID, a name, a lead agent, and a set of agents.
It generates a random ID using the namegenerator library and initializes the team with the provided context, name, and agents.
*/
func NewTeam(ctx context.Context, name string, agents ...*Agent) *Team {
	errnie.Trace()

	return &Team{
		ctx:  ctx,
		ID:   namegenerator.NewNameGenerator(time.Now().UnixNano()).Generate(),
		Name: name,
		Lead: NewAgent(
			ctx,
			"teamlead",
			[]tools.Tool{},
		),
		Agents:  agents,
		Prompt:  NewPrompt(name),
		history: make([]Chunk, 0),
		active:  false,
	}
}

func (team *Team) Initialize() {
	errnie.Trace()
	team.replacePlaceholders()

	team.Lead.Prompt.System = append(team.Lead.Prompt.System, team.Prompt.System...)
	team.Lead.Prompt.User = append(team.Lead.Prompt.User, team.Prompt.User...)

	for _, agent := range team.Agents {
		agent.Prompt.System = append(agent.Prompt.System, team.Prompt.System...)
		agent.Prompt.User = append(agent.Prompt.User, team.Prompt.User...)
	}
}

/*
AddAgents is a method that adds additional agents to the team.
It appends the provided agents to the existing list of agents in the team.
*/
func (team *Team) AddAgents(agents ...*Agent) {
	errnie.Trace()
	team.Agents = append(team.Agents, agents...)
}

/*
SetPrompt is a method that sets the prompt for the team.
It updates the prompt for the lead agent and all team agents.
*/
func (team *Team) SetPrompt(prompt string) {
	errnie.Trace()
	formatted := fmt.Sprintf(
		"### Original Prompt\n\n> %s\n\n",
		prompt,
	)
	team.Lead.Prompt.User = append(team.Lead.Prompt.User, formatted)
	team.Lead.Prompt.User = append(team.Lead.Prompt.User, "")

	for _, agent := range team.Agents {
		agent.Prompt.User = append(agent.Prompt.User, formatted)
		agent.Prompt.User = append(agent.Prompt.User, "")
	}
}

func (team *Team) makeChunk(chunk Chunk) Chunk {
	errnie.Trace()

	out := chunk
	out.Response = ""

	return out
}

/*
Generate is a method that generates responses based on a given chunk.
It returns a channel that emits Chunk objects, each containing a response from the AI service.
*/
func (team *Team) Generate(chunk Chunk) chan Chunk {
	errnie.Trace()

	if !team.active {
		errnie.Warn("Team is not active")
		return nil
	}

	out := make(chan Chunk)

	go func() {
		defer close(out)

		errnie.Info("---TEAM: %s---\n", team.ID)
		idx := 0

		for {
			agent := team.Agents[idx%len(team.Agents)]

			for chunk := range team.Lead.Generate(team.makeChunk(chunk)) {
				chunk.Agent = team.Lead
				team.history = append(team.history, chunk)
				out <- chunk
			}

			if shouldEnd := team.updateConversation(team.Lead); shouldEnd {
				break
			}

			for chunk := range agent.Generate(team.makeChunk(chunk)) {
				chunk.Agent = agent
				team.history = append(team.history, chunk)
				out <- chunk
			}

			if shouldEnd := team.updateConversation(agent); shouldEnd {
				break
			}

			idx++
		}
	}()

	return out
}

/*
Start is a method that activates the team, allowing its agents to generate responses.
It iterates through the list of agents and starts each agent.
*/
func (team *Team) Start() {
	errnie.Trace()

	for _, agent := range team.Agents {
		agent.Start()
	}

	team.Lead.Start()
	team.active = true
}

/*
Stop is a method that deactivates the team, stopping all its agents from generating responses.
It iterates through the list of agents and stops each agent.
*/
func (team *Team) Stop() {
	errnie.Trace()

	for _, agent := range team.Agents {
		agent.Stop()
	}

	team.active = false
}

func (team *Team) updateConversation(agent *Agent) bool {
	errnie.Trace()
	shouldEnd := false

	conversation := fmt.Sprintf("[%s @ %s] ", agent.ID, time.Now().Format("2006-01-02 15:04:05"))
	for _, chunk := range team.history {
		conversation += chunk.Response
	}

	team.fullHistory += conversation

	team.Lead.Prompt.User[len(team.Lead.Prompt.User)-1] = fmt.Sprintf(
		"\n<details><summary>Conversation History</summary>\n\n  %s</details>\n\n## Response\n\n> Please provide a response to the current conversation.\n\n",
		team.fullHistory,
	)

	team.history = make([]Chunk, 0)
	commands := ExtractJSON(conversation)

	if len(commands) > 0 {
		for _, command := range commands {
			switch command["action"] {
			case "update":
				conversation = strings.Join([]string{
					command["summary"].(string),
					command["next_steps"].(string),
				}, "\n\n")
			case "finalize":
				shouldEnd = true
				buf, err := json.Marshal(command)
				if err != nil {
					errnie.Error(err)
				}

				conversation = string(buf)
				team.Response = conversation
			}
		}
	}

	// Update the last user prompt with the conversation history.
	team.Prompt.User[len(team.Prompt.User)-1] = fmt.Sprintf(
		"\n<details><summary>Conversation History</summary>\n\n  %s</details>\n\n## Response\n\n> Please provide a response to the current conversation.\n\n",
		conversation,
	)

	// Update the last user prompt for each agent.
	for _, agent := range team.Agents {
		newPrompt := team.Prompt.User[len(team.Prompt.User)-1]
		errnie.Debug("Updating history for agent: %s\n---\n%s\n---\n\n", agent.ID, newPrompt)
		agent.Prompt.User[len(agent.Prompt.User)-1] = newPrompt
	}

	return shouldEnd
}

func (team *Team) replacePlaceholders() {
	errnie.Trace()

	teamMembers := fmt.Sprintf(" - **%s** (%s)\n", team.Lead.ID, team.Lead.Type)
	for _, agent := range team.Agents {
		teamMembers += fmt.Sprintf("  - **%s** (%s)\n", agent.ID, agent.Type)
	}

	team.Prompt.System[0] = strings.ReplaceAll(team.Prompt.System[0], "{team}", teamMembers)
}
