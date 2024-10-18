package format

import (
	"encoding/json"
	"fmt"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Managing struct {
	Goals []Goal     `json:"goals" jsonschema:"description=The main goals to be achieved"`
	Plan  []PlanStep `json:"next_steps" jsonschema:"description=The next steps to be taken"`
	Done  bool       `json:"done" jsonschema:"description=Indicates if the management plan is complete;required=true"`
}

func NewManaging() *Managing {
	return &Managing{}
}

func (mp *Managing) Print(data []byte) error {
	if err := errnie.Error(json.Unmarshal(data, mp)); err != nil {
		return err
	}

	fmt.Println(mp.String())
	return nil
}

func (mp Managing) String() string {
	output := utils.Dark("[MANAGEMENT PLAN]") + "\n"

	output += "\t" + utils.Muted("[GOALS]") + "\n"
	for _, goal := range mp.Goals {
		output += "\t\t" + utils.Green("Goal: ") + goal.Description + "\n"
		output += "\t\t" + utils.Yellow("Status: ") + goal.Status + "\n"
	}
	output += "\t" + utils.Muted("[/GOALS]") + "\n"

	output += "\t" + utils.Muted("[PLAN]") + "\n"
	for _, goal := range mp.Plan {
		output += "\t\t" + utils.Green("Goal: ") + goal.Action + "\n"
		output += "\t\t" + utils.Red("Risks: ") + goal.Risks + "\n"
		output += "\t\t" + utils.Yellow("Rewards: ") + goal.Rewards + "\n"
	}
	output += "\t" + utils.Muted("[/PLAN]") + "\n"

	output += "\t" + utils.Red("Done: ") + BoolToString(mp.Done) + "\n"
	output += utils.Dark("[/MANAGEMENT PLAN]") + "\n"
	return output
}

type Goal struct {
	Description string `json:"description" jsonschema:"description=A description of the goal"`
	Status      string `json:"status" jsonschema:"description=The current status of the goal"`
}

type PlanStep struct {
	Action  string `json:"action" jsonschema:"description=The action to be taken"`
	Risks   string `json:"risks" jsonschema:"description=The risks that may be involved with the action"`
	Rewards string `json:"rewards" jsonschema:"description=The rewards that are gained from the action"`
}

// Helper function to convert bool to string
func BoolToString(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
