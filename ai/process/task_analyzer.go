package process

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/spf13/viper"
)

type TaskAnalysis struct {
	RequiredLayers []LayerGroup `json:"required_layers" jsonschema:"required,description:Layer groups required for this task"`
	Reasoning      string       `json:"reasoning" jsonschema:"required,description:Explanation of why these layers were selected"`
	Complexity     int          `json:"complexity" jsonschema:"required,description:Estimated complexity level (1-5)"`
}

type LayerGroup struct {
	Name      string   `json:"name" jsonschema:"required,description:Name of the layer group"`
	Layers    []string `json:"layers" jsonschema:"required,description:Specific layers needed in this group"`
	Priority  int      `json:"priority" jsonschema:"required,description:Processing priority (1-5)"`
	Rationale string   `json:"rationale" jsonschema:"required,description:Why this layer group is needed"`
}

func NewTaskAnalyzer() Process {
	return &TaskAnalysis{}
}

func (ta *TaskAnalysis) GenerateSchema() string {
	schema := jsonschema.Reflect(&TaskAnalysis{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return ""
	}
	return string(out)
}

func (ta *TaskAnalysis) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.task_analyzer.prompt", key))
	return strings.ReplaceAll(prompt, "{{schemas}}", ta.GenerateSchema())
}
