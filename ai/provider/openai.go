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
	).Yield(func(accumulator *twoface.Accumulator) {
		openAIMessages := make([]sdk.ChatCompletionMessageParamUnion, len(artifacts))

		defer close(accumulator.Out)

		errnie.Log("===START===")

		for i, msg := range artifacts {
			role := msg.Peek("role")
			payload := msg.Peek("payload")

			errnie.Log("OpenAI.Generate role %s payload %s", role, payload)

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

		errnie.Log("===END===")

		stream := openai.client.Chat.Completions.NewStreaming(context.Background(), sdk.ChatCompletionNewParams{
			Messages: sdk.F(openAIMessages),
			Model:    sdk.F(openai.model),
		})

		for stream.Next() {
			evt := stream.Current()
			if len(evt.Choices) > 0 && evt.Choices[0].Delta.Content != "" {
				response := data.New("openai", "assistant", openai.model, []byte(evt.Choices[0].Delta.Content))
				accumulator.Out <- response
			}
		}

		if err := stream.Err(); err != nil {
			errnie.Error(err)
		}
	}).Generate()
}
