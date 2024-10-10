package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type BackwardReasoning struct {
	Goal      string   `json:"goal"`
	Steps     []string `json:"steps"`
	Reason    string   `json:"reason"`
	FinalPlan string   `json:"final_plan"`
}

func NewBackwardReasoning() *BackwardReasoning {
	errnie.Trace()
	return &BackwardReasoning{}
}

func (br *BackwardReasoning) FinalAnswer() string {
	return br.FinalPlan
}

func (br *BackwardReasoning) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(br)
}

func (br *BackwardReasoning) ToString() string {
	out := []string{}
	out = append(out, dark("  [BACKWARD REASONING]"))
	out = append(out, blue("    Goal: ")+highlight(br.Goal))
	out = append(out, green("    Steps:"))
	for _, step := range br.Steps {
		out = append(out, green("      - ")+highlight(step))
	}
	out = append(out, yellow("        Reason: ")+highlight(br.Reason))
	out = append(out, red("    Final Plan: ")+highlight(br.FinalPlan))
	out = append(out, dark("  [/BACKWARD REASONING]"))
	return strings.Join(out, "\n")
}
