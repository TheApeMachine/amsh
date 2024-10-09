package format

import (
	"fmt"
	"strings"

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
		Steps []struct {
			Thought   string `json:"thought"`
			Reasoning string `json:"reasoning"`
			Strategy  string `json:"strategy"`
		} `json:"steps"`
		OrderedStrategies []string `json:"ordered_strategies"`
	}
}

func NewReasoningStrategy() *ReasoningStrategy {
	errnie.Trace()

	definition, err := jsonschema.GenerateSchemaForType(
		ReasoningStrategy{}.Template,
	)

	if errnie.Error(err) != nil {
		return nil
	}

	return &ReasoningStrategy{
		Label:      "reasoning_strategy",
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
	builder := strings.Builder{}
	builder.WriteString("[REASONING STRATEGY]\n")
	for _, step := range format.Template.Steps { // Changed 'Strategies' to 'Steps'
		builder.WriteString("  [STEP]")
		builder.WriteString(fmt.Sprintf("      Thought: %s\n", step.Thought))
		builder.WriteString(fmt.Sprintf("    Reasoning: %s\n", step.Reasoning))
		builder.WriteString(fmt.Sprintf("     Strategy: %s\n", step.Strategy))
		builder.WriteString("  [STEP]")
	}
	builder.WriteString("[/REASONING STRATEGY]")
	return builder.String()
}
