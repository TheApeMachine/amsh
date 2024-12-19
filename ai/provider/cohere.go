package provider

import (
	"context"
	"io"

	cohereCore "github.com/cohere-ai/cohere-go/v2"
	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/errnie"
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
func (cohere *Cohere) Generate(artifacts []*data.Artifact) <-chan *data.Artifact {
	return twoface.NewAccumulator(
		"cohere",
		"provider",
		"completion",
		artifacts...,
	).Yield(func(accumulator *twoface.Accumulator) {
		defer close(accumulator.Out)

		errnie.Log("===START===")
		prompt := cohere.convertMessagesToCoherePrompt(artifacts)
		errnie.Log("===END===")

		stream, err := cohere.client.ChatStream(context.Background(), &cohereCore.ChatStreamRequest{
			Message: prompt,
			Model:   &cohere.model,
		})
		if err != nil {
			errnie.Error(err)
			return
		}
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errnie.Error(err)
				break
			}

			if resp.TextGeneration != nil {
				response := data.New("cohere", "assistant", cohere.model, []byte(resp.TextGeneration.Text))
				accumulator.Out <- response
			}
		}
	}).Generate()
}

// convertMessagesToCoherePrompt converts the message array into a string prompt
// that Cohere can understand
func (cohere *Cohere) convertMessagesToCoherePrompt(artifacts []*data.Artifact) string {
	var prompt string
	for _, artifact := range artifacts {
		switch artifact.Peek("role") {
		case "system":
			prompt += "System: " + artifact.Peek("payload") + "\n"
		case "user":
			prompt += "Human: " + artifact.Peek("payload") + "\n"
		case "assistant":
			prompt += "Assistant: " + artifact.Peek("payload") + "\n"
		}
	}
	return prompt
}
