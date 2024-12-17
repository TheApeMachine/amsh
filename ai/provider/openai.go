package provider

import (
	"context"

	sdk "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/errnie"
)

type OpenAI struct {
	client *sdk.Client
	model  string
}

func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{
		client: sdk.NewClient(
			option.WithAPIKey(apiKey),
		),
		model: model,
	}
}

func (openai *OpenAI) Generate(artifacts []*data.Artifact) <-chan *data.Artifact {
	return twoface.NewAccumulator(
		"openai",
		"provider",
		"completion",
		artifacts...,
	).Yield(func(artifacts []*data.Artifact, out chan<- *data.Artifact) {
		openAIMessages := make([]sdk.ChatCompletionMessageParamUnion, len(artifacts))
		errnie.Trace("OpenAI.Generate", "artifacts_count", len(artifacts))

		defer close(out)

		for i, msg := range artifacts {
			role := msg.Peek("role")
			payload := msg.Peek("payload")
			errnie.Trace("OpenAI.Generate", "processing_message", i, "role", role, "payload", payload)

			switch role {
			case "user":
				openAIMessages[i] = sdk.UserMessage(payload)
			case "assistant":
				openAIMessages[i] = sdk.AssistantMessage(payload)
			case "system":
				openAIMessages[i] = sdk.SystemMessage(payload)
			case "tool":
				openAIMessages[i] = sdk.ToolMessage(msg.Peek("name"), payload)
			default:
				errnie.Warn("OpenAI.Generate unknown_role %s", role)
			}
		}

		stream := openai.client.Chat.Completions.NewStreaming(context.Background(), sdk.ChatCompletionNewParams{
			Messages: sdk.F(openAIMessages),
			Model:    sdk.F(openai.model),
		})

		for stream.Next() {
			evt := stream.Current()
			if len(evt.Choices) > 0 && evt.Choices[0].Delta.Content != "" {
				response := data.New("test", "assistant", "payload", []byte(evt.Choices[0].Delta.Content))
				errnie.Trace("OpenAI.Generate", "generated_response", string(response.Peek("payload")))
				out <- response
			}
		}

		if err := stream.Err(); err != nil {
			errnie.Error(err)
		}
	}).Generate()
}
