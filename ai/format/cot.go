package format

import (
	"fmt"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Response interface {
	Name() string
	Schema() *jsonschema.Definition
	ToString() string
}

type ChainOfThought struct {
	Label      string                 `json:"name"`
	Definition *jsonschema.Definition `json:"schema"`
	Template   struct {
		Steps []struct {
			Thought   string `json:"thought"`
			Reasoning string `json:"reasoning"`
			NextStep  string `json:"next_step"`
		}
		Action string `json:"action"`
		Result string `json:"result"`
	}
}

func NewChainOfThought() *ChainOfThought {
	errnie.Trace()

	definition, err := jsonschema.GenerateSchemaForType(
		ChainOfThought{}.Template,
	)

	if errnie.Error(err) != nil {
		return nil
	}

	return &ChainOfThought{
		Label:      "chain_of_thought",
		Definition: definition,
	}
}

func (format *ChainOfThought) Name() string {
	errnie.Trace()
	return format.Label
}

func (format *ChainOfThought) Schema() *jsonschema.Definition {
	errnie.Trace()
	return format.Definition
}

func (chain ChainOfThought) ToString() string {
	output := "[ Chain of Thought ]\n"
	for _, step := range chain.Template.Steps {
		output += fmt.Sprintf("[step]\n  thought: %s\n  reasoning: %s\n  next step: %s\n[/step]\n\n", step.Thought, step.Reasoning, step.NextStep)
	}
	output += fmt.Sprintf("action: %s\nresult: %s\n", chain.Template.Action, chain.Template.Result)
	return output
}
