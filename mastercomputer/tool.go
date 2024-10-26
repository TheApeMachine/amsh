package mastercomputer

import (
	"encoding/json"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/errnie"
)

/*
SideEffect is an interface that an object can implement if it wants to act as a Tool's side effect.
*/
type SideEffect interface {
	Start() string
}

/*
Tool defines the common structure for all tools, which are defined in the config.yml file.
*/
type Tool struct {
	Name           string                 `yaml:"name"`
	Description    string                 `yaml:"description"`
	Parameters     map[string]interface{} `yaml:"parameters"`
	SideEffect     SideEffect
	SideEffectFunc func(parameters map[string]any) SideEffect // Added field
}

func NewTool(name, description string, parameters map[string]any, sideEffectFunc func(parameters map[string]any) SideEffect) *Tool {
	errnie.Trace()
	return &Tool{
		Name:           name,
		Description:    description,
		Parameters:     parameters,
		SideEffectFunc: sideEffectFunc, // Initialize with the function
	}
}

func (tool *Tool) Use(toolCall openai.ChatCompletionMessageToolCall, workerID string) (out string) {
	errnie.Trace()
	arguments := map[string]any{}
	var err error

	if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &arguments); err != nil {
		return err.Error()
	}

	arguments["worker_id"] = workerID

	if tool.SideEffectFunc != nil {
		tool.SideEffect = tool.SideEffectFunc(arguments)
	}
	if tool.SideEffect != nil {
		if out = tool.SideEffect.Start(); out != "" {
			return out
		}
	}

	return "Tool not implemented"
}

// Implementing the Schema method to satisfy openai.ChatCompletionToolParam interface
func (tool *Tool) Schema() openai.ChatCompletionToolParam {
	errnie.Trace()
	return openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String(tool.Name),
			Description: openai.String(tool.Description),
			Parameters:  openai.F(mapConfigToFunctionParameters(tool.Parameters)),
		}),
	}
}

// Helper function to map config parameters to openai.FunctionParameters
func mapConfigToFunctionParameters(params map[string]interface{}) openai.FunctionParameters {
	errnie.Trace()
	properties := make(map[string]interface{})
	required := []string{}

	for key, value := range params {
		if paramMap, ok := value.(map[string]interface{}); ok {
			properties[key] = paramMap
			if _, isRequired := paramMap["required"]; isRequired {
				required = append(required, key)
			}
		}
	}

	return openai.FunctionParameters{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}
