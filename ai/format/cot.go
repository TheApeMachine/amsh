package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type ChainOfThought struct {
	Steps []struct {
		Thought   string `json:"thought"`
		Reasoning string `json:"reasoning"`
		NextStep  string `json:"next_step"`
	}
	Action string `json:"action"`
	Result string `json:"result"`
}

func NewChainOfThought() *ChainOfThought {
	errnie.Trace()
	return &ChainOfThought{}
}

func (chain *ChainOfThought) FinalAnswer() string {
	return chain.Result
}

func (chain *ChainOfThought) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(chain)
}

func (chain *ChainOfThought) ToString() string {
	out := []string{}
	out = append(out, dark("  [CHAIN OF THOUGHT]"))
	for _, step := range chain.Steps {
		out = append(out, muted("    [STEP]"))
		out = append(out, red("        Thought: ")+highlight(step.Thought))
		out = append(out, yellow("      Reasoning: ")+highlight(step.Reasoning))
		out = append(out, green("      Next Step: ")+highlight(step.NextStep))
		out = append(out, muted("    [/STEP]"))
	}
	out = append(out, blue("    Action: ")+highlight(chain.Action))
	out = append(out, blue("    Result: ")+highlight(chain.Result))
	out = append(out, dark("  [/CHAIN OF THOUGHT]"))
	return strings.Join(out, "\n")
}
