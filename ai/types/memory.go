package types

import (
	"context"
	"encoding/json"
	"fmt"
)

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
