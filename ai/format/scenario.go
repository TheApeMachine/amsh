package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type ScenarioAnalysis struct {
	Scenario        string   `json:"scenario"`
	Assumptions     []string `json:"assumptions"`
	ExpectedOutcome string   `json:"expected_outcome"`
	Result          string   `json:"result"`
}

func NewScenarioAnalysis() *ScenarioAnalysis {
	errnie.Trace()
	return &ScenarioAnalysis{}
}

func (sa *ScenarioAnalysis) FinalAnswer() string {
	return sa.Result
}

func (sa *ScenarioAnalysis) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(sa)
}

func (sa *ScenarioAnalysis) ToString() string {
	out := []string{}
	out = append(out, dark("  [SCENARIO ANALYSIS]"))
	out = append(out, blue("    Scenario: ")+highlight(sa.Scenario))
	out = append(out, green("    Assumptions:"))
	for _, assumption := range sa.Assumptions {
		out = append(out, green("      - ")+highlight(assumption))
	}
	out = append(out, yellow("    Expected Outcome: ")+highlight(sa.ExpectedOutcome))
	out = append(out, red("    Result: ")+highlight(sa.Result))
	out = append(out, dark("  [/SCENARIO ANALYSIS]"))
	return strings.Join(out, "\n")
}
