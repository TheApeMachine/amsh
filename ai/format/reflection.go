package format

import (
	"encoding/json"
	"fmt"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type SelfReflection struct {
	Considerations  []Consideration `json:"considerations" jsonschema:"title=Considerations,description=Considerations about how the feedback connects to the method used to execute the task."`
	PromptAdditions string          `json:"prompt_additions" jsonschema:"title=Prompt Additions,description=Additions to the system prompt that will help the AI to improve its approach and performance."`
}

func NewSelfReflection() *SelfReflection {
	return &SelfReflection{}
}

func (sr *SelfReflection) Print(data []byte) (isDone bool, err error) {
	if err := errnie.Error(json.Unmarshal(data, sr)); err != nil {
		return false, err
	}

	fmt.Println(sr.String())
	return true, nil
}

func (sr SelfReflection) String() string {
	output := utils.Dark("[REFLECTION]") + "\n"

	for _, consideration := range sr.Considerations {
		output += "\t" + utils.Muted("[CONSIDERATION]") + "\n"
		output += "\t\t" + utils.Red("Task Method: ") + consideration.TaskMethod + "\n"
		output += "\t\t" + utils.Yellow("Feedback Received: ") + consideration.FeedbackReceived + "\n"
		output += "\t\t" + utils.Green("Improvement: ") + consideration.Improvement + "\n"
		output += "\t" + utils.Muted("[/CONSIDERATION]") + "\n"
	}

	output += "\t" + utils.Blue("Prompt Additions: ") + sr.PromptAdditions + "\n"
	output += utils.Muted("[/REFLECTION]") + "\n"

	return output
}

type Consideration struct {
	TaskMethod       string `json:"task_method" jsonschema:"title=Task Method,description=A method used while executing the task."`
	FeedbackReceived string `json:"feedback_received" jsonschema:"title=Feedback Received,description=Feedback you received that connects to the method."`
	Improvement      string `json:"improvement" jsonschema:"title=Improvement,description=An improvement to the method that will help you to improve your approach and performance."`
}
