package process

import (
	"encoding/json"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/berrt"
	"github.com/theapemachine/amsh/errnie"
)

type Execution struct {
	Actions []Action `json:"actions" jsonschema:"required; description:The detailedactions to be executed"`
}

type Action struct {
	TeamMember string   `json:"team_member" jsonschema:"required; description:The key of the team member"`
	Prompt     string   `json:"prompt" jsonschema:"required; description:The prompt for the action"`
	SubActions []Action `json:"sub_actions" jsonschema:"description:The sub-actions for the action"`
}

func NewExecution() *Execution {
	return &Execution{}
}

func (execution *Execution) Extract(response string) *Execution {
	jsonStart := strings.Index(response, "```json")
	jsonEnd := strings.LastIndex(response, "```")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil
	}

	json := response[jsonStart+7 : jsonEnd]

	return execution.Unmarshal([]byte(json))
}

func (execution *Execution) Marshal() string {
	buf, err := json.Marshal(execution)

	if err != nil {
		errnie.Error(err)
	}

	return string(buf)
}

func (execution *Execution) Unmarshal(buf []byte) *Execution {
	err := json.Unmarshal(buf, execution)

	if err != nil {
		errnie.Error(err)
	}

	return execution
}

func (execution *Execution) GenerateSchema() string {
	schema := jsonschema.Reflect(&Execution{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

func (execution *Execution) Format() string {
	pretty, _ := json.MarshalIndent(execution, "", "  ")
	return string(pretty)
}

func (execution *Execution) String() {
	berrt.Info("Execution", execution)
}
