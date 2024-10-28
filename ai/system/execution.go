package system

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/provider"
)

// Execution manages dependencies, communication, and artifact tracking.
type Execution struct {
	process   *Process
	teams     map[string]*ai.Team
	artifacts map[string]string // Centralized artifact storage for cross-team dependency
}

// NewExecution creates a new Execution instance.
func NewExecution(process *Process, teams map[string]*ai.Team) *Execution {
	return &Execution{
		process:   process,
		teams:     teams,
		artifacts: make(map[string]string),
	}
}

// Execute runs the steps in dependency order.
func (e *Execution) Execute() <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		for _, team := range e.process.Teams {
			log.Info("Executing team plan", "team", team.Name)
			for _, step := range team.Plan.Steps {
				log.Info("Processing step", "step", step.ID, "team", team.Name)

				// Wait until all dependencies are met
				for !e.areDependenciesSatisfied(step.Dependencies) {
					time.Sleep(100 * time.Millisecond)
				}

				// Execute step with a custom prompt
				stepOut := e.teams[team.Name].Execute(step.Agent, step.ID, step.Prompt)

				var artifactContent string
				for event := range stepOut {
					out <- event
					artifactContent += event.Content
				}

				// Store output in artifacts
				e.artifacts[step.ID] = artifactContent
			}
		}
	}()

	return out
}

// areDependenciesSatisfied checks if dependencies (within and across teams) are satisfied.
func (e *Execution) areDependenciesSatisfied(dependencies []Dependency) bool {
	for _, dep := range dependencies {
		if _, exists := e.artifacts[dep.From]; !exists {
			return false
		}
	}
	return true
}

// dependenciesToStrings converts a slice of Dependency to a slice of strings.
func dependenciesToStrings(dependencies []Dependency) []string {
	out := []string{}

	for _, dep := range dependencies {
		out = append(out, fmt.Sprintf("%s -> %s", dep.From, dep.To))
	}

	return out
}
