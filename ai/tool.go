package ai

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/ai/types"
)

// Tool interface is now defined in types package

// ToolHandler is a function that implements the actual tool logic
type ToolHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// BaseTool provides a basic implementation of Tool interface
type BaseTool struct {
	name        string
	description string
	parameters  map[string]interface{}
	handler     ToolHandler
}

// NewBaseTool creates a new tool with the given configuration
func NewBaseTool(name, description string, parameters map[string]interface{}, handler ToolHandler) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		parameters:  parameters,
		handler:     handler,
	}
}

// Execute runs the tool with the given arguments
func (t *BaseTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	if t.handler == nil {
		return "", fmt.Errorf("no handler registered for tool %s", t.name)
	}
	return t.handler(ctx, args)
}

// GetSchema returns the tool's schema
func (t *BaseTool) GetSchema() types.ToolSchema {
	return types.ToolSchema{
		Name:        t.name,
		Description: t.description,
		Parameters:  t.parameters,
	}
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
