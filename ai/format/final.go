package format

import (
	"fmt"
	"strings"

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
	builder := strings.Builder{}
	builder.WriteString("[FINAL]\n")
	for _, step := range format.Template.Verify {
		builder.WriteString(fmt.Sprintf("       Thought: %s\n", step.Thought))
		builder.WriteString(fmt.Sprintf("     Reasoning: %s\n", step.Reasoning))
		builder.WriteString(fmt.Sprintf("  Verification: %s\n", step.Verification))
	}
	builder.WriteString(fmt.Sprintf("  Final Answer: %s\n", format.Template.FinalAnswer))
	builder.WriteString("[/FINAL]")
	return builder.String()
}
