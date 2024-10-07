package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/errnie"
)

type stage string

const (
	PreResponse  stage = "pre-response"
	PostResponse stage = "post-response"
)

/*
Agent represents an AI agent with its unique ID, type, prompt, memory, toolset, and active status.
It provides methods to start and stop the agent, as well as to generate responses based on a given chunk.
*/
type Agent struct {
	ctx      context.Context
	conn     *Conn
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Prompt   *Prompt  `json:"prompt"`
	Memories []string `json:"memories"`
	toolset  []tools.Tool
	active   bool
	response []Chunk
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
		toolset: toolset,
		active:  false,
	}
}

func (agent *Agent) SetPrompt(prompt string) {
	agent.Prompt.User = append(agent.Prompt.User, fmt.Sprintf(
		"## Final Response\n\n**Extract any commands from the following response.**\n\n%s\n\n## Response\n\n> Please provide your response.\n\n",
		prompt,
	))
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
		chunk.Agent = agent
		tmp := agent.getChunk(chunk)
		tmp.Agent = agent
		tmp.Response = fmt.Sprintf("[%s (%s)] ", agent.ID, agent.Type)
		out <- tmp
		defer close(out)

		prompt := agent.handleMemory(PreResponse, chunk)

		for chunk := range agent.conn.Next(agent.ctx, prompt, chunk) {
			agent.response = append(agent.response, chunk)
			out <- chunk
		}

		prompt = agent.handleMemory(PostResponse, chunk)
	}()

	return out
}

func (agent *Agent) getChunk(chunk Chunk) Chunk {
	errnie.Trace()
	return Chunk{
		SessionID: chunk.SessionID,
		Iteration: chunk.Iteration,
		Team:      chunk.Team,
	}
}

/*
handleMemory is a method that handles the memory of the agent.
*/
func (agent *Agent) handleMemory(stage stage, chunk Chunk) *Prompt {
	errnie.Trace()

	if agent.Type == "memory" {
		return agent.Prompt
	}

	memory := NewAgent(agent.ctx, "memory", []tools.Tool{})
	memory.Start()

	ReplaceHolders(
		memory.Prompt.System[0], [][]string{{"{name}", chunk.Agent.ID}},
	)

	extra := "(only use storage commands)"
	if stage == PreResponse {
		extra = "(only use query commands)"
	}

	ReplaceHolders(
		memory.Prompt.User[0], [][]string{
			{"{session_id}", chunk.SessionID},
			{"{team}", chunk.Team.ID},
			{"{agent}", chunk.Agent.ID},
			{"{stage}", string(stage) + " " + extra},
			{"{response}", ChunksToResponse(agent.response)},
		},
	)

	response := ""
	chunk = agent.getChunk(chunk)
	chunk.Agent = memory

	for chunk := range memory.Generate(chunk) {
		response += chunk.Response
	}

	commands := ExtractJSON(response)

	for _, command := range commands {
		if cmd, ok := command["command"].(string); ok {
			parameters, paramsOk := command["parameters"].(map[string]any)
			if !paramsOk {
				continue
			}

			switch cmd {
			case "store_local":
				if data, ok := parameters["data"].(string); ok {
					agent.Memories = append(agent.Memories, data)
				}
			case "query_local":
				if query, ok := parameters["query"].(string); ok {
					builder := strings.Builder{}

					for _, memory := range agent.Memories {
						if strings.Contains(
							strings.ToLower(memory),
							strings.ToLower(query),
						) {
							builder.WriteString("  *" + memory + "*\n\n")
						}
					}

					ReplaceHolders(
						agent.Prompt.User[0], [][]string{
							{"{memory}", builder.String()},
						},
					)
				}
			case "query_graph", "store_graph", "query_vector", "store_vector":
			default:
				errnie.Error(errors.New("unknown command" + cmd))
			}
		}
	}

	return agent.Prompt
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
