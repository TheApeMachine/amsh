package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type DivideAndConquer struct {
	SubProblems []struct {
		SubProblem string `json:"sub_problem"`
		Solution   string `json:"solution"`
	}
	AggregatedResult string `json:"aggregated_result"`
}

func NewDivideAndConquer() *DivideAndConquer {
	errnie.Trace()
	return &DivideAndConquer{}
}

func (dac *DivideAndConquer) FinalAnswer() string {
	return dac.AggregatedResult
}

func (dac *DivideAndConquer) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(dac)
}

func (dac *DivideAndConquer) ToString() string {
	out := []string{}
	out = append(out, dark("  [DIVIDE AND CONQUER]"))
	for _, subProblem := range dac.SubProblems {
		out = append(out, muted("    [SUB-PROBLEM]"))
		out = append(out, red("       Problem: ")+highlight(subProblem.SubProblem))
		out = append(out, green("      Solution: ")+highlight(subProblem.Solution))
		out = append(out, muted("    [/SUB-PROBLEM]"))
	}
	out = append(out, blue("    Aggregated Result: ")+highlight(dac.AggregatedResult))
	out = append(out, dark("  [/DIVIDE AND CONQUER]"))
	return strings.Join(out, "\n")
}
