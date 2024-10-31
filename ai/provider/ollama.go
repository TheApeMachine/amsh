package provider

import (
	"context"
	"net/http"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/ollama/ollama/api"
	"github.com/theapemachine/amsh/utils"
)

type Ollama struct {
	Model string
}

func NewOllama(model string) *Ollama {
	return &Ollama{Model: model}
}

func (ollama *Ollama) Generate(ctx context.Context, params GenerationParams, messages []Message) <-chan Event {
	log.Info("generating with", "model", ollama.Model)
	eventChan := make(chan Event)

	go func() {
		defer close(eventChan)

		client := api.NewClient(
			&url.URL{Scheme: "http", Host: "localhost:11434"},
			&http.Client{},
		)

		// Convert messages into a prompt
		prompt := messages[len(messages)-1].Content

		req := &api.GenerateRequest{
			Model:  ollama.Model,
			Prompt: prompt,
			Stream: utils.BoolPtr(true),
			// Convert temperature and other params if needed
			Options: map[string]interface{}{
				"temperature": params.Temperature,
			},
		}

		respFunc := func(resp api.GenerateResponse) error {
			eventChan <- Event{
				Type:    EventToken,
				Content: resp.Response,
			}
			return nil
		}

		if err := client.Generate(ctx, req, respFunc); err != nil {
			eventChan <- Event{Type: EventError, Error: err}
			return
		}

		eventChan <- Event{Type: EventDone}
	}()

	return eventChan
}

func (ollama *Ollama) GenerateSync(ctx context.Context, params GenerationParams, messages []Message) (string, error) {
	var result string

	for event := range ollama.Generate(ctx, params, messages) {
		switch event.Type {
		case EventToken:
			result += event.Content
		case EventError:
			return "", event.Error
		}
	}

	return result, nil
}

func (ollama *Ollama) Configure(config map[string]interface{}) {
	// Handle any Ollama-specific configuration here
	// For example, if we need to update the model:
	if model, ok := config["model"].(string); ok {
		ollama.Model = model
	}
}
