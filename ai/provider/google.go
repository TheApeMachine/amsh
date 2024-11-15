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

// Helper function to convert messages to Google format
func (google *Google) convertMessages(messages []Message) []genai.Part {
	var parts []genai.Part
	for _, msg := range messages {
		content := genai.Content{
			Parts: []genai.Part{genai.Text(msg.Content)},
		}
		parts = append(parts, content.Parts...)
	}
	return parts
}

// Helper function to process stream responses
func (google *Google) processStream(iter *genai.GenerateContentResponseIterator, events chan<- Event) {
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

		for _, part := range resp.Candidates[0].Content.Parts {
			if text, ok := part.(genai.Text); ok {
				events <- Event{Type: EventToken, Content: string(text)}
			}
		}
	}
}

func (google *Google) Generate(ctx context.Context, params GenerationParams) <-chan Event {
	errnie.Info("generating with %s", google.model)
	events := make(chan Event, 64)

	go func() {
		defer close(events)

		parts := google.convertMessages(params.Messages)
		temp := float32(params.Temperature)

		model := google.client.GenerativeModel(google.model)
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(params.Messages[0].Content)},
		}
		model.SystemInstruction.Role = "system"
		model.Temperature = &temp

		iter := model.GenerateContentStream(ctx, parts...)
		google.processStream(iter, events)
	}()

	return events
}

// Add Configure method
func (google *Google) Configure(config map[string]interface{}) {
	// Google-specific configuration can be added here if needed
}
