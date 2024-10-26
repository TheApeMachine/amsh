package mastercomputer

import (
	"context"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/berrt"
	"github.com/theapemachine/amsh/errnie"
)

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

type Completion struct {
	ctx    context.Context
	cancel context.CancelFunc
	client *openai.Client
}

func NewCompletion(pctx context.Context) *Completion {
	errnie.Trace()
	ctx, cancel := context.WithCancel(pctx)

	return &Completion{
		ctx:    ctx,
		cancel: cancel,
		client: openai.NewClient(),
	}
}

func (completion *Completion) Execute(params openai.ChatCompletionNewParams) (response *openai.ChatCompletion) {
	errnie.Trace()
	var err error

	if response, err = completion.client.Chat.Completions.New(completion.ctx, params); err != nil {
		berrt.Error("OpenAI API Error", err)
		return nil
	}

	return response
}
