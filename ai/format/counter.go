package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type CounterfactualThinking struct {
	Scenario            string `json:"scenario"`
	AlternativeAction   string `json:"alternative_action"`
	HypotheticalOutcome string `json:"hypothetical_outcome"`
}

func NewCounterfactualThinking() *CounterfactualThinking {
	errnie.Trace()
	return &CounterfactualThinking{}
}

func (ct *CounterfactualThinking) FinalAnswer() string {
	return ct.HypotheticalOutcome
}

func (ct *CounterfactualThinking) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(ct)
}

func (ct *CounterfactualThinking) ToString() string {
	out := []string{}
	out = append(out, dark("  [COUNTERFACTUAL THINKING]"))
	out = append(out, blue("    Scenario: ")+highlight(ct.Scenario))
	out = append(out, green("    Alternative Action: ")+highlight(ct.AlternativeAction))
	out = append(out, yellow("    Hypothetical Outcome: ")+highlight(ct.HypotheticalOutcome))
	out = append(out, dark("  [/COUNTERFACTUAL THINKING]"))
	return strings.Join(out, "\n")
}
