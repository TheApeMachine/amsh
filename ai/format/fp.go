package format

import (
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type FirstPrinciplesReasoning struct {
	Label      string                 `json:"name"`
	Definition *jsonschema.Definition `json:"schema"`
	Template   struct {
		Principle string `json:"principle"`
		Breakdown []struct {
			KeyConcept  string `json:"key_concept"`
			Explanation string `json:"explanation"`
		}
		DerivedSolution string `json:"derived_solution"`
	}
}

func NewFirstPrinciplesReasoning() *FirstPrinciplesReasoning {
	errnie.Trace()

	definition, err := jsonschema.GenerateSchemaForType(
		FirstPrinciplesReasoning{}.Template,
	)

	if errnie.Error(err) != nil {
		return nil
	}

	return &FirstPrinciplesReasoning{
		Label:      "first_principles_reasoning",
		Definition: definition,
	}
}

func (format *FirstPrinciplesReasoning) Name() string {
	errnie.Trace()
	return format.Label
}

func (format *FirstPrinciplesReasoning) Schema() *jsonschema.Definition {
	errnie.Trace()
	return format.Definition
}

func (firstPrinciples FirstPrinciplesReasoning) ToString() string {
	builder := strings.Builder{}
	builder.WriteString("[FIRST PRINCIPLES REASONING]\n")
	builder.WriteString(fmt.Sprintf("  Principle: %s\n", firstPrinciples.Template.Principle))
	for _, breakdown := range firstPrinciples.Template.Breakdown {
		builder.WriteString(fmt.Sprintf("    Key Concept: %s\n", breakdown.KeyConcept))
		builder.WriteString(fmt.Sprintf("    Explanation: %s\n", breakdown.Explanation))
	}
	builder.WriteString(fmt.Sprintf("  Derived Solution: %s\n", firstPrinciples.Template.DerivedSolution))
	builder.WriteString("[/FIRST PRINCIPLES REASONING]")
	return builder.String()
}
