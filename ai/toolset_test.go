package ai

import "github.com/theapemachine/amsh/ai/types"

// NewTestToolset creates a toolset with basic tools for testing
func NewTestToolset() *Toolset {
	toolset := &Toolset{
		tools: make(map[string]types.Tool),
	}

	// Register memory tool
	memoryTool := NewBaseTool(
		"memory",
		"Access and manipulate the memory storage systems",
		map[string]interface{}{
			"store": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"vector", "graph"},
				"description": "The type of storage to use",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"store", "retrieve", "search", "delete"},
				"description": "The operation to perform",
			},
			"data": map[string]interface{}{
				"type":        "string",
				"description": "The data to store or query parameters",
			},
		},
		types.MemoryTool,
	)
	toolset.tools["memory"] = memoryTool

	return toolset
}
