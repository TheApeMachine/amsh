package format

import (
	"fmt"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Strategy string

const (
	StrategyChainOfThought  Strategy = "chain_of_thought"
	StrategyTreeOfThought   Strategy = "tree_of_thought"
	StrategyFirstPrinciples Strategy = "first_principles"
)

type ReasoningStrategy struct {
	Label      string                 `json:"name"`
	Definition *jsonschema.Definition `json:"schema"`
	Template   struct {
		Strategies []struct {
			Thought   string   `json:"thought"`
			Reasoning string   `json:"reasoning"`
			Strategy  Strategy `json:"next_step"`
		}
	}
}

func NewReasoningStrategy() *ReasoningStrategy {
	errnie.Trace()

	definition, err := jsonschema.GenerateSchemaForType(
		ChainOfThought{}.Template,
	)

	if errnie.Error(err) != nil {
		return nil
	}

	return &ReasoningStrategy{
		Label:      "strategy",
		Definition: definition,
	}
}

func (format *ReasoningStrategy) Name() string {
	errnie.Trace()
	return format.Label
}

func (format *ReasoningStrategy) Schema() *jsonschema.Definition {
	errnie.Trace()
	return format.Definition
}

func (format *ReasoningStrategy) ToString() string {
	output := "[ Reasoning Strategy ]\n"
	for _, step := range format.Template.Strategies {
		output += fmt.Sprintf("[step]\n  thought: %s\n  reasoning: %s\n  next step: %s\n[/step]\n\n", step.Thought, step.Reasoning, step.Strategy)
	}
	return output
}
