package format

import (
	"fmt"

	"github.com/theapemachine/amsh/utils"
)

type Verifying struct {
	OriginalReasoning string            `json:"original_reasoning" jsonschema:"description=The original reasoning being verified"`
	Evaluation        Evaluation        `json:"evaluation" jsonschema:"description=Critical evaluation of the reasoning"`
	Feedback          []FeedbackPoint   `json:"feedback" jsonschema:"description=Detailed feedback on the reasoning"`
	Improvements      []Improvement     `json:"improvements" jsonschema:"description=Suggested improvements to the reasoning"`
	Challenges        []VerifyChallenge `json:"challenges" jsonschema:"description=Challenges to the original reasoning"`
	VerificationScore float64           `json:"verification_score" jsonschema:"description=Overall score of the verification (0.0 to 1.0)"`
	Done              bool              `json:"done" jsonschema:"description=Indicates if the verification plan is complete;required=true"`
}

func (vp Verifying) String() string {
	output := utils.Dark("[VERIFICATION PLAN]") + "\n"

	output += "\t" + utils.Blue("Original Reasoning: ") + vp.OriginalReasoning + "\n\n"

	output += "\t" + utils.Muted("[EVALUATION]") + "\n"
	output += "\t\t" + utils.Green("Strengths: ") + vp.Evaluation.Strengths + "\n"
	output += "\t\t" + utils.Red("Weaknesses: ") + vp.Evaluation.Weaknesses + "\n"
	output += "\t\t" + utils.Yellow("Overall Assessment: ") + vp.Evaluation.OverallAssessment + "\n"
	output += "\t" + utils.Muted("[/EVALUATION]") + "\n\n"

	output += "\t" + utils.Muted("[FEEDBACK]") + "\n"
	for _, point := range vp.Feedback {
		output += "\t\t" + utils.Blue("Point: ") + point.Point + "\n"
		output += "\t\t" + utils.Green("Explanation: ") + point.Explanation + "\n\n"
	}
	output += "\t" + utils.Muted("[/FEEDBACK]") + "\n\n"

	output += "\t" + utils.Muted("[IMPROVEMENTS]") + "\n"
	for _, improvement := range vp.Improvements {
		output += "\t\t" + utils.Blue("Suggestion: ") + improvement.Suggestion + "\n"
		output += "\t\t" + utils.Green("Rationale: ") + improvement.Rationale + "\n\n"
	}
	output += "\t" + utils.Muted("[/IMPROVEMENTS]") + "\n\n"

	output += "\t" + utils.Muted("[CHALLENGES]") + "\n"
	for _, challenge := range vp.Challenges {
		output += "\t\t" + utils.Red("Challenge: ") + challenge.Description + "\n"
		output += "\t\t" + utils.Yellow("Potential Impact: ") + challenge.PotentialImpact + "\n\n"
	}
	output += "\t" + utils.Muted("[/CHALLENGES]") + "\n\n"

	output += "\t" + utils.Blue("Verification Score: ") + FloatToString(vp.VerificationScore) + "\n"
	output += "\t" + utils.Red("Done: ") + BoolToString(vp.Done) + "\n"
	output += utils.Dark("[/VERIFICATION PLAN]") + "\n"
	return output
}

type Evaluation struct {
	Strengths         string `json:"strengths" jsonschema:"description=Strengths of the original reasoning"`
	Weaknesses        string `json:"weaknesses" jsonschema:"description=Weaknesses of the original reasoning"`
	OverallAssessment string `json:"overall_assessment" jsonschema:"description=Overall assessment of the reasoning"`
}

type FeedbackPoint struct {
	Point       string `json:"point" jsonschema:"description=A specific point of feedback"`
	Explanation string `json:"explanation" jsonschema:"description=Detailed explanation of the feedback point"`
}

type Improvement struct {
	Suggestion string `json:"suggestion" jsonschema:"description=A suggested improvement to the reasoning"`
	Rationale  string `json:"rationale" jsonschema:"description=The rationale behind the suggested improvement"`
}

type VerifyChallenge struct {
	Description     string `json:"description" jsonschema:"description=A challenge to the original reasoning"`
	PotentialImpact string `json:"potential_impact" jsonschema:"description=The potential impact if this challenge is valid"`
}

func FloatToString(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
