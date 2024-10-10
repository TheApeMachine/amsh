package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type RiskAssessment struct {
	RiskFactor  string `json:"risk_factor"`
	Probability string `json:"probability"`
	Impact      string `json:"impact"`
	Mitigation  string `json:"mitigation"`
}

func NewRiskAssessment() *RiskAssessment {
	errnie.Trace()
	return &RiskAssessment{}
}

func (ra *RiskAssessment) FinalAnswer() string {
	return ra.Mitigation
}

func (ra *RiskAssessment) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(ra)
}

func (ra *RiskAssessment) ToString() string {
	out := []string{}
	out = append(out, dark("  [RISK ASSESSMENT]"))
	out = append(out, blue("    Risk Factor: ")+highlight(ra.RiskFactor))
	out = append(out, yellow("    Probability: ")+highlight(ra.Probability))
	out = append(out, red("    Impact: ")+highlight(ra.Impact))
	out = append(out, green("    Mitigation: ")+highlight(ra.Mitigation))
	out = append(out, dark("  [/RISK ASSESSMENT]"))
	return strings.Join(out, "\n")
}
