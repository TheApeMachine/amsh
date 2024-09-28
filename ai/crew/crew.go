package crew

import (
	"context"
	"encoding/json"

	"github.com/theapemachine/amsh/errnie"
)

/*
Crew orchestrates the collaboration of multiple AI agents.
*/
type Crew struct {
	Director  Director
	Writer    Writer
	Flow      Flow
	Extractor Extractor
}

/*
NewCrew initializes a new Crew with predefined specialized agents.
*/
func NewCrew(ctx context.Context) *Crew {
	return &Crew{
		Director:  NewDirector(ctx),
		Writer:    NewWriter(ctx),
		Flow:      NewFlow(ctx),
		Extractor: NewExtractor(ctx),
	}
}

/*
Direction encapsulates the high-level guidance for the story's progression.
*/
type Direction struct {
	Change      string `json:"change"`
	Description string `json:"description"`
}

/*
Script represents a detailed scene description with associated actions.
*/
type Script struct {
	Scene   string   `json:"scene"`
	Actions []string `json:"actions"`
}

/*
FlowDecision represents the decision on how the story should progress.
*/
type FlowDecision struct {
	Flow   string `json:"flow"`
	Scope  string `json:"scope"`
	Repeat bool   `json:"repeat"`
}

/*
Skill represents a capability of an agent.
*/
type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       string `json:"level"`
}

/*
Memory encapsulates a single recollection or piece of information.
*/
type Memory struct {
	Timestamp string `json:"timestamp"`
	Scene     string `json:"scene"`
	Action    string `json:"action"`
	Content   string `json:"content"`
}

/*
Experience represents a significant event or period in an agent's history.
*/
type Experience struct {
	Title       string `json:"title"`
	Location    string `json:"location"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Description string `json:"description"`
}

/*
Relationship defines connections between an agent and other entities.
*/
type Relationship struct {
	Target      string        `json:"target"`
	Type        string        `json:"type"`
	Status      string        `json:"status"`
	Description string        `json:"description"`
	Experiences []*Experience `json:"experiences"`
	Memories    []*Memory     `json:"memories"`
}

/*
Profile encapsulates all aspects of an agent's persona.
*/
type Profile struct {
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Skills        []*Skill        `json:"skills"`
	Experiences   []*Experience   `json:"experiences"`
	Memories      []*Memory       `json:"memories"`
	Relationships []*Relationship `json:"relationships"`
}

/*
String converts the Profile to a JSON string.
*/
func (profile *Profile) String() string {
	jsonData, err := json.Marshal(profile)
	if err != nil {
		errnie.Error(err.Error())
		return ""
	}
	return string(jsonData)
}

/*
Unmarshal populates the Profile from a JSON string.
*/
func (profile *Profile) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), profile)
}

/*
Colors is a list of ANSI escape codes for colored output.
*/
var Colors = []string{
	"\033[32m", "\033[33m", "\033[34m", "\033[35m", "\033[36m",
	"\033[37m", "\033[31m", "\033[90m", "\033[91m", "\033[92m",
}

/*
Reset ANSI escape code.
*/
var reset = "\033[0m"
