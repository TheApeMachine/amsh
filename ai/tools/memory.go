package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Memory struct {
	ToolName  string `json:"tool_name" jsonschema:"title=Tool Name,description=The name of the tool that must be 'memory',enum=memory"`
	Operation string `json:"operation" jsonschema:"title=Operation,description=The memory operation to perform,enum=get,enum=set,enum=delete,enum=query"`
	Key       string `json:"key" jsonschema:"title=Key,description=The unique identifier for storing or retrieving data,required"`
	Value     string `json:"value" jsonschema:"title=Value,description=The data to be stored (required for 'set' operation)"`
}

func NewMemory() *Memory {
	return &Memory{}
}

func (memory *Memory) GenerateSchema() string {
	schema := jsonschema.Reflect(&Memory{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

func (memory *Memory) Use(ctx context.Context, args map[string]any) string {
	return ""
}

// MemoryTool handles interactions with both vector and graph storage systems
func MemoryTool(ctx context.Context, args map[string]interface{}) (string, error) {
	// Validate required parameters
	store, ok := args["store"].(string)
	if !ok {
		return "", fmt.Errorf("store parameter is required")
	}

	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	data, ok := args["data"].(string)
	if !ok {
		return "", fmt.Errorf("data parameter is required")
	}

	// Create a new proxy with the provided parameters
	proxy := NewProxy(map[string]any{
		"store":     store,
		"operation": operation,
		"data":      data,
	})

	// Execute the operation and return the result
	result := proxy.Start()

	// If the result is already a string (error message), return it directly
	if _, err := json.Marshal(result); err != nil {
		return result, nil
	}

	// Otherwise, format the result as JSON string
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format result: %w", err)
	}

	return string(jsonResult), nil
}
