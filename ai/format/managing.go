package format

import (
	"encoding/json"
	"fmt"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Managing struct {
	Goals            []Goal    `json:"goals" jsonschema:"description=The main goals to be achieved"`
	Roadmap          []Roadmap `json:"roadmap" jsonschema:"description=The roadmap to be followed"`
	RequestIteration bool      `json:"request_iteration" jsonschema:"description=Request another iteration to continue your current task, or make tool calls, before handing it off to another agent"`
}

func NewManaging() *Managing {
	return &Managing{}
}

func (mp *Managing) Print(data []byte) (isDone bool, err error) {
	if err := errnie.Error(json.Unmarshal(data, mp)); err != nil {
		return false, err
	}

	fmt.Println(mp.String())
	return !mp.RequestIteration, nil
}

func (mp Managing) String() string {
	output := utils.Dark("[PLAN]") + "\n"

	for _, goal := range mp.Goals {
		output += "\t" + utils.Muted("[GOAL]") + "\n"
		output += "\t\t" + utils.Blue("Description: ") + goal.Description + "\n"
		for _, requirement := range goal.Requirements {
			output += "\t\t" + utils.Muted("[REQUIREMENT]") + "\n"
			output += "\t\t\t" + utils.Red("Description: ") + requirement.Description + "\n"
			output += "\t\t\t" + utils.Yellow("Value: ") + requirement.Value + "\n"
			output += "\t\t" + utils.Muted("[/REQUIREMENT]") + "\n"
		}
		for _, obstacle := range goal.Obstacles {
			output += "\t\t" + utils.Muted("[OBSTACLE]") + "\n"
			output += "\t\t\t" + utils.Red("Description: ") + obstacle.Description + "\n"
			output += "\t\t\t" + utils.Yellow("Impact: ") + obstacle.Impact + "\n"
			output += "\t\t\t" + utils.Green("Solution: ") + obstacle.Solution + "\n"
			output += "\t\t" + utils.Muted("[/OBSTACLE]") + "\n"
		}
		output += "\t" + utils.Muted("[/GOAL]") + "\n"
	}

	for _, roadmap := range mp.Roadmap {
		output += "\t" + utils.Muted("[ROADMAP]") + "\n"
		output += "\t\t" + utils.Blue("Description: ") + roadmap.Description + "\n"
		for _, epic := range roadmap.Epics {
			output += "\t\t" + utils.Muted("[EPIC]") + "\n"
			output += "\t\t\t" + utils.Red("Description: ") + epic.Description + "\n"
			output += "\t\t\t" + utils.Yellow("Value: ") + epic.Value + "\n"
			for _, story := range epic.UserStories {
				output += "\t\t\t" + utils.Muted("[USER STORY]") + "\n"
				output += "\t\t\t\t" + utils.Red("Description: ") + story.Description + "\n"
				output += "\t\t\t\t" + utils.Yellow("Value: ") + story.Value + "\n"
				for _, subtask := range story.SubTasks {
					output += "\t\t\t\t" + utils.Muted("[SUBTASK]") + "\n"
					output += "\t\t\t\t\t" + utils.Red("Description: ") + subtask.Description + "\n"
					output += "\t\t\t\t\t" + utils.Yellow("Value: ") + subtask.Value + "\n"
					output += "\t\t\t\t\t" + utils.Green("Workload: ") + subtask.Workload + "\n"
					output += "\t\t\t\t" + utils.Muted("[/SUBTASK]") + "\n"
				}
				output += "\t\t\t" + utils.Muted("[/USER STORY]") + "\n"
			}
			output += "\t\t" + utils.Muted("[/EPIC]") + "\n"
		}
		output += "\t" + utils.Muted("[/ROADMAP]") + "\n\n"
	}

	output += "\t" + utils.Blue("Request Iteration: ") + BoolToString(mp.RequestIteration) + "\n"
	output += utils.Dark("[/PLAN]")
	return output
}

type Goal struct {
	Description  string        `json:"description" jsonschema:"description=A description of the goal"`
	Requirements []Requirement `json:"requirements" jsonschema:"description=The requirements that the goal needs to be achieved"`
	Obstacles    []Obstacle    `json:"obstacles" jsonschema:"description=The obstacles that the goal needs to overcome"`
}

type Requirement struct {
	Description string `json:"description" jsonschema:"description=A description of the requirement"`
	Value       string `json:"value" jsonschema:"description=The value that the requirement provides"`
}

type Obstacle struct {
	Description string `json:"description" jsonschema:"description=A description of the obstacle"`
	Impact      string `json:"impact" jsonschema:"description=The impact that the obstacle has on the goal;enum=high;enum=medium;enum=low"`
	Solution    string `json:"solution" jsonschema:"description=A description of the solution, work around, or mitigation strategy to overcome the obstacle"`
}

type Roadmap struct {
	Description string `json:"description" jsonschema:"description=A description of the roadmap and its main purpose"`
	Epics       []Epic `json:"epics" jsonschema:"description=The epics that the roadmap belongs to"`
}

type Epic struct {
	Description string      `json:"description" jsonschema:"description=A description of the epic"`
	Value       string      `json:"value" jsonschema:"description=The value that the epic provides"`
	UserStories []UserStory `json:"user_stories" jsonschema:"description=A list of user stories that describe the epic"`
}

type UserStory struct {
	Description string    `json:"description" jsonschema:"description=A Gherkin-style description of the user story"`
	Value       string    `json:"value" jsonschema:"description=The value that the user story provides"`
	SubTasks    []SubTask `json:"sub_tasks" jsonschema:"description=The sub tasks that make up the user story"`
}

type SubTask struct {
	Description string `json:"description" jsonschema:"description=A description of the sub task"`
	Value       string `json:"value" jsonschema:"description=The value that the sub task provides"`
	Workload    string `json:"workload" jsonschema:"description=The workload that is required to be completed to achieve the goal;enum=reasoning;enum=executing;enum=researching;enum=other"`
}

// Helper function to convert bool to string
func BoolToString(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
