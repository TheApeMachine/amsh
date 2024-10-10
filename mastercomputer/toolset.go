package mastercomputer

import (
	"context"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Feature interface {
	Initialize() error
	Run(ctx context.Context, parentID string, args map[string]any) (string, error)
}

type Toolset struct {
	tools    map[string][]openai.Tool
	Function *openai.FunctionDefinition
}

func NewToolSet(ctx context.Context) *Toolset {
	errnie.Trace()

	return &Toolset{
		tools: map[string][]openai.Tool{
			"system": {
				openai.Tool{
					Type:     openai.ToolTypeFunction,
					Function: NewLink().Function,
				},
				openai.Tool{
					Type:     openai.ToolTypeFunction,
					Function: NewWorker().Function,
				},
				openai.Tool{
					Type:     openai.ToolTypeFunction,
					Function: NewCommand().Function,
				},
				openai.Tool{
					Type:     openai.ToolTypeFunction,
					Function: NewLogicCircuit().Function,
				},
				openai.Tool{
					Type:     openai.ToolTypeFunction,
					Function: NewBrowser().Function,
				},
				openai.Tool{
					Type:     openai.ToolTypeFunction,
					Function: NewEnvironment().Function,
				},
			},
			"research": {
				openai.Tool{
					Type:     openai.ToolTypeFunction,
					Function: NewBrowser().Function,
				},
			},
			"development": {
				openai.Tool{
					Type:     openai.ToolTypeFunction,
					Function: NewEnvironment().Function,
				},
			},
		},
		Function: &openai.FunctionDefinition{
			Name:        "toolset",
			Description: "Use to create a toolset, which is a collection of tools that can be assigned to workers.",
			Strict:      true,
			Parameters: jsonschema.Definition{
				Type:                 jsonschema.Object,
				AdditionalProperties: false,
				Description:          "Use to create a toolset, which is a collection of tools that can be assigned to workers.",
				Properties: map[string]jsonschema.Definition{
					"tools": {
						Type:        jsonschema.Array,
						Description: "The list of tools to use",
						Enum:        []string{"worker", "command", "link", "logic_circuit"},
					},
				},
				Required: []string{"tools"},
			},
		},
	}
}

func (toolset *Toolset) Tools(key string) []openai.Tool {
	errnie.Trace()

	return toolset.tools[key]
}
