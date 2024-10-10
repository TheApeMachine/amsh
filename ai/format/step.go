package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type StepByStepExecution struct {
	Steps []struct {
		Action  string `json:"action"`
		Outcome string `json:"outcome"`
	}
	Conclusion string `json:"conclusion"`
}

func NewStepByStepExecution() *StepByStepExecution {
	errnie.Trace()
	return &StepByStepExecution{}
}

func (sbs *StepByStepExecution) FinalAnswer() string {
	return sbs.Conclusion
}

func (sbs *StepByStepExecution) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(sbs)
}

func (sbs *StepByStepExecution) ToString() string {
	out := []string{}
	out = append(out, dark("  [STEP BY STEP EXECUTION]"))
	for _, step := range sbs.Steps {
		out = append(out, muted("    [STEP]"))
		out = append(out, yellow("       Action: ")+highlight(step.Action))
		out = append(out, green("      Outcome: ")+highlight(step.Outcome))
		out = append(out, muted("    [/STEP]"))
	}
	out = append(out, blue("    Conclusion: ")+highlight(sbs.Conclusion))
	out = append(out, dark("  [/STEP BY STEP EXECUTION]"))
	return strings.Join(out, "\n")
}
