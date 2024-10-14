package mastercomputer

import (
	"context"
	"errors"

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
	errnie.Trace()

	return &Completion{
		ctx:    ctx,
		client: openai.NewClient(),
	}
}

func (completion *Completion) Execute(
	ctx context.Context, params openai.ChatCompletionNewParams,
) (*openai.ChatCompletion, error) {
	response, err := completion.client.Chat.Completions.New(ctx, params)

	if err != nil {
		return nil, errnie.Error(err)
	}

	if response == nil {
		return nil, errnie.Error(errors.New("received nil response from OpenAI"))
	}

	return response, nil
}
