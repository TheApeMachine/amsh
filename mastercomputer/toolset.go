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
	"browser": makeTool(
		"browser",
		"Use a fully functional chrome browser.",
		BrowserToolSchema(),
	),
	"store_vector_memory": makeTool(
		"store_vector_memory",
		"Store a memory inside the vector database.",
		StoreVectorMemoryToolSchema(),
	),
	"search_vector_memory": makeTool(
		"search_vector_memory",
		"Search the vector database for a memory.",
		SearchVectorMemoryToolSchema(),
	),
	"delete_vector_memory": makeTool(
		"delete_vector_memory",
		"Delete a memory from the vector database.",
		DeleteVectorMemoryToolSchema(),
	),
	"store_graph_memory": makeTool(
		"store_graph_memory",
		"Store a memory inside the graph database.",
		StoreGraphMemoryToolSchema(),
	),
	"search_graph_memory": makeTool(
		"search_graph_memory",
		"Search the graph database for a memory.",
		SearchGraphMemoryToolSchema(),
	),
	"delete_graph_memory": makeTool(
		"delete_graph_memory",
		"Delete a memory from the graph database.",
		DeleteGraphMemoryToolSchema(),
	),
}

/*
NewToolset returns a new Toolset for the given key.
*/
func NewToolset(key string) *Toolset {
	errnie.Trace()

	toolsets := map[string][]string{
		"reasoning": {
			"publish_message",
			"store_vector_memory",
			"search_vector_memory",
			"delete_vector_memory",
			"store_graph_memory",
			"search_graph_memory",
			"delete_graph_memory",
			"worker",
			"browser",
		},
		"execution": {
			"publish_message",
			"store_vector_memory",
			"search_vector_memory",
			"delete_vector_memory",
			"store_graph_memory",
			"search_graph_memory",
			"delete_graph_memory",
			"environment",
			"browser",
		},
		"routing": {
			"publish_message",
			"store_vector_memory",
			"search_vector_memory",
			"delete_vector_memory",
			"store_graph_memory",
			"search_graph_memory",
			"delete_graph_memory",
		},
		"none": {
			"publish_message",
			"store_vector_memory",
			"search_vector_memory",
			"delete_vector_memory",
			"store_graph_memory",
			"search_graph_memory",
			"delete_graph_memory",
		},
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
		"properties": map[string]any{
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
		"properties": map[string]any{
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
		"properties": map[string]any{
			"justification": map[string]string{
				"type":        "string",
				"description": "A valid reason why you are requesting a new environment.",
			},
		},
		"required": []string{"justification"},
	}
}

/*
BrowserToolSchema is the schema for the browser tool, which
enables a worker to browse the web.
*/
func BrowserToolSchema() openai.FunctionParameters {
	errnie.Trace()

	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]string{
				"type":        "string",
				"description": "The URL you want to browse.",
			},
			"javascript": map[string]string{
				"type":        "string",
				"description": "The JavaScript code you want to run on the page. Must be a function that returns a string.",
			},
		},
		"required": []string{"url", "javascript"},
	}
}

/*
VectorMemoryToolSchema is the schema for the vector memory tool, which
enables a worker to create and modify vector memory.
*/
func StoreVectorMemoryToolSchema() openai.FunctionParameters {
	errnie.Trace()

	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"memory": map[string]string{
				"type":        "string",
				"description": "The memory you want to store inside the vector database.",
			},
		},
		"required": []string{"memory"},
	}
}

/*
SearchVectorMemoryToolSchema is the schema for the search vector memory tool, which
enables a worker to search the vector database for a memory.
*/
func SearchVectorMemoryToolSchema() openai.FunctionParameters {
	errnie.Trace()

	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]string{
				"type":        "string",
				"description": "The query you want to search for inside the vector database.",
			},
		},
		"required": []string{"query"},
	}
}

/*
DeleteVectorMemoryToolSchema is the schema for the delete vector memory tool, which
enables a worker to delete a memory from the vector database.
*/
func DeleteVectorMemoryToolSchema() openai.FunctionParameters {
	errnie.Trace()

	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"key": map[string]string{
				"type":        "string",
				"description": "The key of the memory you want to delete from the vector database.",
			},
		},
		"required": []string{"key"},
	}
}

/*
StoreGraphMemoryToolSchema is the schema for the store graph memory tool, which
enables a worker to store a graph memory in the graph database.
*/
func StoreGraphMemoryToolSchema() openai.FunctionParameters {
	errnie.Trace()

	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"cypher": map[string]string{
				"type":        "string",
				"description": "The cypher query to store the memory.",
			},
		},
		"required": []string{"cypher"},
	}
}

/*
SearchGraphMemoryToolSchema is the schema for the search graph memory tool, which
enables a worker to search the graph database for a memory.
*/
func SearchGraphMemoryToolSchema() openai.FunctionParameters {
	errnie.Trace()

	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"cypher": map[string]string{
				"type":        "string",
				"description": "The cypher query to search for the memory.",
			},
		},
		"required": []string{"cypher"},
	}
}

/*
DeleteGraphMemoryToolSchema is the schema for the delete graph memory tool, which
enables a worker to delete a memory from the graph database.
*/
func DeleteGraphMemoryToolSchema() openai.FunctionParameters {
	errnie.Trace()

	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"cypher": map[string]string{
				"type":        "string",
				"description": "The cypher query to delete the memory.",
			},
		},
		"required": []string{"cypher"},
	}
}
