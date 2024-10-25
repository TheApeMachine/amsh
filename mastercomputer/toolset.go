package mastercomputer

import (
	"encoding/json"

	"github.com/openai/openai-go"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

// Toolset represents a set of tools available to a worker.
type Toolset struct {
	toolMap   map[string]openai.ChatCompletionToolParam
	baseTools []string
	workloads map[string][]string
}

/*
NewToolset returns a toolset based on the workload of the worker.
*/
func NewToolset() *Toolset {
	return &Toolset{
		toolMap: func() map[string]openai.ChatCompletionToolParam {
			toolMap := map[string]openai.ChatCompletionToolParam{}
			toolsets := viper.GetViper().GetStringMapStringSlice("toolsets")
			toolsConfig := viper.GetStringMap("tools") // Retrieve once for efficiency

			for _, toolset := range toolsets {
				for _, toolName := range toolset {
					toolInterface, exists := toolsConfig[toolName]
					if !exists {
						// Handle the case where the toolName does not exist in toolsConfig
						continue
					}

					// Assert that toolInterface is a map[string]interface{}
					toolDef, ok := toolInterface.(map[string]interface{})
					if !ok {
						continue
					}

					description, descOk := toolDef["description"].(string)
					parameters, paramOk := toolDef["parameters"].(map[string]interface{})
					if !descOk || !paramOk {
						continue
					}

					tool := NewTool(toolName, description, parameters)
					toolMap[toolName] = tool.Schema()
				}
			}
			return toolMap
		}(),
	}
}

/*
WithBaseTools adds the common tools, such as memory management, and the ability to change state.
*/
func (toolset *Toolset) WithBaseTools(in []openai.ChatCompletionToolParam) (out []openai.ChatCompletionToolParam) {
	errnie.Trace()

	for _, tool := range toolset.baseTools {
		if tool, ok := toolset.toolMap[tool]; ok {
			out = append(out, tool)
		}
	}

	return append(out, in...)
}

/*
Assign a set of tools to a worker based on the workload.
*/
func (toolset *Toolset) Assign(workload string) (out []openai.ChatCompletionToolParam) {
	errnie.Trace()

	if tools, ok := toolset.workloads[workload]; ok {
		for _, tool := range tools {
			if tool, ok := toolset.toolMap[tool]; ok {
				out = append(out, tool)
			}
		}
	}

	return out
}

/*
Use a tool, based on the tool call passed in.
*/
func (toolset *Toolset) Use(
	toolCall openai.ChatCompletionMessageToolCall,
) (out openai.ChatCompletionToolMessageParam) {
	errnie.Trace()

	args := map[string]any{}
	var err error

	if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return openai.ToolMessage(toolCall.ID, err.Error())
	}

	return openai.ToolMessage(toolCall.ID, "Tool not implemented")
}
