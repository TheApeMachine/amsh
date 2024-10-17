package ai

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/twoface"
)

type Tool interface {
	Call(args map[string]any, owner twoface.Process) (string, error)
	Ctx() context.Context
	ID() string
	Name() string
	Schema() openai.ChatCompletionToolParam
}

// MakeTool reduces boilerplate for creating a tool.
func MakeTool(name, description string, schema openai.FunctionParameters) openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String(name),
			Description: openai.String(description),
			Parameters:  openai.F(schema),
		}),
	}
}
