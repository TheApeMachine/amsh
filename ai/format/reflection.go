package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

/*
SelfReflection is a reasoning strategy where the AI system reflects on its own thoughts and actions.
*/
type SelfReflection struct {
	Reflections []struct {
		PreviousThought string `json:"previous_thought"`
		Reflection      string `json:"reflection"`
	}
	Summary string `json:"summary"`
}

func NewSelfReflection() *SelfReflection {
	errnie.Trace()
	return &SelfReflection{}
}

func (reflection *SelfReflection) FinalAnswer() string {
	return reflection.Summary
}

func (reflection *SelfReflection) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(reflection)
}

func (reflection *SelfReflection) ToString() string {
	out := []string{}
	out = append(out, dark("  [SELF REFLECTION]"))
	for _, r := range reflection.Reflections {
		out = append(out, muted("    [REFLECTION]"))
		out = append(out, yellow("      Thought: ")+highlight(r.PreviousThought))
		out = append(out, green("      Reflection: ")+highlight(r.Reflection))
		out = append(out, muted("    [/REFLECTION]"))
	}
	out = append(out, blue("    Summary: ")+highlight(reflection.Summary))
	out = append(out, dark("  [/SELF REFLECTION]"))
	return strings.Join(out, "\n")
}
