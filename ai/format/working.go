package format

import (
	"encoding/json"
	"fmt"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Working struct {
	Objectives []WorkingObjective `json:"objectives" jsonschema:"description=List of objectives for the task"`
	Steps      []WorkStep         `json:"steps" jsonschema:"description=Steps to complete the task"`
	Resources  []WorkingResource  `json:"resources" jsonschema:"description=Resources used or needed for the task"`
	Outputs    []Output           `json:"outputs" jsonschema:"description=Outputs generated during the task"`
	Notes      []Note             `json:"notes" jsonschema:"description=Additional notes or observations"`
	Status     string             `json:"status" jsonschema:"description=Current status of the task"`
	Progress   float64            `json:"progress" jsonschema:"description=Overall progress of the task (0.0 to 1.0)"`
	Done       bool               `json:"done" jsonschema:"description=Indicates if the work plan is complete;required=true"`
}

func NewWorking() *Working {
	return &Working{}
}

func (wp *Working) Print(data []byte) error {
	if err := errnie.Error(json.Unmarshal(data, wp)); err != nil {
		return err
	}

	fmt.Println(wp.String())
	return nil
}

func (wp Working) String() string {
	output := utils.Dark("[WORK PLAN]") + "\n"

	output += "\t" + utils.Muted("[OBJECTIVES]") + "\n"
	for _, obj := range wp.Objectives {
		output += "\t\t" + utils.Blue("Objective: ") + obj.Description + "\n"
		output += "\t\t" + utils.Green("Status: ") + obj.Status + "\n"
	}
	output += "\t" + utils.Muted("[/OBJECTIVES]") + "\n"

	output += "\t" + utils.Muted("[STEPS]") + "\n"
	for i, step := range wp.Steps {
		output += "\t\t" + utils.Blue(fmt.Sprintf("Step %d: ", i+1)) + step.Description + "\n"
		output += "\t\t" + utils.Green("Status: ") + step.Status + "\n"
		if step.Result != "" {
			output += "\t\t" + utils.Yellow("Result: ") + step.Result + "\n"
		}
		output += "\n"
	}
	output += "\t" + utils.Muted("[/STEPS]") + "\n"

	output += "\t" + utils.Muted("[RESOURCES]") + "\n"
	for _, resource := range wp.Resources {
		output += "\t\t" + utils.Blue("Name: ") + resource.Name + "\n"
		output += "\t\t" + utils.Green("Type: ") + resource.Type + "\n"
		output += "\t\t" + utils.Yellow("Status: ") + resource.Status + "\n"
	}
	output += "\t" + utils.Muted("[/RESOURCES]") + "\n"

	output += "\t" + utils.Muted("[OUTPUTS]") + "\n"
	for _, out := range wp.Outputs {
		output += "\t\t" + utils.Blue("Type: ") + out.Type + "\n"
		output += "\t\t" + utils.Green("Description: ") + out.Description + "\n"
		output += "\t\t" + utils.Yellow("Location: ") + out.Location + "\n"
	}
	output += "\t" + utils.Muted("[/OUTPUTS]") + "\n"

	output += "\t" + utils.Muted("[NOTES]") + "\n"
	for _, note := range wp.Notes {
		output += "\t\t" + utils.Blue("Note: ") + note.Content + "\n"
		output += "\t\t" + utils.Green("Timestamp: ") + note.Timestamp + "\n"
	}
	output += "\t" + utils.Muted("[/NOTES]") + "\n"

	output += "\t" + utils.Blue("Status: ") + wp.Status + "\n"
	output += "\t" + utils.Green("Progress: ") + FloatToString(wp.Progress) + "\n"
	output += "\t" + utils.Red("Done: ") + BoolToString(wp.Done) + "\n"
	output += utils.Dark("[/WORK PLAN]") + "\n"
	return output
}

type WorkingObjective struct {
	Description string `json:"description" jsonschema:"description=Description of the objective"`
	Status      string `json:"status" jsonschema:"description=Current status of the objective"`
}

type WorkStep struct {
	Description string `json:"description" jsonschema:"description=Description of the step"`
	Status      string `json:"status" jsonschema:"description=Current status of the step"`
	Result      string `json:"result" jsonschema:"description=Result or outcome of the step"`
}

type WorkingResource struct {
	Name   string `json:"name" jsonschema:"description=Name of the resource"`
	Type   string `json:"type" jsonschema:"description=Type of the resource"`
	Status string `json:"status" jsonschema:"description=Current status of the resource"`
}

type Output struct {
	Type        string `json:"type" jsonschema:"description=Type of the output"`
	Description string `json:"description" jsonschema:"description=Description of the output"`
	Location    string `json:"location" jsonschema:"description=Location or reference to the output"`
}

type Note struct {
	Content   string `json:"content" jsonschema:"description=Content of the note"`
	Timestamp string `json:"timestamp" jsonschema:"description=Timestamp when the note was created"`
}
