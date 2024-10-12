package mastercomputer

import "github.com/openai/openai-go"

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
