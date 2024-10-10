package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type FirstPrinciplesReasoning struct {
	Principle string `json:"principle"`
	Breakdown []struct {
		KeyConcept  string `json:"key_concept"`
		Explanation string `json:"explanation"`
	}
	DerivedSolution string `json:"derived_solution"`
}

func NewFirstPrinciplesReasoning() *FirstPrinciplesReasoning {
	errnie.Trace()
	return &FirstPrinciplesReasoning{}
}

func (firstPrinciples *FirstPrinciplesReasoning) FinalAnswer() string {
	return firstPrinciples.DerivedSolution
}

func (firstPrinciples *FirstPrinciplesReasoning) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(firstPrinciples)
}

func (firstPrinciples *FirstPrinciplesReasoning) ToString() string {
	out := []string{}
	out = append(out, dark("  [FIRST PRINCIPLES REASONING]"))
	out = append(out, red("    Principle: ")+highlight(firstPrinciples.Principle))
	for _, breakdown := range firstPrinciples.Breakdown {
		out = append(out, yellow("      Key Concept: ")+highlight(breakdown.KeyConcept))
		out = append(out, green("      Explanation: ")+highlight(breakdown.Explanation))
	}
	out = append(out, blue("    Derived Solution: ")+highlight(firstPrinciples.DerivedSolution))
	out = append(out, dark("  [/FIRST PRINCIPLES REASONING]"))
	return strings.Join(out, "\n")
}
