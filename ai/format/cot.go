package format

import (
	"fmt"
	"strings"

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
	builder := strings.Builder{}
	builder.WriteString("[CHAIN OF THOUGHT]\n")
	for _, step := range chain.Template.Steps {
		builder.WriteString("  [STEP]")
		builder.WriteString(fmt.Sprintf("      Thought: %s\n", step.Thought))
		builder.WriteString(fmt.Sprintf("    Reasoning: %s\n", step.Reasoning))
		builder.WriteString(fmt.Sprintf("    Next Step: %s\n", step.NextStep))
		builder.WriteString("  [STEP]")
	}
	builder.WriteString(fmt.Sprintf("Action: %s\n", chain.Template.Action))
	builder.WriteString(fmt.Sprintf("Result: %s\n", chain.Template.Result))
	builder.WriteString("[/CHAIN OF THOUGHT]")
	return builder.String()
}
