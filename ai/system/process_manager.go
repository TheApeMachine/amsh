package system

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/learning"
	"github.com/theapemachine/amsh/ai/planning"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/reasoning"
	"github.com/theapemachine/amsh/ai/types"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

/*
ProcessManager handles the lifecycle of a workload, mapping it across teams, agents, and processes.
It is ultimately controlled by an Agent called the Sequencer, which has been prompted to orchestrate
all the moving parts needed to make the system work.
*/
type ProcessManager struct {
	arch      *Architecture
	processes map[string]string
	sequencer *ai.Agent
	reasoner  *ai.Agent
	analyst   *ai.Agent
	team      *ai.Team
	engine    *reasoning.Engine
	learner   *learning.LearningAdapter
	planner   *planning.Planner
	toolset   *ai.Toolset
	recruiter *ai.Agent
}

/*
NewProcessManager sets up the process manager, and the Agent that will act as the sequencer.
*/
func NewProcessManager(arch *Architecture) *ProcessManager {
	toolset := ai.NewToolset()
	team := ai.NewTeam(toolset)
	kb := reasoning.NewKnowledgeBase()
	validator := reasoning.NewValidator(kb)
	metaReasoner := reasoning.NewMetaReasoner()

	// Initialize meta reasoner with default strategies
	metaReasoner.InitializeResources(map[string]float64{
		"cpu":    1.0,
		"memory": 1.0,
		"time":   1.0,
	})

	// Add default strategies
	// Add default strategies
	metaReasoner.AddDefaultStrategies([]types.MetaStrategy{
		{
			Name:     "general_analysis",
			Priority: 1,
			Constraints: []string{
				"complexity_low",
				"certainty_medium",
			},
			Resources: map[string]float64{
				"cpu":    0.2,
				"memory": 0.2,
				"time":   0.3,
			},
		},
		{
			Name:     "fact_verification",
			Priority: 2,
			Constraints: []string{
				"complexity_low",
				"certainty_high",
			},
			Resources: map[string]float64{
				"cpu":    0.3,
				"memory": 0.3,
				"time":   0.4,
			},
		},
	})

	// Create engine and learning system
	engine := reasoning.NewEngine(validator, metaReasoner)
	learner := learning.NewLearningAdapter()
	planner := planning.NewPlanner()

	agents := makeAgents("recruiter", "analyst", "reasoner", "sequencer")
	return &ProcessManager{
		arch:      arch,
		processes: make(map[string]string),
		sequencer: agents["sequencer"],
		recruiter: agents["recruiter"],
		analyst:   agents["analyst"],
		reasoner:  agents["reasoner"],
		team:      team,
		engine:    engine,
		learner:   learner,
		planner:   planner,
		toolset:   toolset,
	}
}

func makeAgents(roles ...string) map[string]*ai.Agent {
	v := viper.GetViper()
	toolset := ai.NewToolset()
	agents := make(map[string]*ai.Agent)

	for _, role := range roles {
		agents[role] = ai.NewAgent(
			utils.NewName(),
			"sequencer",
			v.GetString("ai.setups.marvin.system"),
			v.GetString("ai.setups.marvin.agents.sequencer.role"),
			toolset.GetToolsForRole("sequencer"),
			provider.NewRandomProvider(map[string]string{
				"openai":    os.Getenv("OPENAI_API_KEY"),
				"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
				"google":    os.Getenv("GOOGLE_API_KEY"),
				"cohere":    os.Getenv("COHERE_API_KEY"),
			}),
		)
	}

	return agents
}

func (pm *ProcessManager) createTeamForProcess(process *ai.Process) *ai.Team {
	log.Info("Creating team for process", "process", process.Name)
	team := ai.NewTeam(pm.toolset)
	if team == nil {
		return nil // Early return if team creation fails
	}

	// Get unique agent roles from all steps
	agentRoles := make(map[string]bool)
	for _, step := range process.Steps {
		for _, agentRole := range step.Agents {
			agentRoles[agentRole] = true
		}
	}

	// Create or get agents for each role
	for role := range agentRoles {
		// Get specialization from config
		specialization := viper.GetString(fmt.Sprintf("ai.setups.marvin.agents.%s.specialization", role))

		// Use getOrCreateAgent instead of creating new agent directly
		agent := pm.getOrCreateAgent(role, specialization)
		team.AddMember(agent)
	}

	return team
}

func (pm *ProcessManager) recordProcessOutcome(process *ai.Process, plan *planning.Plan) {
	// Record process completion status
	success := process.Status == ai.ProcessStatusComplete

	// Calculate overall confidence from steps and plan goals
	var totalConfidence float64
	goalsCompleted := 0

	// Check both process steps and plan goals
	for _, step := range process.Steps {
		if step.Status == ai.StepStatusCompleted {
			totalConfidence += 1.0
		}
	}

	for _, goal := range plan.Goals {
		if goal.Status == "completed" {
			goalsCompleted++
		}
	}

	// Average confidence across both steps and goals
	totalSteps := len(process.Steps) + len(plan.Goals)
	confidence := (totalConfidence + float64(goalsCompleted)) / float64(totalSteps)

	// Create reasoning chain for learning
	chain := &types.ReasoningChain{
		Steps:      make([]types.ReasoningStep, len(process.Steps)),
		Validated:  success,
		Confidence: confidence,
	}

	// Record the outcome in the learning system
	pm.learner.RecordStrategyExecution(&types.MetaStrategy{
		Name:     process.Name,
		Priority: 1, // Default priority since Plan doesn't have a Priority field
	}, chain)
}

/*
HandleProcess is the unified entry point for handling any process.
It handles the routing to appropriate teams and agents based on the process key.
High-level, the process should follow this flow:

1. Analyze the user prompt, and the origin from which it was received. This will determine a lot of the overall strategy.
2. Formulate a plan that includes not only the goals and steps, but also the type of reasoning that will be needed for each step.
3. Determine whether an existing process can be used, or whether a new, customized process needs to be created dynamically.
4. Determine the team composition needed to execute the process.
5. Have the recruiter create a team for the process, and for each needed role checking if there is an existing configuration, or a new custom role needs to be created dynamically.
6. Prompt the teamlead for the team so it is aware of the process and the team composition, and understand how to achieve and recognize a final response with their team.
7. Execute the process with the team.
8. Record the outcome of the process for learning.
9. Stream the output back to the origin from which the process was requested.

IMPORTANT: A lot of these steps should be determined not by the config or static code, but by correctly prompting and calling the LLM provider!
*/
func (pm *ProcessManager) HandleProcess(ctx context.Context, userPrompt string) <-chan []byte {
	log.Info("Starting process handling", "prompt", userPrompt)
	responseChan := make(chan []byte)

	go func() {
		defer close(responseChan)

		prompt := pm.extractPromptText(userPrompt)
		log.Info("Extracted prompt", "text", prompt)

		// First, let the recruiter determine needed roles
		pm.recruiter.Update(fmt.Sprintf(`Analyze this task and determine required agent roles:
			Task: %s
			
			Respond with roles in this format:
			ROLE: [role_name]
			REASON: [why this role is needed]`, prompt))

		roles, err := pm.recruiter.ExecuteTask()
		if err != nil {
			log.Error("Recruiter failed", "error", err)
			responseChan <- pm.makeEvent(provider.Event{
				Error: fmt.Errorf("recruiter failed: %w", err),
			})
			return
		}
		log.Info("Recruiter determined roles", "roles", roles)

		// Create team with the determined roles
		process := &ai.Process{
			Name:   prompt,
			Steps:  pm.createStepsFromRoles(roles),
			Status: ai.ProcessStatusRunning,
		}

		team := pm.createTeamForProcess(process)
		if team == nil {
			log.Error("Failed to create team")
			responseChan <- pm.makeEvent(provider.Event{
				Error: fmt.Errorf("team creation failed"),
			})
			return
		}
		log.Info("Team created", "members", team.Members())

		// Get execution plan from sequencer
		pm.sequencer.Update(fmt.Sprintf(`Plan execution for:
			Task: %s
			Team: %s
			
			Provide steps in this format:
			STEP: [step_name]
			AGENTS: [agent_roles]
			INPUT: [what agents need]`, prompt, roles))

		executionPlan, err := pm.sequencer.ExecuteTask()
		if err != nil {
			log.Error("Sequencer failed", "error", err)
			responseChan <- pm.makeEvent(provider.Event{
				Error: fmt.Errorf("sequencer failed: %w", err),
			})
			return
		}
		log.Info("Execution plan created")

		// Update process with execution plan steps
		process.Steps = pm.createStepsFromPlan(executionPlan)

		// Execute the process with the team
		log.Info("Starting process execution")
		outputChan := team.ExecuteProcess(ctx, process)

		// Stream only the final results
		var finalResponse string
		for output := range outputChan {
			finalResponse = string(output)
		}
		log.Info("Process completed")

		// Send only the final response
		responseChan <- []byte(finalResponse)

		// Record the process outcome
		pm.recordProcessOutcome(process, &planning.Plan{
			Name:  prompt,
			Goals: pm.convertStepsToGoals(process.Steps),
		})
	}()

	return responseChan
}

// extractPromptText handles different input formats and returns the actual prompt text
func (pm *ProcessManager) extractPromptText(input string) string {
	log.Info("Extracting prompt text")
	// First try to parse as JSON
	var jsonMsg struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(input), &jsonMsg); err == nil && jsonMsg.Text != "" {
		return jsonMsg.Text
	}

	// If not JSON or no text field, return the input as-is
	return input
}

// getOrCreateAgent either retrieves an existing agent or creates a new one
func (pm *ProcessManager) getOrCreateAgent(role, specialization string) *ai.Agent {
	log.Info("Getting or creating agent", "role", role, "specialization", specialization)
	// First try to get existing agent
	if agent := pm.team.GetAgent(role); agent != nil {
		return agent
	}

	// Create new agent with specialization
	systemPrompt := fmt.Sprintf("You are an AI agent specialized in %s. %s", role, specialization)
	userPrompt := fmt.Sprintf("Use your %s expertise to help process user requests", role)

	return ai.NewAgent(
		utils.NewName(),
		types.Role(role),
		systemPrompt,
		userPrompt,
		pm.toolset.GetToolsForRole(role),
		provider.NewRandomProvider(map[string]string{
			"openai":    os.Getenv("OPENAI_API_KEY"),
			"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
			"google":    os.Getenv("GOOGLE_API_KEY"),
			"cohere":    os.Getenv("COHERE_API_KEY"),
		}),
	)
}

func (pm *ProcessManager) makeEvent(response provider.Event) []byte {
	var (
		buf []byte
		err error
	)

	if buf, err = json.Marshal(response); err != nil {
		errnie.Error(err)
		return nil
	}

	return buf
}

// parseRolesFromRecruiter parses the recruiter's response to extract needed roles
func (pm *ProcessManager) parseRolesFromRecruiter(response string) []string {
	log.Info("Parsing roles from recruiter", "response", response)
	// The recruiter should return a structured response that we can parse
	var roles []string

	// Split response into lines and look for role definitions
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines that define roles
		if strings.HasPrefix(line, "ROLE:") {
			role := strings.TrimSpace(strings.TrimPrefix(line, "ROLE:"))
			roles = append(roles, role)
		}
	}

	// If no roles were found, include at least an analyst
	if len(roles) == 0 {
		roles = append(roles, "analyst")
	}

	// Log parsed roles
	log.Info("Parsed roles from recruiter", "roles", roles)

	// Remove the call to the undefined method
	// roles = pm.learner.RefineRolesBasedOnExperience(roles)

	return roles
}

// createStepsFromRoles converts the LLM-provided role list into ProcessSteps
func (pm *ProcessManager) createStepsFromRoles(rolesResponse string) []ai.ProcessStep {
	log.Info("Creating steps from roles", "roles", rolesResponse)

	// Parse the roles from the LLM response
	roles := pm.parseRolesFromRecruiter(rolesResponse)
	steps := make([]ai.ProcessStep, 0)

	// Create a step for each role
	for _, role := range roles {
		step := ai.ProcessStep{
			Name:   fmt.Sprintf("%s_analysis", role),
			Agents: []string{role},
			Status: ai.StepStatusPending,
		}
		steps = append(steps, step)
	}

	return steps
}

// createStepsFromPlan converts the LLM-provided execution plan into ProcessSteps
func (pm *ProcessManager) createStepsFromPlan(executionPlan string) []ai.ProcessStep {
	log.Info("Creating steps from plan", "plan", executionPlan)

	// Have the sequencer parse its own plan into structured steps
	pm.sequencer.Update(fmt.Sprintf(`Parse your execution plan into discrete steps:
        
        Original Plan:
        %s
        
        For each step, provide:
        1. Step name
        2. Required agents
        3. Dependencies on other steps
        `, executionPlan))

	planDetails, err := pm.sequencer.ExecuteTask()
	if err != nil {
		log.Error("Failed to parse execution plan", "error", err)
		// Return a single fallback step if parsing fails
		return []ai.ProcessStep{{
			Name:   "execute_task",
			Agents: []string{"analyst"}, // Fallback to analyst
			Status: ai.StepStatusPending,
		}}
	}

	// Have the analyst help structure the plan details
	pm.analyst.Update(fmt.Sprintf(`Convert this plan into structured steps:
        %s
        
        Extract each step's:
        - Name
        - Required agents
        - Input requirements
        `, planDetails))

	structuredPlan, err := pm.analyst.ExecuteTask()
	if err != nil {
		log.Error("Failed to structure plan", "error", err)
		return []ai.ProcessStep{{
			Name:   "execute_task",
			Agents: []string{"analyst"},
			Status: ai.StepStatusPending,
		}}
	}

	// Parse the structured plan into steps
	var steps []ai.ProcessStep

	// Split the response into lines and look for step definitions
	lines := strings.Split(structuredPlan, "\n")
	currentStep := ai.ProcessStep{}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "Step:"):
			// If we were building a step, append it
			if currentStep.Name != "" {
				steps = append(steps, currentStep)
			}
			currentStep = ai.ProcessStep{
				Name:   strings.TrimSpace(strings.TrimPrefix(line, "Step:")),
				Status: ai.StepStatusPending,
			}

		case strings.HasPrefix(line, "Agents:"):
			agents := strings.TrimSpace(strings.TrimPrefix(line, "Agents:"))
			currentStep.Agents = strings.Split(agents, ",")

		case strings.HasPrefix(line, "Input:"):
			currentStep.Input = strings.TrimSpace(strings.TrimPrefix(line, "Input:"))
		}
	}

	// Append the last step if it exists
	if currentStep.Name != "" {
		steps = append(steps, currentStep)
	}

	// If no steps were created, return a fallback step
	if len(steps) == 0 {
		steps = append(steps, ai.ProcessStep{
			Name:   "execute_task",
			Agents: []string{"analyst"},
			Status: ai.StepStatusPending,
		})
	}

	return steps
}

// Convert ProcessSteps to planning.Goals
func (pm *ProcessManager) convertStepsToGoals(steps []ai.ProcessStep) []planning.Goal {
	log.Info("Converting steps to goals", "steps", steps)
	goals := make([]planning.Goal, len(steps))
	for i, step := range steps {
		goals[i] = planning.Goal{
			Name:        step.Name,
			Description: step.Input,
			Status:      pm.convertStepStatusToGoalStatus(step.Status),
		}
	}
	return goals
}

// Helper to convert between status types
func (pm *ProcessManager) convertStepStatusToGoalStatus(status ai.StepStatus) planning.GoalStatus {
	log.Info("Converting step status to goal status", "status", status)
	switch status {
	case ai.StepStatusPending:
		return planning.GoalStatusPending
	case ai.StepStatusRunning:
		return planning.GoalStatusActive
	case ai.StepStatusCompleted:
		return planning.GoalStatusComplete
	case ai.StepStatusFailed:
		return planning.GoalStatusBlocked
	default:
		return planning.GoalStatusPending
	}
}
