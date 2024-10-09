package format

import (
	"fmt"

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
	output := "[ First Principles Reasoning ]\n"
	output += fmt.Sprintf("Principle: %s\n", firstPrinciples.Template.Principle)
	for _, concept := range firstPrinciples.Template.Breakdown {
		output += fmt.Sprintf("[concept]\n  key concept: %s\n  explanation: %s\n[/concept]\n", concept.KeyConcept, concept.Explanation)
	}
	output += fmt.Sprintf("Derived Solution: %s\n", firstPrinciples.Template.DerivedSolution)
	return output
}
