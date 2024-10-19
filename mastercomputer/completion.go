package mastercomputer

import (
	"context"
	"errors"
	"math"
	"time"

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
) (response *openai.ChatCompletion, err error) {
	maxRetries := 3
	baseDelay := 3 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, err := completion.executeWithStream(ctx, params)
		if err == nil {
			return response, nil
		}

		if attempt == maxRetries {
			errnie.Error(err)
			return response, err
		}

		delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay
		errnie.Info("retry attempt %d failed, retrying in %v: %v", attempt+1, delay, err)
		time.Sleep(delay)
	}

	return nil, errnie.Error(errors.New("max retries reached"))
}

func (completion *Completion) executeWithStream(
	ctx context.Context, params openai.ChatCompletionNewParams,
) (*openai.ChatCompletion, error) {
	response, err := completion.client.Chat.Completions.New(ctx, params)

	if errnie.Error(err) != nil {
		return nil, err
	}

	return response, nil
}
