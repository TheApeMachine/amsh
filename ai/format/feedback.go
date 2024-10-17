package format

import (
	"github.com/theapemachine/amsh/utils"
)

type Verify struct {
	Feedback                   []Feedback `json:"feedback" jsonschema:"title=Feedback,description=The feedback to the reasoning"`
	FinalContinuousImprovement string     `json:"final_continuous_improvement" jsonschema:"title=Final Continuous Improvement,description=The best single change the reasoner could make to improve their performance"`
}

func (v Verify) String() string {
	output := utils.Dark("[VERIFY]") + "\n"

	for _, feedback := range v.Feedback {
		for _, observation := range feedback.Observations {
			output += "\t" + utils.Red("Reasoning Segment: ") + observation.ReasoningSegment + "\n"

			output += "\t" + utils.Muted("[VERIFICATION]") + "\n"
			output += "\t\t" + utils.Yellow("Verification: ") + observation.Verification + "\n"
			output += "\t\t" + utils.Green("Correctness : ") + observation.Correctness + "\n"
			output += "\t" + utils.Muted("[/VERIFICATION]") + "\n"

			output += "\t" + utils.Muted("[IMPROVEMENTS]") + "\n"
			for _, improvement := range observation.Improvements {
				output += "\t\t" + utils.Blue("- ") + improvement + "\n"
			}
			output += "\t" + utils.Muted("[/IMPROVEMENTS]") + "\n"
		}
	}

	output += "\t" + utils.Blue("Final Continuous Improvement: ") + v.FinalContinuousImprovement + "\n"
	output += utils.Dark("[/VERIFY]") + "\n"
	return output
}

type Feedback struct {
	Observations []Observation `json:"observations" jsonschema:"title=Observations,description=The observations you have while using the strategy"`
}

type Observation struct {
	ReasoningSegment string   `json:"reasoning_segment" jsonschema:"title=Reasoning Segment,description=The reasoning segment from the reasoning you are reviewing"`
	Verification     string   `json:"verification" jsonschema:"title=Verification,description=The verification of the reasoning segment"`
	Correctness      string   `json:"correctness" jsonschema:"title=Correctness,description=The correctness of the reasoning segment"`
	Improvements     []string `json:"improvements" jsonschema:"title=Improvements,description=Improvements in approach the reasoner could make"`
}
