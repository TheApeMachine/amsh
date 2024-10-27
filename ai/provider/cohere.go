package provider

import (
	"context"
	"io"

	cohere "github.com/cohere-ai/cohere-go/v2"
	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
)

type Cohere struct {
	client    *cohereclient.Client
	model     string
	maxTokens int
}

func NewCohere(apiKey string, model string) (*Cohere, error) {
	client := cohereclient.NewClient(cohereclient.WithToken(apiKey))

	return &Cohere{
		client:    client,
		model:     model,
		maxTokens: 2000,
	}, nil
}

func (c *Cohere) Generate(ctx context.Context, messages []Message) <-chan Event {
	events := make(chan Event)

	go func() {
		defer close(events)

		prompt := convertMessagesToCoherePrompt(messages)

		stream, err := c.client.ChatStream(ctx, &cohere.ChatStreamRequest{
			Message: prompt,
		})
		if err != nil {
			events <- Event{Type: EventError, Error: err}
			return
		}
		defer stream.Close()

		for {
			message, err := stream.Recv()
			if err == io.EOF {
				events <- Event{Type: EventDone}
				return
			}
			if err != nil {
				events <- Event{Type: EventError, Error: err}
				return
			}

			events <- Event{Type: EventToken, Content: message.TextGeneration.Text}
		}
	}()

	return events
}

func (c *Cohere) GenerateSync(ctx context.Context, messages []Message) (string, error) {
	prompt := convertMessagesToCoherePrompt(messages)

	resp, err := c.client.Chat(ctx, &cohere.ChatRequest{
		Message: prompt,
	})
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}

// convertMessagesToCoherePrompt converts the message array into a string prompt
// that Cohere can understand
func convertMessagesToCoherePrompt(messages []Message) string {
	var prompt string
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			prompt += "System: " + msg.Content + "\n"
		case "user":
			prompt += "Human: " + msg.Content + "\n"
		case "assistant":
			prompt += "Assistant: " + msg.Content + "\n"
		}
	}
	return prompt
}
