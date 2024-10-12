package mastercomputer

import (
	"context"
	"errors"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
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

func GetParams(system, user string, toolset []openai.ChatCompletionToolParam) openai.ChatCompletionNewParams {
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
					Schema:      openai.F(GenerateSchema[format.Strategy]()),
					Strict:      openai.Bool(false),
				}),
			},
		),
		Tools: openai.F(toolset),
		Seed:  openai.Int(0),
		Model: openai.F(openai.ChatModelGPT4oMini),
	}
}

type Completion struct {
	ctx    context.Context
	client *openai.Client
	err    error
}

func NewCompletion(ctx context.Context) *Completion {
	return &Completion{
		ctx:    ctx,
		client: openai.NewClient(),
	}
}

func (completion *Completion) Execute(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	errnie.Trace()

	response, err := completion.client.Chat.Completions.New(ctx, params)
	if err != nil {
		errnie.Error(err)
		return nil, err
	}

	if response == nil {
		err = errors.New("no response from OpenAI")
		errnie.Error(err)
		return nil, err
	}

	return response, nil
}
