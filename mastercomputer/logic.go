package mastercomputer

import (
	"context"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type LogicCircuit struct {
	Function *openai.FunctionDefinition
}

func NewLogicCircuit() *LogicCircuit {
	return &LogicCircuit{
		Function: &openai.FunctionDefinition{
			Name:        "logic_circuit",
			Description: "Use to create a logic circuit, which is a construct of workers primed to follow a set of instructions.",
			Strict:      true,
			Parameters: jsonschema.Definition{
				Type:                 jsonschema.Object,
				AdditionalProperties: false,
				Description:          "Use to create a logic circuit, which is a construct of workers primed to follow a set of instructions.",
				Properties: map[string]jsonschema.Definition{
					"workers": {
						Type:        jsonschema.Array,
						Description: "The list of IDs of the workers that need to become part of the logic circuit",
						Items: &jsonschema.Definition{
							Type: jsonschema.String,
						},
					},
					"instructions": {
						Type:        jsonschema.String,
						Description: "The instructions for the logic circuit, optionally including the objectives, goals, and method/structure of the approach.",
					},
					"output": {
						Type:        jsonschema.String,
						Description: "The desired output, or output format the logic circuit should return",
					},
					"ttl": {
						Type:        jsonschema.Number,
						Description: "The time to live for the logic circuit, in seconds, or iterations",
					},
					"timeunit": {
						Type:        jsonschema.String,
						Description: "The time unit for the ttl, either 'seconds' or 'iterations'",
						Enum:        []string{"seconds", "iterations"},
					},
				},
				Required: []string{"workers", "instructions", "output", "ttl", "timeunit"},
			},
		},
	}
}

func (logic *LogicCircuit) Initialize() error {
	errnie.Trace()
	return nil
}

func (logic *LogicCircuit) Run(ctx context.Context, args map[string]any) (string, error) {
	errnie.Trace()
	return "", nil
}
