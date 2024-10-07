package ai

import (
	"context"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/theapemachine/amsh/ai/memory"
	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/errnie"
)

/*
Agent represents an AI agent with its unique ID, type, prompt, memory, toolset, and active status.
It provides methods to start and stop the agent, as well as to generate responses based on a given chunk.
*/
type Agent struct {
	ctx     context.Context
	conn    *Conn
	ID      string  `json:"id"`
	Type    string  `json:"type"`
	Prompt  *Prompt `json:"prompt"`
	memory  *memory.Memory
	toolset []tools.Tool
	active  bool
}

/*
NewAgent creates a new Agent with a unique ID, a connection to an AI service, and the specified type and toolset.
It generates a random ID using the namegenerator library and initializes the agent with the provided context, type, and toolset.
*/
func NewAgent(
	ctx context.Context, Type string, toolset []tools.Tool,
) *Agent {
	errnie.Trace()

	ID := namegenerator.NewNameGenerator(time.Now().UnixNano()).Generate()
	prompt := NewPrompt(Type)
	prompt.System[0] = strings.ReplaceAll(prompt.System[0], "{name}", ID)

	return &Agent{
		ctx:     ctx,
		conn:    NewConn(),
		ID:      ID,
		Type:    Type,
		Prompt:  prompt,
		memory:  memory.NewMemory(ID),
		toolset: toolset,
		active:  false,
	}
}

/*
Generate is a method that generates responses based on a given chunk.
It returns a channel that emits Chunk objects, each containing a response from the AI service.
*/
func (agent *Agent) Generate(chunk Chunk) chan Chunk {
	errnie.Trace()

	if !agent.active {
		errnie.Warn("Agent is not active")
		return nil
	}

	out := make(chan Chunk)

	go func() {
		errnie.Info("---AGENT: %s (%s)---\n\n", agent.ID, agent.Type)
		errnie.Debug("SYSTEM:\n\n")
		for _, s := range agent.Prompt.System {
			errnie.Debug("%s", s)
		}
		errnie.Debug("USER:\n\n")
		for _, u := range agent.Prompt.User {
			errnie.Debug("%s", u)
		}

		defer close(out)

		for chunk := range agent.conn.Next(agent.ctx, agent.Prompt, chunk) {
			out <- chunk
		}

		out <- Chunk{
			SessionID: chunk.SessionID,
			Iteration: chunk.Iteration,
			Team:      chunk.Team,
			Agent:     chunk.Agent,
			Response:  "\n\n",
		}
	}()

	return out
}

/*
Start is a method that activates the agent, allowing it to generate responses.
*/
func (agent *Agent) Start() {
	errnie.Trace()
	agent.active = true
}

/*
Stop is a method that deactivates the agent, stopping it from generating responses.
*/
func (agent *Agent) Stop() {
	errnie.Trace()
	agent.active = false
}
