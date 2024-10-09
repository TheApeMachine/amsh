package format

import (
	"fmt"
	"strings"

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
	builder := strings.Builder{}
	builder.WriteString("[TREE OF THOUGHT]\n")
	builder.WriteString(fmt.Sprintf("  Root Thought: %s\n", tree.Template.RootThought))

	// Follow all branches.
	for _, branch := range tree.Template.Branches {
		builder.WriteString(fmt.Sprintf("    Thought: %s\n", branch.Thought))
		builder.WriteString(fmt.Sprintf("    Reasoning: %s\n", branch.Reasoning))
		for _, subBranch := range branch.Branches {
			builder.WriteString(fmt.Sprintf("      Sub Thought: %s\n", subBranch.SubThought))
			builder.WriteString(fmt.Sprintf("        Reasoning: %s\n", subBranch.Reasoning))
			builder.WriteString(fmt.Sprintf("          Outcome: %s\n", subBranch.Outcome))
			for _, nextBranch := range subBranch.NextBranches {
				builder.WriteString(fmt.Sprintf("        Sub Thought: %s\n", nextBranch.SubThought))
				builder.WriteString(fmt.Sprintf("          Reasoning: %s\n", nextBranch.Reasoning))
				builder.WriteString(fmt.Sprintf("            Outcome: %s\n", nextBranch.Outcome))
			}
		}
	}

	builder.WriteString(fmt.Sprintf("  Final Conclusion: %s\n", tree.Template.FinalConclusion))
	builder.WriteString("[/TREE OF THOUGHT]")
	return builder.String()
}
