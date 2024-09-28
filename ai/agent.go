package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/errnie"
	"google.golang.org/api/iterator"
)

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       string `json:"level"`
}

type Memory struct {
	Timestamp time.Time `json:"timestamp"`
	Scene     string    `json:"scene"`
	Action    string    `json:"action"`
	Content   string    `json:"content"`
}

type Experience struct {
	Title       string `json:"title"`
	Location    string `json:"location"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Description string `json:"description"`
}

type Relationship struct {
	Target      string       `json:"target"`
	Type        string       `json:"type"`
	Status      string       `json:"status"`
	Description string       `json:"description"`
	Experiences []Experience `json:"experiences"`
	Memories    []Memory     `json:"memories"`
}

type Profile struct {
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Skills        []Skill        `json:"skills"`
	Experiences   []Experience   `json:"experiences"`
	Memories      []Memory       `json:"memories"`
	Relationships []Relationship `json:"relationships"`
}

func (profile *Profile) String() string {
	json, err := json.Marshal(profile)

	if err != nil {
		errnie.Error(err.Error())
		return ""
	}

	return string(json)
}

func (profile *Profile) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), profile)
}

type AgentState struct {
	ID      string  `json:"id"`
	History string  `json:"history"`
	Profile Profile `json:"profile"`
	Color   string  `json:"color"`
}

// Agent is a configurable wrapper around an AI model.
type Agent struct {
	ctx     context.Context
	conn    *Conn
	ID      string
	history string
	profile *Profile
	color   string
	stream  *openai.ChatCompletionStream
	err     error
}

// NewAgent initializes the agent with an ID.
func NewAgent(ctx context.Context, conn *Conn, ID string, color string) *Agent {
	return &Agent{
		ctx:  ctx,
		conn: conn,
		ID:   ID,
		profile: &Profile{
			Experiences:   make([]Experience, 0),
			Memories:      make([]Memory, 0),
			Relationships: make([]Relationship, 0),
		},
		color: color,
	}
}

// Generate initiates the generation of the agent's response.
func (agent *Agent) Generate(ctx context.Context, system, user string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		agent.NextGemini(system, user, out)
	}()

	return out
}

/*
Save the agent state to a file.
*/
func (agent *Agent) Save() *AgentState {
	state := AgentState{
		ID:      agent.ID,
		History: agent.history,
		Profile: *agent.profile,
		Color:   agent.color,
	}

	json, err := json.Marshal(state)
	if err != nil {
		errnie.Error(err.Error())
		return nil
	}

	os.WriteFile(fmt.Sprintf("profiles/%s.json", agent.ID), json, 0644)
	return &state
}

/*
Load the agent state from a file.
*/
func (agent *Agent) Load() *AgentState {
	buf, err := os.ReadFile(fmt.Sprintf("profiles/%s.json", agent.ID))
	if err != nil {
		errnie.Error(err.Error())
		return nil
	}

	state := AgentState{}
	if err := json.Unmarshal(buf, &state); err != nil {
		errnie.Error(err.Error())
		return nil
	}

	agent.history = state.History
	agent.profile = &state.Profile
	agent.color = state.Color

	return &state
}

func (agent *Agent) NextGemini(system, user string, out chan string) {
	model := agent.conn.gemini.GenerativeModel("gemini-1.5-flash")
	iter := model.GenerateContentStream(agent.ctx, genai.Text(system+"\n\n"+user))

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			errnie.Error(err.Error())
			break
		}

		for _, candidate := range resp.Candidates {
			for _, part := range candidate.Content.Parts {
				out <- fmt.Sprintf("%s", part)
			}
		}
	}
}

// NextOpenAI handles the OpenAI API interaction.
func (agent *Agent) NextOpenAI(system, user string, out chan string) {
	request := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini, // Replace with the actual model you're using
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: system,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: user,
			},
		},
		Stream: true,
	}

	if agent.stream, agent.err = agent.conn.client.CreateChatCompletionStream(agent.ctx, request); agent.err != nil {
		errnie.Error(agent.err.Error())
		return
	}

	var response openai.ChatCompletionStreamResponse

	for {
		if response, agent.err = agent.stream.Recv(); agent.err != nil {
			if agent.err.Error() != "EOF" {
				errnie.Error(agent.err.Error())
			}

			break
		}

		chunk := response.Choices[0].Delta.Content

		if chunk == "" {
			continue
		}

		agent.history += chunk
		out <- chunk
	}
}

func (agent *Agent) ChatCompletion(system, user string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: system,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: user,
			},
		},
	}

	resp, err := agent.conn.client.CreateChatCompletion(agent.ctx, req)
	if err != nil {
		errnie.Error(err.Error())
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
