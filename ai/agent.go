package ai

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

type Memory struct {
	Timestamp time.Time `json:"timestamp"`
	Scene     string    `json:"scene"`
	Action    string    `json:"action"`
	Content   string    `json:"content"`
}

type Profile struct {
	Name      string   `json:"name"`
	Backstory string   `json:"backstory"`
	Resume    string   `json:"resume"`
	Memories  []Memory `json:"memories"`
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

// Agent is a configurable wrapper around an AI model.
type Agent struct {
	ctx     context.Context
	conn    *Conn
	ID      string
	history string
	profile *Profile
	stream  *openai.ChatCompletionStream
	err     error
}

// NewAgent initializes the agent with an ID.
func NewAgent(ctx context.Context, conn *Conn, ID string) *Agent {
	return &Agent{
		ctx:  ctx,
		conn: conn,
		ID:   ID,
		profile: &Profile{
			Memories: make([]Memory, 0),
		},
	}
}

// Generate initiates the generation of the agent's response.
func (agent *Agent) Generate(ctx context.Context, system, user string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		agent.NextOpenAI(system, user, out)
	}()

	return out
}

/*
UpdateProfile uses an LLM to analyze the agent's history and update the profile.
*/
func (agent *Agent) UpdateProfile() {
	system := viper.GetViper().GetString("ai.crew.extractor.system")
	user := viper.GetViper().GetString("ai.crew.extractor.user")

	user = strings.ReplaceAll(user, "<{profile}>", agent.profile.String())
	user = strings.ReplaceAll(user, "<{history}>", agent.history)

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
		return
	}

	agent.profile.Unmarshal(resp.Choices[0].Message.Content)
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
