package mastercomputer

import (
	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/errnie"
)

type Toolset struct {
	tools []openai.ChatCompletionToolParam
}

/*
makeTool reduces some of the boilerplate code for creating a tool.
*/
func makeTool(name, description string, schema openai.FunctionParameters) openai.ChatCompletionToolParam {
	errnie.Trace()
	return openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String(name),
			Description: openai.String(description),
			Parameters:  openai.F(schema),
		}),
	}
}

/*
toolsMap is a map of tool names to tool definitions.
*/
var toolsMap = map[string]openai.ChatCompletionToolParam{
	"publish_message": makeTool(
		"publish_message",
		"Publish a message to a topic channel. You must be subscribed to the channel to publish to it.",
		PublishMessageToolSchema(),
	),
	"worker": makeTool(
		"worker",
		"Create any type of worker, by providing a system prompt, a user prompt, and a toolset.",
		WorkerToolSchema(),
	),
	"environment": makeTool(
		"environment",
		"Request a new environment.",
		EnvironmentToolSchema(),
	),
}

/*
NewToolset returns a new Toolset for the given key.
*/
func NewToolset(key string) *Toolset {
	errnie.Trace()

	toolsets := map[string][]string{
		"reasoning": {"publish_message", "worker"},
		"execution": {"publish_message", "environment"},
		"routing":   {"publish_message"},
		"none":      {},
	}

	if tools, exists := toolsets[key]; exists {
		var selectedTools []openai.ChatCompletionToolParam
		for _, toolKey := range tools {
			selectedTools = append(selectedTools, toolsMap[toolKey])
		}
		return &Toolset{tools: selectedTools}
	}

	return nil
}

/*
WorkerToolSchema is the schema for the worker tool, which enables
a worker to create other workers.
*/
func WorkerToolSchema() openai.FunctionParameters {
	errnie.Trace()

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
			"format": map[string]any{
				"type":        "string",
				"enum":        []string{"reasoning", "execution", "routing", "freeform"},
				"description": "The response format the worker should use",
			},
			"strict": map[string]any{
				"type":        "boolean",
				"description": "Whether the response format should be strict",
			},
			"toolset": map[string]any{
				"type":        "string",
				"enum":        []string{"core", "extended", "full"},
				"description": "The toolset the worker should use",
			},
		},
		"required": []string{"system", "user", "format", "strict", "toolset"},
	}
}

/*
PublishMessageToolSchema is the schema for the publish_message tool, which
enables a worker to publish a message to a topic channel.
*/
func PublishMessageToolSchema() openai.FunctionParameters {
	errnie.Trace()

	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]interface{}{
			"topic": map[string]string{
				"type":        "string",
				"description": "The topic channel you want to post to. You must be subscribed to the channel to post to it.",
			},
			"message": map[string]string{
				"type":        "string",
				"description": "The content of the message you want to post.",
			},
			"type": map[string]any{
				"type":        "string",
				"enum":        []string{"general", "task"},
				"description": "The type of message you want to post.",
			},
			"priority": map[string]any{
				"type":        "integer",
				"enum":        []string{"low", "normal", "high"},
				"description": "The priority of the message you want to post.",
			},
		},
		"required": []string{"topic", "message", "type", "priority"},
	}
}

/*
EnvironmentToolSchema is the schema for the environment tool, which
enables a worker to request a new environment.
*/
func EnvironmentToolSchema() openai.FunctionParameters {
	errnie.Trace()

	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]interface{}{
			"justification": map[string]string{
				"type":        "string",
				"description": "A valid reason why you are requesting a new environment.",
			},
		},
		"required": []string{"justification"},
	}
}
