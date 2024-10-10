package format

import (
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type RolePlaySimulation struct {
	Role        string `json:"role"`
	Interaction string `json:"interaction"`
	Response    string `json:"response"`
}

func NewRolePlaySimulation() *RolePlaySimulation {
	errnie.Trace()
	return &RolePlaySimulation{}
}

func (rps *RolePlaySimulation) FinalAnswer() string {
	return rps.Response
}

func (rps *RolePlaySimulation) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(rps)
}

func (rps *RolePlaySimulation) ToString() string {
	builder := strings.Builder{}
	builder.WriteString(dark("  [ROLE PLAY SIMULATION]\n"))
	builder.WriteString(fmt.Sprintf(blue("    Role: %s\n"), highlight(rps.Role)))
	builder.WriteString(fmt.Sprintf(green("    Interaction: %s\n"), highlight(rps.Interaction)))
	builder.WriteString(fmt.Sprintf(yellow("    Response: %s\n"), highlight(rps.Response)))
	builder.WriteString(dark("  [/ROLE PLAY SIMULATION]\n"))
	return builder.String()
}
