package mastercomputer

import (
	"context"

	"github.com/openai/openai-go"
)

type Feature interface {
	Initialize() error
	Run(ctx context.Context, parentID string, args map[string]any) (string, error)
}

type WorkerTool struct {
	System  string `json:"system" jsonschema_description:"The system prompt"`
	User    string `json:"user" jsonschema_description:"The user prompt"`
	Toolset string `json:"toolset" jsonschema:"enum=core,enum=extended,enum=full" jsonschema_description:"The toolset the worker should use"`
}

func NewWorkerTool() openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("worker"),
			Description: openai.String("Create any type of worker, by providing a system prompt, a user prompt, and a toolset"),
			Parameters: openai.F(openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"system": map[string]string{
						"type": "string",
					},
					"user": map[string]string{
						"type": "string",
					},
					"toolset": map[string]any{
						"type": "string",
						"enum": []string{"core", "extended", "full"},
					},
				},
				"required": []string{"system", "user", "toolset"},
			}),
		}),
	}
}
