package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type HypothesisTesting struct {
	Hypothesis string `json:"hypothesis"`
	Test       string `json:"test"`
	Outcome    string `json:"outcome"`
	Conclusion string `json:"conclusion"`
}

func NewHypothesisTesting() *HypothesisTesting {
	errnie.Trace()
	return &HypothesisTesting{}
}

func (ht *HypothesisTesting) FinalAnswer() string {
	return ht.Conclusion
}

func (ht *HypothesisTesting) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(ht)
}

func (ht *HypothesisTesting) ToString() string {
	out := []string{}
	out = append(out, dark("  [HYPOTHESIS TESTING]"))
	out = append(out, blue("    Hypothesis: ")+highlight(ht.Hypothesis))
	out = append(out, green("    Test: ")+highlight(ht.Test))
	out = append(out, yellow("    Outcome: ")+highlight(ht.Outcome))
	out = append(out, red("    Conclusion: ")+highlight(ht.Conclusion))
	out = append(out, dark("  [/HYPOTHESIS TESTING]"))
	return strings.Join(out, "\n")
}
