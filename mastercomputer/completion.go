package mastercomputer

import (
	"context"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/errnie"
)

func GenerateSchema[T any]() interface{} {
	errnie.Trace()

	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

type Completion struct {
	ctx    context.Context
	client *openai.Client
}

func NewCompletion(ctx context.Context) *Completion {
	errnie.Info("new completion")
	return &Completion{
		ctx:    ctx,
		client: openai.NewClient(),
	}
}

func (completion *Completion) Execute(
	ctx context.Context, params openai.ChatCompletionNewParams,
) (response openai.ChatCompletionAccumulator, err error) {
	errnie.Info("executing completion")

	stream := completion.client.Chat.Completions.NewStreaming(ctx, params)
	acc := openai.ChatCompletionAccumulator{}

	for stream.Next() {
		chunk := stream.Current()
		println(chunk.Choices[0].Delta.Content)
		acc.AddChunk(chunk)

		// When this fires, the current chunk value will not contain content data
		if content, ok := acc.JustFinishedContent(); ok {
			println("Content stream finished:", content)
			println()
		}

		if tool, ok := acc.JustFinishedToolCall(); ok {
			println("Tool call stream finished:", tool.Index, tool.Name, tool.Arguments)
			println()
		}

		if refusal, ok := acc.JustFinishedRefusal(); ok {
			println("Refusal stream finished:", refusal)
			println()
		}
	}

	if err := stream.Err(); err != nil {
		errnie.Error(err)
	}

	return acc, nil
}
