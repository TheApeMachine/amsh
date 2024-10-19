package format

import (
	"encoding/json"
	"fmt"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Executing struct {
	TaskName      string               `json:"task_name" jsonschema:"description=The name of the task being executed"`
	Description   string               `json:"description" jsonschema:"description=A detailed description of the task"`
	Steps         []ExecuteStep        `json:"steps" jsonschema:"description=The steps required to complete the task"`
	CurrentStep   int                  `json:"current_step" jsonschema:"description=The index of the current step being executed"`
	Resources     []Resource           `json:"resources" jsonschema:"description=Resources required for task execution"`
	Progress      float64              `json:"progress" jsonschema:"description=Overall progress of the task (0.0 to 1.0)"`
	QualityChecks []QualityCheck       `json:"quality_checks" jsonschema:"description=Quality checks performed during execution"`
	Challenges    []ExecuteChallenge   `json:"challenges" jsonschema:"description=Challenges encountered during execution"`
	Improvements  []ExecuteImprovement `json:"improvements" jsonschema:"description=Improvements made during execution"`
	Done          bool                 `json:"done" jsonschema:"description=Indicates if the execution plan is complete;required=true"`
}

func NewExecuting() *Executing {
	return &Executing{}
}

func (ep *Executing) Print(data []byte) (isDone bool, err error) {
	if err := errnie.Error(json.Unmarshal(data, ep)); err != nil {
		return false, err
	}

	fmt.Println(ep.String())
	return ep.Done, nil
}

func (ep Executing) String() string {
	output := utils.Dark("[EXECUTION PLAN]") + "\n"

	output += "\t" + utils.Blue("Task Name: ") + ep.TaskName + "\n"
	output += "\t" + utils.Green("Description: ") + ep.Description + "\n"

	output += "\t" + utils.Muted("[STEPS]") + "\n"
	for i, step := range ep.Steps {
		status := "Pending"
		if i < ep.CurrentStep {
			status = "Completed"
		} else if i == ep.CurrentStep {
			status = "In Progress"
		}
		output += "\t\t" + utils.Blue(fmt.Sprintf("Step %d: ", i+1)) + step.Description + "\n"
		output += "\t\t" + utils.Yellow("Status: ") + status + "\n"
		output += "\t\t" + utils.Green("Details: ") + step.Details + "\n"
	}
	output += "\t" + utils.Muted("[/STEPS]") + "\n"

	output += "\t" + utils.Muted("[RESOURCES]") + "\n"
	for _, resource := range ep.Resources {
		output += "\t\t" + utils.Blue("Name: ") + resource.Name + "\n"
		output += "\t\t" + utils.Green("Type: ") + resource.Type + "\n"
		output += "\t\t" + utils.Yellow("Status: ") + resource.Status + "\n"
	}
	output += "\t" + utils.Muted("[/RESOURCES]") + "\n"

	output += "\t" + utils.Muted("[QUALITY CHECKS]") + "\n"
	for _, check := range ep.QualityChecks {
		output += "\t\t" + utils.Blue("Check: ") + check.Description + "\n"
		output += "\t\t" + utils.Green("Result: ") + check.Result + "\n"
	}
	output += "\t" + utils.Muted("[/QUALITY CHECKS]") + "\n"

	output += "\t" + utils.Muted("[CHALLENGES]") + "\n"
	for _, challenge := range ep.Challenges {
		output += "\t\t" + utils.Red("Challenge: ") + challenge.Description + "\n"
		output += "\t\t" + utils.Yellow("Resolution: ") + challenge.Resolution + "\n"
	}
	output += "\t" + utils.Muted("[/CHALLENGES]") + "\n"

	output += "\t" + utils.Muted("[IMPROVEMENTS]") + "\n"
	for _, improvement := range ep.Improvements {
		output += "\t\t" + utils.Blue("Improvement: ") + improvement.Description + "\n"
		output += "\t\t" + utils.Green("Impact: ") + improvement.Impact + "\n"
	}
	output += "\t" + utils.Muted("[/IMPROVEMENTS]") + "\n"

	output += "\t" + utils.Blue("Progress: ") + FloatToString(ep.Progress) + "\n"
	output += "\t" + utils.Red("Done: ") + BoolToString(ep.Done) + "\n"
	output += utils.Dark("[/EXECUTION PLAN]") + "\n"
	return output
}

type ExecuteStep struct {
	Description string `json:"description" jsonschema:"description=A description of the step"`
	Details     string `json:"details" jsonschema:"description=Detailed information about the step"`
}

type Resource struct {
	Name   string `json:"name" jsonschema:"description=The name of the resource"`
	Type   string `json:"type" jsonschema:"description=The type of the resource"`
	Status string `json:"status" jsonschema:"description=The current status of the resource"`
}

type QualityCheck struct {
	Description string `json:"description" jsonschema:"description=A description of the quality check"`
	Result      string `json:"result" jsonschema:"description=The result of the quality check"`
}

type ExecuteChallenge struct {
	Description string `json:"description" jsonschema:"description=A description of the challenge encountered"`
	Resolution  string `json:"resolution" jsonschema:"description=The resolution or current status of the challenge"`
}

type ExecuteImprovement struct {
	Description string `json:"description" jsonschema:"description=A description of the improvement made"`
	Impact      string `json:"impact" jsonschema:"description=The impact of the improvement on the task execution"`
}
