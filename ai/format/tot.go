package format

import (
	"fmt"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type TreeOfThought struct {
	Label      string                 `json:"name"`
	Definition *jsonschema.Definition `json:"schema"`
	Template   struct {
		RootThought string `json:"root_thought"`
		Branches    []struct {
			Thought   string `json:"thought"`
			Reasoning string `json:"reasoning"`
			Branches  []struct {
				SubThought   string `json:"sub_thought"`
				Reasoning    string `json:"reasoning"`
				Outcome      string `json:"outcome"`
				NextBranches []struct {
					SubThought string `json:"sub_thought"`
					Reasoning  string `json:"reasoning"`
					Outcome    string `json:"outcome"`
				}
			}
		}
		FinalConclusion string `json:"final_conclusion"`
	}
}

func NewTreeOfThought() *TreeOfThought {
	errnie.Trace()

	definition, err := jsonschema.GenerateSchemaForType(
		TreeOfThought{}.Template,
	)

	if errnie.Error(err) != nil {
		return nil
	}

	return &TreeOfThought{
		Label:      "tree_of_thought",
		Definition: definition,
	}
}

func (format *TreeOfThought) Name() string {
	errnie.Trace()
	return format.Label
}

func (format *TreeOfThought) Schema() *jsonschema.Definition {
	errnie.Trace()
	return format.Definition
}

func (tree TreeOfThought) ToString() string {
	output := "[ Tree of Thought ]\n"
	output += fmt.Sprintf("Root Thought: %s\n", tree.Template.RootThought)
	for _, branch := range tree.Template.Branches {
		output += fmt.Sprintf("[branch]\n  thought: %s\n  reasoning: %s\n[/branch]\n", branch.Thought, branch.Reasoning)
	}
	output += fmt.Sprintf("Final Conclusion: %s\n", tree.Template.FinalConclusion)
	return output
}
