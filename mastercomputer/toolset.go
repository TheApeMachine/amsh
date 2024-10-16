package mastercomputer

import (
	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/integration/boards"
	"github.com/theapemachine/amsh/integration/trengo"
)

// Toolset represents a set of tools available to a worker.
type Toolset struct {
	tools []openai.ChatCompletionToolParam
}

// toolsMap is a map of tool names to tool definitions.
var toolsMap = map[string]openai.ChatCompletionToolParam{
	"publish_message": ai.MakeTool(
		"publish_message",
		"Publish a message to a topic channel. You must be subscribed to the channel to publish to it.",
		PublishMessageToolSchema(),
	),
	"worker": ai.MakeTool(
		"worker",
		"Create any type of worker by providing prompts and tools.",
		WorkerToolSchema(),
	),
}

func init() {
	// Boards
	toolsMap["create_workitem"] = boards.NewTools().GetSchemas()["create_ticket"]
	toolsMap["search_workitems"] = boards.NewTools().GetSchemas()["search_tickets"]
	toolsMap["get_workitem"] = boards.NewTools().GetSchemas()["get_ticket"]
	// Trengo
	toolsMap["search_tickets"] = trengo.NewTools().GetSchemas()["search_tickets"]
	toolsMap["get_ticket"] = trengo.NewTools().GetSchemas()["get_ticket"]
}

// NewToolset returns a new Toolset for the given key.
func NewToolset(key string) *Toolset {
	toolsets := map[string][]string{
		"reasoning": {"publish_message", "worker"},
		"messaging": {"publish_message"},
		"boards":    {"create_workitem", "search_workitems", "get_workitem"},
		"trengo":    {"list_labels", "assign_label"},
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
