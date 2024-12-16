package provider

import (
	"context"

	sdk "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/errnie"
)

// Define a threshold for buffer processing
const bufferThreshold = 1024 * 4 // 4KB

type OpenAI struct {
	accumulator *twoface.Accumulator
	client      *sdk.Client
	model       string
}

/*
NewOpenAI creates an OpenAI provider, which is configurable with an endpoint,
api key, and model, given that a lot of other providers use the OpenAI API
specifications.
*/
func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{
		accumulator: twoface.NewAccumulator(),
		client: sdk.NewClient(
			option.WithAPIKey(apiKey),
		),
		model: model,
	}
}

func (openai *OpenAI) makeMessages(artifact *data.Artifact) sdk.ChatCompletionMessageParamUnion {
	switch artifact.Peek("role") {
	case "user":
		return sdk.UserMessage(artifact.Peek("payload"))
	case "assistant":
		return sdk.AssistantMessage(artifact.Peek("payload"))
	case "system":
		return sdk.SystemMessage(artifact.Peek("payload"))
	}

	return nil
}

func (openai *OpenAI) Generate(artifact *data.Artifact) <-chan *data.Artifact {
	events := make(chan *data.Artifact, 64)

	go func() {
		defer close(events)

		openAIMessages := make([]sdk.ChatCompletionMessageParamUnion, len(artifact.Messages))
		for i, msg := range artifact.Messages {
			switch msg.Role {
			case "user":
				openAIMessages[i] = sdk.UserMessage(msg.Content)
			case "assistant":
				openAIMessages[i] = sdk.AssistantMessage(msg.Content)
			case "system":
				openAIMessages[i] = sdk.SystemMessage(msg.Content)
			case "tool":
				openAIMessages[i] = sdk.ToolMessage(msg.Name, msg.Content)
			}
		}

		stream := openai.client.Chat.Completions.NewStreaming(ctx, sdk.ChatCompletionNewParams{
			Messages: sdk.F(openAIMessages),
			Model:    sdk.F(openai.model),
			// Temperature:      openai.F(params.Temperature),
			// FrequencyPenalty: openai.F(params.FrequencyPenalty),
			// PresencePenalty:  openai.F(params.PresencePenalty),
		})

		for stream.Next() {
			evt := stream.Current()
			if len(evt.Choices) > 0 {
				events <- Event{Type: EventToken, Content: evt.Choices[0].Delta.Content}
			}
		}

		if err := stream.Err(); err != nil {
			errnie.Error(err)
			events <- Event{Type: EventError, Error: err}
			return
		}

		events <- Event{Type: EventDone}
	}()

	return events
}

func (openai *OpenAI) Configure(config map[string]interface{}) {
	// OpenAI-specific configuration can be added here if needed
}
