package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Final struct {
	Verify []struct {
		Thought      string `json:"thought"`
		Reasoning    string `json:"reasoning"`
		Verification string `json:"verification"`
	}
	FinalConclusion string `json:"final_conclusion"`
}

func NewFinal() *Final {
	errnie.Trace()

	return &Final{}
}

func (final *Final) FinalAnswer() string {
	return final.FinalConclusion
}

func (final *Final) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(final)
}

func (format *Final) ToString() string {
	out := []string{}
	out = append(out, dark("  [FINAL CONCLUSION]"))
	for _, step := range format.Verify {
		out = append(out, muted("    [VERIFICATION]"))
		out = append(out, red("      Thought: ")+highlight(step.Thought))
		out = append(out, yellow("      Reasoning: ")+highlight(step.Reasoning))
		out = append(out, green("      Verification: ")+highlight(step.Verification))
		out = append(out, muted("    [/VERIFICATION]"))
	}
	out = append(out, blue("    Final Conclusion: ")+highlight(format.FinalConclusion))
	out = append(out, dark("  [/FINAL CONCLUSION]"))
	return strings.Join(out, "\n")
}
