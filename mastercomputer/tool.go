package mastercomputer

import (
	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/ai"
)

/*
SideEffect is an interface that an object can implement if it wants to act as a Tool's side effect.
*/
type SideEffect interface {
	Call() string
	Parameters() map[string]interface{}
}

/*
Tool defines the common structure for all tools, which are defined in the config.yml file.
*/
type Tool struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Parameters  map[string]interface{} `yaml:"parameters"`
}

func NewTool(namer, description string, parameters map[string]interface{}) *Tool {
	return &Tool{
		Description: description,
		Parameters:  parameters,
	}
}

func (tool *Tool) Use(toolCall openai.ToolCall, sideEffect SideEffect) {
}

// Implementing the Schema method to satisfy openai.ChatCompletionToolParam interface
func (tool *Tool) Schema() openai.ChatCompletionToolParam {
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

func (toolset *Toolset) makeBoolParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
}

func (toolset *Toolset) makeStringParam(description string) map[string]string {
	return map[string]string{
		"type":        "string",
		"description": description,
	}
}

func (toolset *Toolset) makeIntParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "integer",
		"description": description,
	}
}

func (toolset *Toolset) makeEnumParam(description string, values []string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"enum":        values,
		"description": description,
	}
}

func (toolset *Toolset) makeArrayParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": description,
		"items":       toolset.makeStringParam("The items in the array."),
	}
}

func (toolset *Toolset) makeSchema(name, description string, params map[string]interface{}) openai.ChatCompletionToolParam {
	return ai.MakeTool(
		name,
		description,
		openai.FunctionParameters{
			"type":       "object",
			"properties": params,
			"required":   toolset.getKeys(params),
		},
	)
}

func (toolset *Toolset) getKeys(params map[string]interface{}) []string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	return keys
}
