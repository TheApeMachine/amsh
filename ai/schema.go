package ai

import (
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type ChainOfThought struct {
	Steps []struct {
		Thought   string `json:"thought"`
		Reasoning string `json:"reasoning"`
		NextStep  string `json:"next_step"`
	}
	Action string `json:"action"`
	Result string `json:"result"`
}

func NewChainOfThought() *jsonschema.Definition {
	definition, err := jsonschema.GenerateSchemaForType(ChainOfThought{})
	if errnie.Error(err) != nil {
		return nil
	}

	return definition
}
