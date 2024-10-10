package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type ProsAndConsAnalysis struct {
	Options []struct {
		Option string   `json:"option"`
		Pros   []string `json:"pros"`
		Cons   []string `json:"cons"`
	}
	Recommendation string `json:"recommendation"`
}

func NewProsAndConsAnalysis() *ProsAndConsAnalysis {
	errnie.Trace()
	return &ProsAndConsAnalysis{}
}

func (pac *ProsAndConsAnalysis) FinalAnswer() string {
	return pac.Recommendation
}

func (pac *ProsAndConsAnalysis) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(pac)
}

func (pac *ProsAndConsAnalysis) ToString() string {
	out := []string{}
	out = append(out, dark("  [PROS AND CONS ANALYSIS]"))
	for _, option := range pac.Options {
		out = append(out, muted("    [OPTION %d]"))
		out = append(out, blue("      Option: ")+highlight(option.Option))
		out = append(out, green("      Pros:"))
		for _, pro := range option.Pros {
			out = append(out, green("        - ")+highlight(pro))
		}
		out = append(out, red("      Cons:"))
		for _, con := range option.Cons {
			out = append(out, red("        - ")+highlight(con))
		}
		out = append(out, muted("    [/OPTION %d]"))
	}
	out = append(out, yellow("    Recommendation: ")+highlight(pac.Recommendation))
	out = append(out, dark("  [/PROS AND CONS ANALYSIS]"))
	return strings.Join(out, "\n")
}
