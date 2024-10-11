package mastercomputer

import "github.com/openai/openai-go"

type WorkerTool struct {
	System  string `json:"system" jsonschema_description:"The system prompt" jsonschema:"required"`
	User    string `json:"user" jsonschema_description:"The user prompt" jsonschema:"required"`
	Toolset string `json:"toolset" jsonschema:"enum=core,enum=extended,enum=full" jsonschema_description:"The toolset the worker should use" jsonschema_required:"true"`
}

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
			"toolset": map[string]any{
				"type":        "string",
				"enum":        []string{"core", "extended", "full"},
				"description": "The toolset the worker should use",
			},
		},
		"required": []string{"system", "user", "toolset"},
	}
}
