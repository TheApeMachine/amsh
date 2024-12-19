package provider

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
	"github.com/theapemachine/errnie"
)

type Ollama struct {
	client *api.Client
	model  string
	system string
}

func NewOllama(model string) *Ollama {
	client := api.NewClient(
		&url.URL{Scheme: "http", Host: "localhost:11434"},
		&http.Client{},
	)

	return &Ollama{
		client: client,
		model:  model,
	}
}

func (o *Ollama) Generate(artifacts []*data.Artifact) <-chan *data.Artifact {
	return twoface.NewAccumulator(
		"ollama",
		"provider",
		"completion",
		artifacts...,
	).Yield(func(accumulator *twoface.Accumulator) {
		defer close(accumulator.Out)

		errnie.Log("===START===")
		prompt := o.convertToOllamaPrompt(artifacts)
		errnie.Log("===END===")

		req := &api.GenerateRequest{
			Model:  o.model,
			Prompt: prompt,
			Stream: utils.BoolPtr(true),
			Options: map[string]interface{}{
				"temperature": 0.7, // You might want to make this configurable
			},
		}

		respFunc := func(resp api.GenerateResponse) error {
			response := data.New("ollama", "assistant", o.model, []byte(resp.Response))
			accumulator.Out <- response
			return nil
		}

		if err := o.client.Generate(context.Background(), req, respFunc); err != nil {
			errnie.Error(err)
		}
	}).Generate()
}

func (o *Ollama) convertToOllamaPrompt(artifacts []*data.Artifact) string {
	var prompt string

	// Add system message if available
	if o.system != "" {
		prompt += "System: " + o.system + "\n"
	}

	for _, artifact := range artifacts {
		role := artifact.Peek("role")
		payload := artifact.Peek("payload")

		errnie.Log("Ollama.Generate role %s payload %s", role, payload)

		switch role {
		case "system":
			o.system = payload
		case "user":
			prompt += "Human: " + payload + "\n"
		case "assistant":
			prompt += "Assistant: " + payload + "\n"
		default:
			errnie.Warn("Ollama.Generate unknown_role %s", role)
		}
	}

	return prompt
}

func (o *Ollama) Configure(config map[string]interface{}) {
	if systemMsg, ok := config["system_message"].(string); ok {
		o.system = systemMsg
	}
	if model, ok := config["model"].(string); ok {
		o.model = model
	}
}
