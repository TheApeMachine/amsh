package provider

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Ollama struct {
	Model string
}

func NewOllama(model string) *Ollama {
	return &Ollama{Model: model}
}

func (ollama *Ollama) Generate(ctx context.Context, params GenerationParams) <-chan Event {
	errnie.Info("generating with %s", ollama.Model)
	eventChan := make(chan Event)

	go func() {
		defer close(eventChan)

		client := api.NewClient(
			&url.URL{Scheme: "http", Host: "localhost:11434"},
			&http.Client{},
		)

		// Convert messages into a prompt
		prompt := params.Messages[len(params.Messages)-1].Content

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

func (ollama *Ollama) Configure(config map[string]interface{}) {
	// Handle any Ollama-specific configuration here
	// For example, if we need to update the model:
	if model, ok := config["model"].(string); ok {
		ollama.Model = model
	}
}
