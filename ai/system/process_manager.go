package system

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
)

/*
ProcessManager handles the lifecycle of a workload, mapping it across teams, agents, and processes.
It is ultimately controlled by an Agent called the Sequencer, which has been prompted to orchestrate
all the moving parts needed to make the system work.
*/
type ProcessManager struct {
	key          string
	toolset      *ai.Toolset
	orchestrator *ai.Agent
	extractor    *ai.Agent
}

/*
NewProcessManager sets up the process manager, and the Agent that will act as the sequencer.
*/
func NewProcessManager(key string) *ProcessManager {
	log.Info("NewProcessManager", "key", key)
	planning := process.NewPlanning()
	toolset := ai.NewToolset()

	return &ProcessManager{
		key:     key,
		toolset: toolset,
		orchestrator: ai.NewAgent(
			fmt.Sprintf("%s-orchestrator", key),
			"orchestrator",
			strings.ReplaceAll(
				viper.GetString(fmt.Sprintf("ai.setups.%s.orchestration.prompt", key)),
				"{{schemas}}",
				planning.GenerateSchema(),
			),
		),
		extractor: ai.NewAgent(
			fmt.Sprintf("%s-extractor", key),
			"extractor",
			strings.ReplaceAll(
				viper.GetString(fmt.Sprintf("ai.setups.%s.extraction.prompt", key)),
				"{{schemas}}",
				toolset.Schemas(),
			),
		),
	}
}

/*
Execute the process manager, using the incoming message as the initial prompt for the process.
*/
func (pm *ProcessManager) Execute(incoming string) <-chan provider.Event {
	log.Info("Execute", "incoming", incoming)
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		var accumulator string

		for event := range pm.orchestrator.Execute(incoming) {
			accumulator += event.Content
			out <- event
		}

		planning := process.NewPlanning().Extract(accumulator)

		if planning == nil {
			errnie.Error(errors.New("failed to extract planning"))
			return
		}

		teams := map[string]*ai.Team{}

		for _, teamConfig := range planning.Teams {
			agents := map[string]*ai.Agent{}

			for _, agentConfig := range teamConfig.Agents {
				agents[agentConfig.RoleKey] = ai.NewAgent(
					fmt.Sprintf("%s-%s", teamConfig.TeamKey, agentConfig.RoleKey),
					agentConfig.RoleKey,
					agentConfig.SystemPrompt,
				)
			}

			teams[teamConfig.TeamKey] = ai.NewTeam(pm.key, teamConfig.TeamKey, agents)
		}

		for _, goal := range planning.Goals {
			for _, step := range goal.Steps {
				var teamAccumulator string

				for event := range teams[step.TeamKey].Execute(step) {
					teamAccumulator += event.Content
					out <- event
				}

				var extractionAccumulator string

				for event := range pm.extractor.Execute(teamAccumulator) {
					extractionAccumulator += event.Content
					out <- event
				}

				toolResult := pm.detectToolCalls(extractionAccumulator)

				out <- provider.Event{
					Type:    provider.EventToolCall,
					Content: toolResult,
				}
			}
		}

	}()

	return out
}

func (processManager ProcessManager) detectToolCalls(content string) string {
	log.Info("detectToolCalls")
	// Extract all markdown JSON blocks
	re := regexp.MustCompile("(?s)json\\s*(\\{.*?\\})\\s*")
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		// Parse the JSON content into our map
		toolCall := map[string]any{}
		jsonContent := strings.TrimSpace(match[1]) // Trim any extraneous whitespace

		if err := json.Unmarshal([]byte(jsonContent), &toolCall); err != nil {
			errnie.Error(err)
			spew.Dump(jsonContent) // Dump the specific JSON content causing the error
			return "Something went wrong"
		}

		// Iterate through the tool calls
		for toolName, arguments := range toolCall {
			log.Info("toolName", toolName)
			// Use the tool if it exists
			if result := processManager.toolset.Use(toolName, arguments.(map[string]any)); result != "" {
				return result
			}
		}
	}

	return "Something went wrong"
}
