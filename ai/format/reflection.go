package format

import (
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

/*
SelfReflection is a reasoning strategy where the AI system reflects on its own thoughts and actions.
*/
type SelfReflection struct {
	Label      string
	Definition *jsonschema.Definition
	Template   struct {
		Reflections []struct {
			PreviousThought string `json:"previous_thought"`
			Reflection      string `json:"reflection"`
		}
		Summary string `json:"summary"`
	}
}

func NewSelfReflection() *SelfReflection {
	errnie.Trace()

	definition, err := jsonschema.GenerateSchemaForType(
		SelfReflection{}.Template,
	)

	if errnie.Error(err) != nil {
		return nil
	}

	return &SelfReflection{
		Label:      "self_reflection",
		Definition: definition,
	}
}

func (reflection *SelfReflection) Name() string {
	errnie.Trace()
	return reflection.Label
}

func (reflection *SelfReflection) Schema() *jsonschema.Definition {
	errnie.Trace()
	return reflection.Definition
}

func (reflection *SelfReflection) ToString() string {
	builder := strings.Builder{}
	builder.WriteString("[SELF REFLECTION]\n")
	for _, reflection := range reflection.Template.Reflections {
		builder.WriteString(fmt.Sprintf("  [REFLECTION]"))
		builder.WriteString(fmt.Sprintf("       Thought: %s\n", reflection.PreviousThought))
		builder.WriteString(fmt.Sprintf("    Reflection: %s\n", reflection.Reflection))
		builder.WriteString("  [REFLECTION]")
	}
	builder.WriteString(fmt.Sprintf("  Summary: %s\n", reflection.Template.Summary))
	builder.WriteString("[/SELF REFLECTION]")
	return builder.String()
}
