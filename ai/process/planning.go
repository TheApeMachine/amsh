package process

import (
	"encoding/json"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/berrt"
	"github.com/theapemachine/amsh/errnie"
)

type Planning struct {
	Goals          []Goal           `json:"goals" jsonschema:"required; description:The goals to achieve"`
	Teams          []Team           `json:"teams" jsonschema:"required; description:The teams to use to achieve the goals"`
	Dependencies   []Dependency     `json:"dependencies" jsonschema:"description:The dependencies between steps"`
	FinalSynthesis []FinalSynthesis `json:"final_synthesis" jsonschema:"required; description:The final synthesis of the goals, once all goals have been achieved"`
	FinalResponse  string           `json:"final_response" jsonschema:"description:The final response to the user, once all goals have been achieved"`
}

type Goal struct {
	Goal  string `json:"goal" jsonschema:"required; description:The goal to achieve"`
	Steps []Step `json:"steps" jsonschema:"required; description:The steps to achieve the goal"`
}

type Team struct {
	TeamKey string  `json:"team_key" jsonschema:"required; description:The key of the team"`
	Agents  []Agent `json:"agents" jsonschema:"required; description:The agents in the team"`
}

type Agent struct {
	RoleKey      string `json:"role_key" jsonschema:"required; description:The key of the role"`
	SystemPrompt string `json:"system_prompt" jsonschema:"required; description:A comprehensive system prompt for the agent"`
}

type Dependency struct {
	FromStep string `json:"from_step" jsonschema:"required; description:The key of the step that must be completed before the to step"`
	ToStep   string `json:"to_step" jsonschema:"required; description:The key of the step that must be completed after the from step"`
}

type Step struct {
	StepKey  string     `json:"step_key" jsonschema:"required; description:The key of the step"`
	TeamKey  string     `json:"team_key" jsonschema:"required; description:The key of the team"`
	Prompt   string     `json:"prompt" jsonschema:"required; description:The prompt for the step"`
	Inputs   []Artifact `json:"inputs" jsonschema:"description:The inputs for the step"`
	Outputs  []Artifact `json:"outputs" jsonschema:"description:The outputs for the step"`
	SubSteps []Step     `json:"sub_steps" jsonschema:"description:The sub-steps for the step"`
}

type Artifact struct {
	Key   string `json:"key" jsonschema:"description:The key of the artifact"`
	Value string `json:"value" jsonschema:"description:The value of the artifact"`
}

type FinalSynthesis struct {
	StepKeys           []string `json:"step_keys" jsonschema:"description:The keys of the steps that must be completed before the final synthesis"`
	ConstructionMethod string   `json:"construction_method" jsonschema:"description:The method used to construct the final synthesis"`
}

func NewPlanning() *Planning {
	return &Planning{}
}

/*
Extract the planning JSON from the response.
*/
func (planning *Planning) Extract(response string) *Planning {
	jsonStart := strings.Index(response, "```json")
	jsonEnd := strings.LastIndex(response, "```")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil
	}

	json := response[jsonStart+7 : jsonEnd]

	return planning.Unmarshal([]byte(json))
}

func (planning *Planning) Marshal() string {
	buf, err := json.Marshal(planning)

	if err != nil {
		errnie.Error(err)
	}

	return string(buf)
}

func (planning *Planning) Unmarshal(buf []byte) *Planning {
	err := json.Unmarshal(buf, planning)

	if err != nil {
		errnie.Error(err)
	}

	return planning
}

func (planning *Planning) GenerateSchema() string {
	schema := jsonschema.Reflect(&Planning{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}

	return string(out)
}

/*
Format the process as a pretty-printed JSON string.
*/
func (planning *Planning) Format() string {
	pretty, _ := json.MarshalIndent(planning, "", "  ")
	return string(pretty)
}

/*
String returns a human-readable string representation of the process.
*/
func (planning *Planning) String() {
	berrt.Info("Planning", planning)
}
