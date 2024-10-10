package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type TreeOfThought struct {
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

func NewTreeOfThought() *TreeOfThought {
	errnie.Trace()
	return &TreeOfThought{}
}

func (tot *TreeOfThought) FinalAnswer() string {
	return tot.FinalConclusion
}

func (tot *TreeOfThought) Schema() (*jsonschema.Definition, error) {
	errnie.Trace()
	return jsonschema.GenerateSchemaForType(tot)
}

func (tot *TreeOfThought) ToString() string {
	out := []string{}
	out = append(out, dark("  [TREE OF THOUGHT]"))
	out = append(out, red("    Root Thought: ")+highlight(tot.RootThought))

	// Follow all branches.
	for _, branch := range tot.Branches {
		out = append(out, yellow("    Thought: ")+highlight(branch.Thought))
		out = append(out, green("    Reasoning: ")+highlight(branch.Reasoning))
		for _, subBranch := range branch.Branches {
			out = append(out, red("      Sub Thought: ")+highlight(subBranch.SubThought))
			out = append(out, yellow("      Reasoning: ")+highlight(subBranch.Reasoning))
			out = append(out, green("      Outcome: ")+highlight(subBranch.Outcome))
			for _, nextBranch := range subBranch.NextBranches {
				out = append(out, red("      Sub Thought: ")+highlight(nextBranch.SubThought))
				out = append(out, yellow("      Reasoning: ")+highlight(nextBranch.Reasoning))
				out = append(out, green("      Outcome: ")+highlight(nextBranch.Outcome))
			}
		}
	}

	out = append(out, blue("    Final Conclusion: ")+highlight(tot.FinalConclusion))
	out = append(out, dark("  [/TREE OF THOUGHT]"))
	return strings.Join(out, "\n")
}
