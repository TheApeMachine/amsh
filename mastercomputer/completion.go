package mastercomputer

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/format"
	"github.com/theapemachine/amsh/errnie"
)

type Completion struct {
	ctx      context.Context
	conn     *ai.Conn
	response openai.ChatCompletionResponse
	err      error
}

func NewCompletion(ctx context.Context) *Completion {
	return &Completion{
		ctx:  ctx,
		conn: ai.NewConn(),
	}
}

func (completion *Completion) Execute(system, user, toolset string, format format.Response) openai.ChatCompletionResponse {
	errnie.Trace()

	if completion.response, completion.err = completion.conn.Request(completion.ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: system,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: user,
			},
		},
		Tools: NewToolSet(completion.ctx).Tools(toolset),
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   format.Name(),
				Schema: format.Schema(),
			},
		},
	}); errnie.Error(completion.err) != nil {
		fmt.Println("Error initializing stream:", completion.err)
		return openai.ChatCompletionResponse{}
	}

	return completion.response
}
