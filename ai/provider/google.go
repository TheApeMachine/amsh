package provider

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/google/generative-ai-go/genai"
	"github.com/theapemachine/amsh/errnie"
	"google.golang.org/api/option"
)

type Google struct {
	client    *genai.Client
	model     string
	maxTokens int
}

func NewGoogle(apiKey string, model string) *Google {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Error("failed to create google client", "error", err)
		return nil
	}

	return &Google{
		client:    client,
		model:     model,
		maxTokens: 4096,
	}
}

func (g *Google) Generate(ctx context.Context, params GenerationParams, messages []Message) <-chan Event {
	errnie.Info("generating with google provider")
	events := make(chan Event, 64)

	go func() {
		defer close(events)

		// Convert messages to Google format
		var parts []genai.Part
		for _, msg := range messages[1:] {
			content := genai.Content{
				Parts: []genai.Part{genai.Text(msg.Content)},
			}
			parts = append(parts, content.Parts...) // Append the Parts from the Content
		}

		temp := float32(params.Temperature)

		model := g.client.GenerativeModel(g.model)
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(messages[0].Content)},
		}
		model.SystemInstruction.Role = "system"
		model.Temperature = &temp
		iter := model.GenerateContentStream(ctx, parts...)

		for {
			resp, err := iter.Next()
			if err != nil {
				if err.Error() == "iterator done" {
					events <- Event{Type: EventDone}
					return
				}
				events <- Event{Type: EventError, Error: err}
				return
			}

			// Process response parts
			for _, part := range resp.Candidates[0].Content.Parts {
				if text, ok := part.(genai.Text); ok {
					events <- Event{Type: EventToken, Content: string(text)}
				}
			}
		}
	}()

	return events
}

func (g *Google) GenerateSync(ctx context.Context, params GenerationParams, messages []Message) (string, error) {
	// Convert messages to Google format
	var parts []genai.Part
	for _, msg := range messages {
		content := genai.Content{
			Parts: []genai.Part{genai.Text(msg.Content)},
		}
		parts = append(parts, content.Parts...) // Same fix as in Generate method
	}

	temp := float32(params.Temperature)
	model := g.client.GenerativeModel(g.model)
	model.Temperature = &temp
	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", err
	}

	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result += string(text)
		}
	}

	return result, nil
}

// Add Configure method
func (google *Google) Configure(config map[string]interface{}) {
	// Google-specific configuration can be added here if needed
}
