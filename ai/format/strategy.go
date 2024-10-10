package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type ReasoningStrategy struct {
	Steps []struct {
		Thought       string `json:"thought"`
		Strategy      string `json:"strategy"`
		Clarification string `json:"clarification"`
	} `json:"steps"`
	OrderedStrategies []string `json:"ordered_strategies"`
}

func NewReasoningStrategy() *ReasoningStrategy {
	errnie.Trace()
	return &ReasoningStrategy{}
}

func (strategy *ReasoningStrategy) FinalAnswer() string {
	return strings.Join(strategy.OrderedStrategies, ",")
}

func (strategy *ReasoningStrategy) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(strategy)
}

func (strategy *ReasoningStrategy) ToString() string {
	out := []string{}
	out = append(out, dark("  [REASONING STRATEGY]"))
	for _, step := range strategy.Steps {
		out = append(out, muted("    [STEP]"))
		out = append(out, red("            Thought: ")+highlight(step.Thought))
		out = append(out, yellow("           Strategy: ")+highlight(step.Strategy))
		out = append(out, green("      Clarification: ")+highlight(step.Clarification))
		out = append(out, muted("    [/STEP]"))
	}
	out = append(out, blue("    OrderedStrategies: ")+highlight(strings.Join(strategy.OrderedStrategies, ", ")))
	out = append(out, dark("  [/REASONING STRATEGY]"))
	return strings.Join(out, "\n")
}
