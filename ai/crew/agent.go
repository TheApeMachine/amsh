package crew

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	openai "github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/errnie"
)

/*
AgentState represents the complete state of an Agent at a given point in time.
*/
type AgentState struct {
	ID      string  `json:"id"`
	History string  `json:"history"`
	Profile Profile `json:"profile"`
	Color   string  `json:"color"`
}

/*
Agent is a configurable wrapper around an AI model.
*/
type Agent struct {
	ctx     context.Context
	ID      string
	history string
	Profile *Profile
	Color   string
	mutex   sync.Mutex
}

/*
NewAgent initializes the agent with an ID.
*/
func NewAgent(ctx context.Context, conn *Conn, ID string, color string) *Agent {
	return &Agent{
		ctx: ctx,
		ID:  ID,
		Profile: &Profile{
			Experiences:   make([]*Experience, 0),
			Memories:      make([]*Memory, 0),
			Relationships: make([]*Relationship, 0),
		},
		Color: color,
	}
}

/*
Generate initiates the generation of the agent's response.
*/
func (agent *Agent) Generate(ctx context.Context, system, user string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		agent.NextOpenAI(system, user, out)
	}()

	return out
}

/*
Save persists the agent state to a file.
*/
func (agent *Agent) Save() *AgentState {
	state := &AgentState{
		ID:      agent.ID,
		History: agent.history,
		Profile: *agent.Profile,
		Color:   agent.Color,
	}

	jsonData, err := json.Marshal(state)
	if err != nil {
		errnie.Error(err.Error())
		return nil
	}

	os.WriteFile(fmt.Sprintf("profiles/%s.json", agent.ID), jsonData, 0644)
	return state
}

/*
Load retrieves the agent state from a file.
*/
func (agent *Agent) Load() *AgentState {
	data, err := os.ReadFile(fmt.Sprintf("profiles/%s.json", agent.ID))
	if err != nil {
		errnie.Error(err.Error())
		return nil
	}

	state := &AgentState{}
	if err := json.Unmarshal(data, state); err != nil {
		errnie.Error(err.Error())
		return nil
	}

	agent.history = state.History
	agent.Profile = &state.Profile
	agent.Color = state.Color

	return state
}

/*
NextOpenAI handles the OpenAI API interaction.
*/
func (agent *Agent) NextOpenAI(system, user string, out chan string) {
	client := openai.NewClient("YOUR_OPENAI_API_KEY")

	request := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		Stream: true,
	}

	stream, err := client.CreateChatCompletionStream(agent.ctx, request)
	if err != nil {
		errnie.Error(err.Error())
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			errnie.Error(err.Error())
			break
		}
		chunk := response.Choices[0].Delta.Content
		if chunk != "" {
			agent.history += chunk
			out <- chunk
		}
	}
}

/*
ChatCompletion generates a single, complete response from the OpenAI API.
*/
func (agent *Agent) ChatCompletion(system, user string) (string, error) {
	client := openai.NewClient("YOUR_OPENAI_API_KEY")

	request := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	}

	response, err := client.CreateChatCompletion(agent.ctx, request)
	if err != nil {
		errnie.Error(err.Error())
		return "", err
	}

	content := response.Choices[0].Message.Content
	agent.history += content

	return content, nil
}
