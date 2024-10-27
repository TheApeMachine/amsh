package types

import (
	"context"
	"fmt"
)

// BaseTool provides a basic implementation of Tool interface
type BaseTool struct {
	name        string
	description string
	parameters  map[string]interface{}
	Handler     ToolHandler // Changed to exported field
}

// NewBaseTool creates a new tool with the given configuration
func NewBaseTool(name, description string, parameters map[string]interface{}, handler ToolHandler) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		parameters:  parameters,
		Handler:     handler,
	}
}

// Execute runs the tool with the given arguments
func (t *BaseTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	if t.Handler == nil {
		return "", fmt.Errorf("no handler registered for tool %s", t.name)
	}
	return t.Handler(ctx, args)
}

// GetSchema returns the tool's schema
func (t *BaseTool) GetSchema() ToolSchema {
	return ToolSchema{
		Name:        t.name,
		Description: t.description,
		Parameters:  t.parameters,
	}
}
