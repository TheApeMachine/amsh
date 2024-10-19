package format

import (
	"encoding/json"
	"fmt"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Researching struct {
	Topic            string          `json:"topic" jsonschema:"description=The main topic being researched"`
	KeyFindings      []Finding       `json:"key_findings" jsonschema:"description=Important discoveries and insights"`
	Connections      []Connection    `json:"connections" jsonschema:"description=Subtle connections between concepts"`
	DetailedAnalysis []AnalysisPoint `json:"detailed_analysis" jsonschema:"description=In-depth analysis of specific aspects"`
	BigPicture       string          `json:"big_picture" jsonschema:"description=Overall context and implications"`
	NextSteps        []ResearchStep  `json:"next_steps" jsonschema:"description=Proposed actions for further research"`
	Done             bool            `json:"done" jsonschema:"description=Indicates if the research plan is complete;required=true"`
}

func NewResearching() *Researching {
	return &Researching{}
}

func (rp *Researching) Print(data []byte) (isDone bool, err error) {
	if err := errnie.Error(json.Unmarshal(data, rp)); err != nil {
		return false, err
	}

	fmt.Println(rp.String())
	return rp.Done, nil
}

func (rp Researching) String() string {
	output := utils.Dark("[RESEARCH PLAN]") + "\n"

	output += "\t" + utils.Blue("Topic: ") + rp.Topic + "\n\n"

	output += "\t" + utils.Muted("[KEY FINDINGS]") + "\n"
	for _, finding := range rp.KeyFindings {
		output += "\t\t" + utils.Green("Finding: ") + finding.Description + "\n"
		output += "\t\t" + utils.Yellow("Importance: ") + finding.Importance + "\n"
	}
	output += "\t" + utils.Muted("[/KEY FINDINGS]") + "\n"

	output += "\t" + utils.Muted("[CONNECTIONS]") + "\n"
	for _, connection := range rp.Connections {
		output += "\t\t" + utils.Red("Concepts: ") + connection.ConceptA + " <-> " + connection.ConceptB + "\n"
		output += "\t\t" + utils.Green("Relationship: ") + connection.Relationship + "\n"
	}
	output += "\t" + utils.Muted("[/CONNECTIONS]") + "\n"

	output += "\t" + utils.Muted("[DETAILED ANALYSIS]") + "\n"
	for _, point := range rp.DetailedAnalysis {
		output += "\t\t" + utils.Blue("Aspect: ") + point.Aspect + "\n"
		output += "\t\t" + utils.Green("Analysis: ") + point.Analysis + "\n"
	}
	output += "\t" + utils.Muted("[/DETAILED ANALYSIS]") + "\n"

	output += "\t" + utils.Blue("Big Picture: ") + rp.BigPicture + "\n"

	output += "\t" + utils.Muted("[NEXT STEPS]") + "\n"
	for _, step := range rp.NextSteps {
		output += "\t\t" + utils.Blue("Action: ") + step.Action + "\n"
		output += "\t\t" + utils.Green("Rationale: ") + step.Rationale + "\n"
	}
	output += "\t" + utils.Muted("[/NEXT STEPS]") + "\n"

	output += "\t" + utils.Red("Done: ") + BoolToString(rp.Done) + "\n"
	output += utils.Dark("[/RESEARCH PLAN]") + "\n"
	return output
}

type Finding struct {
	Description string `json:"description" jsonschema:"description=A description of the key finding"`
	Importance  string `json:"importance" jsonschema:"description=The significance of this finding"`
}

type Connection struct {
	ConceptA     string `json:"concept_a" jsonschema:"description=The first concept in the connection"`
	ConceptB     string `json:"concept_b" jsonschema:"description=The second concept in the connection"`
	Relationship string `json:"relationship" jsonschema:"description=The nature of the relationship between the concepts"`
}

type AnalysisPoint struct {
	Aspect   string `json:"aspect" jsonschema:"description=The specific aspect being analyzed"`
	Analysis string `json:"analysis" jsonschema:"description=Detailed analysis of the aspect"`
}

type ResearchStep struct {
	Action    string `json:"action" jsonschema:"description=The proposed action for further research"`
	Rationale string `json:"rationale" jsonschema:"description=The reasoning behind the proposed action"`
}
