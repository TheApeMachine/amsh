package provider

import (
	"context"

	sdk "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/errnie"
)

// Define a threshold for buffer processing
const bufferThreshold = 1024 * 4 // 4KB

type OpenAI struct {
	client *sdk.Client
	model  string
}

/*
NewOpenAI creates an OpenAI provider, which is configurable with an endpoint,
api key, and model, given that a lot of other providers use the OpenAI API
specifications.
*/
func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{
		client: sdk.NewClient(
			option.WithAPIKey(apiKey),
		),
		model: model,
	}
}

func (openai *OpenAI) Generate(artifacts []*data.Artifact) <-chan *data.Artifact {
	out := make(chan *data.Artifact)

	go func() {
		defer close(out)

		openAIMessages := make([]sdk.ChatCompletionMessageParamUnion, len(artifacts))

		errnie.Trace("OpenAI.Generate", "artifacts_count", len(artifacts))

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
				errnie.Warn("OpenAI.Generate", "unknown_role", role)
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
	}()
	errnie.Trace("OpenAI.Generate", "status", "Generator started")
	return out
}
