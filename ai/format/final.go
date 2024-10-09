package format

import (
	"fmt"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Final struct {
	Label      string                 `json:"name"`
	Definition *jsonschema.Definition `json:"schema"`
	Template   struct {
		Verify []struct {
			Thought      string `json:"thought"`
			Reasoning    string `json:"reasoning"`
			Verification string `json:"verification"`
		}
		FinalAnswer string `json:"final_answer"`
	}
}

func NewFinal() *Final {
	errnie.Trace()

	definition, err := jsonschema.GenerateSchemaForType(
		ChainOfThought{}.Template,
	)

	if errnie.Error(err) != nil {
		return nil
	}

	return &Final{
		Label:      "strategy",
		Definition: definition,
	}
}

func (format *Final) Name() string {
	errnie.Trace()
	return format.Label
}

func (format *Final) Schema() *jsonschema.Definition {
	errnie.Trace()
	return format.Definition
}

func (format *Final) ToString() string {
	output := "[ Final ]\n"
	for _, step := range format.Template.Verify {
		output += fmt.Sprintf("[step]\n  thought: %s\n  reasoning: %s\n  verification: %s\n[/step]\n\n", step.Thought, step.Reasoning, step.Verification)
	}
	output += fmt.Sprintf("final_answer: %s\n", format.Template.FinalAnswer)
	return output
}
