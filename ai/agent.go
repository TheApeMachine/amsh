package ai

import (
	"context"
	"fmt"
	"io"

	"github.com/google/generative-ai-go/genai"
	openai "github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/errnie"
	"golang.org/x/exp/rand"
	"google.golang.org/api/iterator"
)

/*
Agent is a configurable wrapper around an AI model.
*/
type Agent struct {
	ctx              context.Context
	conn             *Conn
	ID               string
	Type             string
	Scope            string
	Responsibilities string
	Color            string
}

/*
NewAgent initializes the agent with an ID.
*/
func NewAgent(
	ctx context.Context,
	conn *Conn,
	ID string,
	Type string,
	scope string,
	responsibilities string,
	color string,
) *Agent {
	return &Agent{
		ctx:              ctx,
		conn:             conn,
		ID:               ID,
		Type:             Type,
		Scope:            scope,
		Responsibilities: responsibilities,
		Color:            color,
	}
}

/*
Generate initiates the generation of the agent's response.
*/
func (agent *Agent) Generate(ctx context.Context, system, user string) <-chan string {
	out := make(chan string)

	// Generate a random int between 0 and 1.
	selected := rand.Intn(2)

	errnie.Debug("SYSTEM:\n\n%s\n\n---\n\nUSER:\n\n%s---\n\n", system, user)

	go func() {
		defer close(out)

		if selected == 0 {
			agent.NextOpenAI(system, user, out)
		} else {
			agent.NextGemini(system, user, out)
		}
	}()

	return out
}

/*
NextOpenAI handles the OpenAI API interaction.
*/
func (agent *Agent) NextOpenAI(system, user string, out chan string) {
	request := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		Stream: true,
	}

	stream, err := agent.conn.client.CreateChatCompletionStream(agent.ctx, request)
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
			out <- chunk
		}
	}
}

/*
ChatCompletion generates a single, complete response from the OpenAI API.
*/
func (agent *Agent) ChatCompletion(system, user string) (string, error) {
	request := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	}

	response, err := agent.conn.client.CreateChatCompletion(agent.ctx, request)
	if err != nil {
		errnie.Error(err.Error())
		return "", err
	}

	content := response.Choices[0].Message.Content
	return content, nil
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
