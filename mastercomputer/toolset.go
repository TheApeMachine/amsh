package mastercomputer

import (
	"github.com/openai/openai-go"
)

// Toolset represents a set of tools available to a worker.
type Toolset struct {
	tools []openai.ChatCompletionToolParam
}

// makeTool reduces boilerplate for creating a tool.
func makeTool(name, description string, schema openai.FunctionParameters) openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String(name),
			Description: openai.String(description),
			Parameters:  openai.F(schema),
		}),
	}
}

// toolsMap is a map of tool names to tool definitions.
var toolsMap = map[string]openai.ChatCompletionToolParam{
	"publish_message": makeTool(
		"publish_message",
		"Publish a message to a topic channel. You must be subscribed to the channel to publish to it.",
		PublishMessageToolSchema(),
	),
	"worker": makeTool(
		"worker",
		"Create any type of worker by providing prompts and tools.",
		WorkerToolSchema(),
	),
	// Add other tools as necessary...
}

// NewToolset returns a new Toolset for the given key.
func NewToolset(key string) *Toolset {
	toolsets := map[string][]string{
		"reasoning": {"publish_message", "worker"},
		"messaging": {"publish_message"},
		// Add other toolsets as necessary...
	}

	if tools, exists := toolsets[key]; exists {
		var selectedTools []openai.ChatCompletionToolParam
		for _, toolKey := range tools {
			if tool, ok := toolsMap[toolKey]; ok {
				selectedTools = append(selectedTools, tool)
			}
		}
		return &Toolset{tools: selectedTools}
	}

	return nil
}

// WorkerToolSchema defines the schema for the worker creation tool.
func WorkerToolSchema() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]interface{}{
			"system": map[string]string{
				"type":        "string",
				"description": "The system prompt",
			},
			"user": map[string]string{
				"type":        "string",
				"description": "The user prompt",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"reasoning", "messaging"},
				"description": "The response format the worker should use",
			},
		},
		"required": []string{"system", "user", "format"},
	}
}

// PublishMessageToolSchema defines the schema for the publish_message tool.
func PublishMessageToolSchema() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]interface{}{
			"topic": map[string]string{
				"type":        "string",
				"description": "The topic channel you want to post to.",
			},
			"message": map[string]string{
				"type":        "string",
				"description": "The content of the message you want to post.",
			},
		},
		"required": []string{"topic", "message"},
	}
}
