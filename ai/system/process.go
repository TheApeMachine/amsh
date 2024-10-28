package system

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

type Step struct {
	ID           string       `json:"id"`
	Agent        string       `json:"agent"`
	Prompt       string       `json:"prompt"`
	Inputs       []string     `json:"inputs"`
	Dependencies []Dependency `json:"dependencies"`
	Substeps     []Step       `json:"substeps"`
}

type Plan struct {
	Goal  string `json:"goal"`
	Steps []Step `json:"steps"`
}

type Dependency struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

type Process struct {
	Category string `json:"category"`
	MainGoal string `json:"main_goal"`
	Teams    []struct {
		Name     string `json:"name"`
		Teamlead string `json:"teamlead"`
		Agents   []struct {
			Name         string   `json:"name"`
			SystemPrompt string   `json:"system_prompt"`
			Tools        []string `json:"tools"`
		} `json:"agents"`
		Plan Plan `json:"plan"`
	} `json:"teams"`
	CrossTeamDependencies []Dependency `json:"cross_team_dependencies"`
	FinalSynthesis        struct {
		Goal                  string `json:"goal"`
		StepsToCombineResults []Step `json:"steps_to_combine_results"`
	} `json:"final_synthesis"`
	Metadata struct {
		Priority  string    `json:"priority"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"metadata"`
}

func NewProcess() *Process {
	return &Process{}
}

func (process *Process) Unmarshal(prompt string) error {
	log.Info("unmarshalling process")

	// Extract the JSON from the prompt
	jsonStart := strings.Index(prompt, "{")
	jsonEnd := strings.LastIndex(prompt, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		log.Error("invalid prompt", "error", errors.New("invalid prompt"))
		return errors.New("invalid prompt")
	}

	// Unmarshal the JSON
	return json.Unmarshal([]byte(prompt[jsonStart:jsonEnd+1]), process)
}
