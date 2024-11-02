package provider

import (
	"context"
	"fmt"
	"io"

	cohere "github.com/cohere-ai/cohere-go/v2"
	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
	"github.com/theapemachine/amsh/errnie"
)

type Cohere struct {
	client    *cohereclient.Client
	model     string
	maxTokens int
}

func NewCohere(apiKey string, model string) *Cohere {
	// Create client with proper API key configuration
	client := cohereclient.NewClient(
		cohereclient.WithToken(apiKey),
	)

	// Validate that we have required parameters
	if apiKey == "" {
		return nil
	}
	if model == "" {
		model = "command" // Set default model if none specified
	}

	return &Cohere{
		client:    client,
		model:     model,
		maxTokens: 4096,
	}
}

func (c *Cohere) Generate(ctx context.Context, params GenerationParams, messages []Message) <-chan Event {
	errnie.Info("generating with cohere provider")
	events := make(chan Event, 64)

	go func() {
		defer close(events)

		// Check for nil client
		if c.client == nil {
			events <- Event{Type: EventError, Error: fmt.Errorf("cohere: client not initialized")}
			return
		}

		prompt := convertMessagesToCoherePrompt(messages)
		temp := params.Temperature

		stream, err := c.client.ChatStream(ctx, &cohere.ChatStreamRequest{
			Message:     prompt,
			Model:       &c.model,
			Temperature: &temp,
		})
		if err != nil {
			events <- Event{Type: EventError, Error: err}
			return
		}
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				events <- Event{Type: EventDone}
				return
			}
			if err != nil {
				events <- Event{Type: EventError, Error: err}
				return
			}

			// Check if response has text content
			if resp.TextGeneration != nil {
				events <- Event{Type: EventToken, Content: resp.TextGeneration.Text}
			}
		}
	}()

	return events
}

func (c *Cohere) GenerateSync(ctx context.Context, params GenerationParams, messages []Message) (string, error) {
	prompt := convertMessagesToCoherePrompt(messages)

	resp, err := c.client.Chat(ctx, &cohere.ChatRequest{
		Message:     prompt,
		Model:       &c.model,
		Temperature: &params.Temperature,
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

// Add Configure method
func (cohere *Cohere) Configure(config map[string]interface{}) {
	// Cohere-specific configuration can be added here if needed
}
