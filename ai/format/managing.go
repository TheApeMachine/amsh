package format

import (
	"github.com/theapemachine/amsh/utils"
)

type Managing struct {
	Overview    string        `json:"overview" jsonschema:"description=A high-level overview of the current plan"`
	Goals       []Goal        `json:"goals" jsonschema:"description=The main goals to be achieved"`
	Assignments []Assignment  `json:"assignments" jsonschema:"description=Task assignments for different agents"`
	Progress    ProgressTrack `json:"progress" jsonschema:"description=Overall progress of the plan"`
	NextSteps   []Step        `json:"next_steps" jsonschema:"description=The next steps to be taken"`
	Done        bool          `json:"done" jsonschema:"description=Indicates if the management plan is complete;required=true"`
}

func (mp Managing) String() string {
	output := utils.Dark("[MANAGEMENT PLAN]") + "\n"

	output += "\t" + utils.Blue("Overview: ") + mp.Overview + "\n\n"

	output += "\t" + utils.Muted("[GOALS]") + "\n"
	for _, goal := range mp.Goals {
		output += "\t\t" + utils.Green("Goal: ") + goal.Description + "\n"
		output += "\t\t" + utils.Yellow("Status: ") + goal.Status + "\n\n"
	}
	output += "\t" + utils.Muted("[/GOALS]") + "\n\n"

	output += "\t" + utils.Muted("[ASSIGNMENTS]") + "\n"
	for _, assignment := range mp.Assignments {
		output += "\t\t" + utils.Red("Agent: ") + assignment.AgentID + "\n"
		output += "\t\t" + utils.Green("Task: ") + assignment.Task + "\n"
		output += "\t\t" + utils.Yellow("Status: ") + assignment.Status + "\n\n"
	}
	output += "\t" + utils.Muted("[/ASSIGNMENTS]") + "\n\n"

	output += "\t" + utils.Muted("[PROGRESS]") + "\n"
	output += "\t\t" + utils.Green("Overall: ") + mp.Progress.Overall + "\n"
	output += "\t\t" + utils.Yellow("Blockers: ") + mp.Progress.Blockers + "\n"
	output += "\t" + utils.Muted("[/PROGRESS]") + "\n\n"

	output += "\t" + utils.Muted("[NEXT STEPS]") + "\n"
	for _, step := range mp.NextSteps {
		output += "\t\t" + utils.Blue("Action: ") + step.Action + "\n"
		output += "\t\t" + utils.Green("Motivation: ") + step.Motivation + "\n\n"
	}
	output += "\t" + utils.Muted("[/NEXT STEPS]") + "\n"

	output += "\t" + utils.Red("Done: ") + BoolToString(mp.Done) + "\n"
	output += utils.Dark("[/MANAGEMENT PLAN]") + "\n"
	return output
}

type Goal struct {
	Description string `json:"description" jsonschema:"description=A description of the goal"`
	Status      string `json:"status" jsonschema:"description=The current status of the goal"`
}

type Assignment struct {
	AgentID string `json:"agent_id" jsonschema:"description=The ID of the agent assigned to the task"`
	Task    string `json:"task" jsonschema:"description=The task assigned to the agent"`
	Status  string `json:"status" jsonschema:"description=The current status of the assignment"`
}

type ProgressTrack struct {
	Overall  string `json:"overall" jsonschema:"description=Overall progress of the plan"`
	Blockers string `json:"blockers" jsonschema:"description=Any blockers or issues hindering progress"`
}

// Helper function to convert bool to string
func BoolToString(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
