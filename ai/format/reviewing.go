package format

import (
	"encoding/json"
	"fmt"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Reviewing struct {
	ProjectName         string                 `json:"project_name" jsonschema:"description=The name of the project or task being reviewed"`
	Overview            string                 `json:"overview" jsonschema:"description=A brief overview of the review process"`
	Sections            []ReviewSection        `json:"sections" jsonschema:"description=Different sections or aspects of the project being reviewed"`
	Strengths           []ReviewStrength       `json:"strengths" jsonschema:"description=Identified strengths of the project"`
	AreasForImprovement []ReviewImprovement    `json:"areas_for_improvement" jsonschema:"description=Areas identified for improvement"`
	Questions           []ReviewQuestion       `json:"questions" jsonschema:"description=Questions raised during the review process"`
	Recommendations     []ReviewRecommendation `json:"recommendations" jsonschema:"description=Recommendations based on the review"`
	OverallRating       float64                `json:"overall_rating" jsonschema:"description=Overall rating of the project (0.0 to 5.0)"`
	Done                bool                   `json:"done" jsonschema:"description=Indicates if the review plan is complete;required=true"`
}

func NewReviewing() *Reviewing {
	return &Reviewing{}
}

func (rp *Reviewing) Print(data []byte) error {
	if err := errnie.Error(json.Unmarshal(data, rp)); err != nil {
		return err
	}

	fmt.Println(rp.String())
	return nil
}

func (rp Reviewing) String() string {
	output := utils.Dark("[REVIEW PLAN]") + "\n"

	output += "\t" + utils.Blue("Project Name: ") + rp.ProjectName + "\n"
	output += "\t" + utils.Green("Overview: ") + rp.Overview + "\n"

	output += "\t" + utils.Muted("[SECTIONS]") + "\n"
	for _, section := range rp.Sections {
		output += "\t\t" + utils.Blue("Section: ") + section.Name + "\n"
		output += "\t\t" + utils.Green("Comments: ") + section.Comments + "\n"
		output += "\t\t" + utils.Yellow("Rating: ") + FloatToString(section.Rating) + "/5.0\n"
	}
	output += "\t" + utils.Muted("[/SECTIONS]") + "\n"

	output += "\t" + utils.Muted("[STRENGTHS]") + "\n"
	for _, strength := range rp.Strengths {
		output += "\t\t" + utils.Blue("Strength: ") + strength.Description + "\n"
		output += "\t\t" + utils.Green("Impact: ") + strength.Impact + "\n"
	}
	output += "\t" + utils.Muted("[/STRENGTHS]") + "\n"

	output += "\t" + utils.Muted("[AREAS FOR IMPROVEMENT]") + "\n"
	for _, improvement := range rp.AreasForImprovement {
		output += "\t\t" + utils.Red("Area: ") + improvement.Description + "\n"
		output += "\t\t" + utils.Yellow("Suggestion: ") + improvement.Suggestion + "\n"
	}
	output += "\t" + utils.Muted("[/AREAS FOR IMPROVEMENT]") + "\n"

	output += "\t" + utils.Muted("[QUESTIONS]") + "\n"
	for _, question := range rp.Questions {
		output += "\t\t" + utils.Blue("Question: ") + question.Query + "\n"
		output += "\t\t" + utils.Green("Context: ") + question.Context + "\n"
	}
	output += "\t" + utils.Muted("[/QUESTIONS]") + "\n"

	output += "\t" + utils.Muted("[RECOMMENDATIONS]") + "\n"
	for _, recommendation := range rp.Recommendations {
		output += "\t\t" + utils.Blue("Recommendation: ") + recommendation.Description + "\n"
		output += "\t\t" + utils.Green("Rationale: ") + recommendation.Rationale + "\n"
		output += "\t\t" + utils.Yellow("Priority: ") + recommendation.Priority + "\n"
	}
	output += "\t" + utils.Muted("[/RECOMMENDATIONS]") + "\n"

	output += "\t" + utils.Blue("Overall Rating: ") + FloatToString(rp.OverallRating) + "/5.0\n"
	output += "\t" + utils.Red("Done: ") + BoolToString(rp.Done) + "\n"
	output += utils.Dark("[/REVIEW PLAN]") + "\n"
	return output
}

type ReviewSection struct {
	Name     string  `json:"name" jsonschema:"description=The name of the section being reviewed"`
	Comments string  `json:"comments" jsonschema:"description=Comments about the section"`
	Rating   float64 `json:"rating" jsonschema:"description=Rating for this section (0.0 to 5.0)"`
}

type ReviewStrength struct {
	Description string `json:"description" jsonschema:"description=A description of the identified strength"`
	Impact      string `json:"impact" jsonschema:"description=The positive impact of this strength"`
}

type ReviewImprovement struct {
	Description string `json:"description" jsonschema:"description=A description of the area needing improvement"`
	Suggestion  string `json:"suggestion" jsonschema:"description=A suggestion for how to improve this area"`
}

type ReviewQuestion struct {
	Query   string `json:"query" jsonschema:"description=The question raised during the review"`
	Context string `json:"context" jsonschema:"description=The context or reason for asking this question"`
}

type ReviewRecommendation struct {
	Description string `json:"description" jsonschema:"description=A description of the recommendation"`
	Rationale   string `json:"rationale" jsonschema:"description=The rationale behind this recommendation"`
	Priority    string `json:"priority" jsonschema:"description=The priority level of this recommendation (e.g., High, Medium, Low)"`
}
