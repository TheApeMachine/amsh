package provider

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/theapemachine/amsh/errnie"
)

type NVIDIA struct {
	client    *openai.Client
	model     string
	maxTokens int
}

func NewNVIDIA(apiKey string, model string) *NVIDIA {
	return &NVIDIA{
		client: openai.NewClient(
			option.WithBaseURL("https://integrate.api.nvidia.com/v1"),
			option.WithAPIKey(apiKey),
		),
		model:     model,
		maxTokens: 4096,
	}
}

func (o *NVIDIA) Generate(ctx context.Context, params GenerationParams) <-chan Event {
	errnie.Info("generating with " + o.model)
	events := make(chan Event, 64)

	go func() {
		defer close(events)

		openAIMessages := make([]openai.ChatCompletionMessageParamUnion, len(params.Messages))
		for i, msg := range params.Messages {
			switch msg.Role {
			case "user":
				openAIMessages[i] = openai.UserMessage(msg.Content)
			case "assistant":
				openAIMessages[i] = openai.AssistantMessage(msg.Content)
			case "system":
				openAIMessages[i] = openai.SystemMessage(msg.Content)
			case "tool":
				openAIMessages[i] = openai.ToolMessage(msg.Name, msg.Content)
			}
		}

		stream := o.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Messages:    openai.F(openAIMessages),
			Model:       openai.F(o.model),
			Temperature: openai.F(params.Temperature),
			TopP:        openai.F(params.TopP),
		})

		for stream.Next() {
			evt := stream.Current()
			if len(evt.Choices) > 0 {
				events <- Event{Type: EventToken, Content: evt.Choices[0].Delta.Content}
			}
		}

		if err := stream.Err(); err != nil {
			events <- Event{Type: EventError, Error: err}
			return
		}

		events <- Event{Type: EventDone}
	}()

	return events
}

func (o *NVIDIA) Configure(config map[string]interface{}) {
	// OpenAI-specific configuration can be added here if needed
}
