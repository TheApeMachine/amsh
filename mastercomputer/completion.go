package mastercomputer

import (
	"context"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/format"
	"github.com/theapemachine/amsh/errnie"
)

func GenerateSchema[T any]() interface{} {
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

func GetParams(system, user string, toolset openai.ChatCompletionToolParam) openai.ChatCompletionNewParams {
	return openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
			openai.UserMessage(user),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        openai.F("reasoning"),
					Description: openai.F("Available reasoning strategies"),
					Schema:      openai.F(GenerateSchema[format.Reasoning]()),
					Strict:      openai.Bool(false),
				}),
			},
		),
		Tools: openai.F([]openai.ChatCompletionToolParam{
			toolset,
		}),
		Seed:  openai.Int(0),
		Model: openai.F(openai.ChatModelGPT4oMini),
	}
}

type Completion struct {
	ctx      context.Context
	conn     *ai.Conn
	response openai.ChatCompletion
	err      error
}

func NewCompletion(ctx context.Context) *Completion {
	return &Completion{
		ctx:  ctx,
		conn: ai.NewConn(),
	}
}

func (completion *Completion) Execute(ctx context.Context, params openai.ChatCompletionNewParams) openai.ChatCompletion {
	errnie.Trace()
	client := openai.NewClient()
	var response *openai.ChatCompletion

	if response, completion.err = client.Chat.Completions.New(ctx, params); completion.err != nil {
		errnie.Error(completion.err)
	}

	return *response
}
