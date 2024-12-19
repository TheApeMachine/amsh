package provider

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/errnie"
	"google.golang.org/api/option"
)

type Google struct {
	client    *genai.Client
	model     string
	maxTokens int
	system    string
}

func NewGoogle(apiKey string, model string) *Google {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		errnie.Error(err)
		return nil
	}

	return &Google{
		client:    client,
		model:     model,
		maxTokens: 4096,
	}
}

func (g *Google) Generate(artifacts []*data.Artifact) <-chan *data.Artifact {
	return twoface.NewAccumulator(
		"google",
		"provider",
		"completion",
		artifacts...,
	).Yield(func(accumulator *twoface.Accumulator) {
		defer close(accumulator.Out)

		errnie.Log("===START===")
		parts := g.convertToGoogleParts(artifacts)
		errnie.Log("===END===")

		model := g.client.GenerativeModel(g.model)
		temp := float32(0.7) // You might want to make this configurable

		// Set system message if available
		if g.system != "" {
			model.SystemInstruction = &genai.Content{
				Parts: []genai.Part{genai.Text(g.system)},
				Role:  "system",
			}
		}
		model.Temperature = &temp

		iter := model.GenerateContentStream(context.Background(), parts...)

		for {
			resp, err := iter.Next()
			if err != nil {
				if err.Error() == "iterator done" {
					break
				}
				errnie.Error(err)
				break
			}

			for _, part := range resp.Candidates[0].Content.Parts {
				if text, ok := part.(genai.Text); ok {
					response := data.New("google", "assistant", g.model, []byte(text))
					accumulator.Out <- response
				}
			}
		}
	}).Generate()
}

func (g *Google) convertToGoogleParts(artifacts []*data.Artifact) []genai.Part {
	var parts []genai.Part

	for _, artifact := range artifacts {
		role := artifact.Peek("role")
		payload := artifact.Peek("payload")

		errnie.Log("Google.Generate role %s payload %s", role, payload)

		// Skip system messages as they're handled separately
		if role == "system" {
			g.system = payload
			continue
		}

		content := genai.Content{
			Parts: []genai.Part{genai.Text(payload)},
		}
		parts = append(parts, content.Parts...)
	}

	return parts
}

func (g *Google) Configure(config map[string]interface{}) {
	if systemMsg, ok := config["system_message"].(string); ok {
		g.system = systemMsg
	}
}
