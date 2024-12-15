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

func (openai *OpenAI) Read(p []byte) (n int, err error) {
	return openai.accumulator.Read(p)
}

func (openai *OpenAI) Write(p []byte) (n int, err error) {
	artifact := data.Empty()
	if err := artifact.Unmarshal(p); err != nil {
		errnie.Error(err)
		return 0, err
	}

	// Process the request in a goroutine
	go func() {
		defer func() {
			openai.accumulator.Close()
		}()

		messages := []sdk.ChatCompletionMessageParamUnion{
			openai.makeMessages(artifact),
		}

		stream := openai.client.Chat.Completions.NewStreaming(context.Background(), sdk.ChatCompletionNewParams{
			Messages: sdk.F(messages),
			Model:    sdk.F(openai.model),
		})

		for stream.Next() {
			if chunk := stream.Current(); len(chunk.Choices) > 0 {
				if delta := chunk.Choices[0].Delta; delta.Content != "" {
					responseArtifact := data.New(
						"openai", "assistant", "chunk", []byte(delta.Content),
					)

					buf := make([]byte, 1024)
					responseArtifact.Marshal(buf)

					if _, err := openai.accumulator.Write(buf); err != nil {
						errnie.Error(err)
						return
					}
				}
			}
		}

		if err := stream.Err(); err != nil {
			errnie.Error(err)
			return
		}
	}()

	return len(p), nil
}

func (openai *OpenAI) Close() error {
	return openai.accumulator.Close()
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

func (openai *OpenAI) Generate(ctx context.Context, params GenerationParams) <-chan Event {
	events := make(chan Event, 64)

	go func() {
		defer close(events)

		openAIMessages := make([]sdk.ChatCompletionMessageParamUnion, len(params.Messages))
		for i, msg := range params.Messages {
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
