package system

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

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
	agent     *ai.Agent
	team      *ai.Team
	engine    *reasoning.Engine
	learner   *learning.LearningAdapter
	planner   *planning.Planner
	toolset   *ai.Toolset
	mu        sync.RWMutex
}

/*
NewProcessManager sets up the process manager, and the Agent that will act as the sequencer.
*/
func NewProcessManager(arch *Architecture) *ProcessManager {
	v := viper.GetViper()
	toolset := ai.NewToolset()

	// Initialize the team and core components
	team := ai.NewTeam(toolset)
	kb := reasoning.NewKnowledgeBase()
	validator := reasoning.NewValidator(kb)
	metaReasoner := reasoning.NewMetaReasoner()

	// Initialize resources
	metaReasoner.InitializeResources(map[string]float64{
		"cpu":    1.0,
		"memory": 1.0,
		"time":   1.0,
	})

	// Create engine and learning system
	engine := reasoning.NewEngine(validator, metaReasoner)
	learner := learning.NewLearningAdapter()
	planner := planning.NewPlanner()

	return &ProcessManager{
		arch:      arch,
		processes: make(map[string]string),
		agent: ai.NewAgent(
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
		),
		team:    team,
		engine:  engine,
		learner: learner,
		planner: planner,
		toolset: toolset,
	}
}

/*
HandleProcess is the unified entry point for handling any process.
It handles the routing to appropriate teams and agents based on the process key.
*/
func (pm *ProcessManager) HandleProcess(ctx context.Context, userPrompt string) <-chan []byte {
	log.Info("Handling process", "userPrompt", userPrompt)

	// Create response channel
	responseChan := make(chan []byte)

	go func() {
		defer close(responseChan)

		// Create execution plan - store result in processes map for future use
		plan, err := pm.createExecutionPlan(ctx, userPrompt)
		if err != nil {
			errnie.Error(err)
			return
		}
		pm.mu.Lock()
		pm.processes[plan.Name] = userPrompt // Store the plan
		pm.mu.Unlock()

		// Get required agents
		researcher := pm.team.GetResearcher()
		analyst := pm.team.GetAnalyst()
		if researcher == nil || analyst == nil {
			errnie.Error(fmt.Errorf("failed to get required agents"))
			return
		}

		// Process reasoning chain
		chain, err := pm.engine.ProcessReasoning(ctx, userPrompt)
		if err != nil {
			errnie.Error(err)
			return
		}

		// Convert reasoning chain for analyst
		typesChain := pm.convertReasoningChain(chain)

		// Research phase
		if err := researcher.ReceiveMessage(userPrompt); err != nil {
			errnie.Error(err)
			return
		}

		findings, err := researcher.ExecuteTask()
		if err != nil {
			errnie.Error(err)
			return
		}

		// Analysis phase
		if err := pm.passChainToAnalyst(analyst, typesChain, userPrompt, findings); err != nil {
			errnie.Error(err)
			return
		}

		solution, err := analyst.ExecuteTask()
		if err != nil {
			errnie.Error(err)
			return
		}

		// Stream responses back - fix Event struct usage
		response := provider.Event{
			Type:    provider.EventToken,
			Content: solution,
		}

		pm.mu.RLock()
		responseChan <- pm.makeEvent(response)
		pm.mu.RUnlock()

		// Update learning system
		pm.learner.RecordStrategyExecution(nil, typesChain)
	}()

	return responseChan
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

func (pm *ProcessManager) createExecutionPlan(ctx context.Context, prompt string) (*planning.Plan, error) {
	planReq := planning.CreatePlanRequest{
		Name:        "Process Execution",
		Description: "Execute AI pipeline for user prompt",
		EndTime:     time.Now().Add(1 * time.Hour),
		Goals: []planning.CreateGoalRequest{
			{
				Name:        "Analysis",
				Description: "Analyze and process user prompt",
				Priority:    1,
				Deadline:    time.Now().Add(15 * time.Minute),
			},
		},
	}

	return pm.planner.CreatePlan(ctx, planReq)
}

// Update type definition to match the reasoning package
func (pm *ProcessManager) convertReasoningChain(chain *types.ReasoningChain) *types.ReasoningChain {
	if chain == nil {
		return nil
	}

	typesChain := &types.ReasoningChain{
		Steps: make([]types.ReasoningStep, len(chain.Steps)),
	}

	for i, step := range chain.Steps {
		typesChain.Steps[i] = types.ReasoningStep{
			Strategy:   step.Strategy,
			Confidence: step.Confidence,
		}
	}

	return typesChain
}

func (pm *ProcessManager) passChainToAnalyst(analyst *ai.Agent, chain *types.ReasoningChain, prompt, research string) error {
	analysis := fmt.Sprintf(`
Prompt: %s

Research Findings: %s

Reasoning Steps:
`, prompt, research)

	for i, step := range chain.Steps {
		analysis += fmt.Sprintf(`
Step %d:
- Strategy: %v
- Confidence: %.2f
`, i+1, step.Strategy, step.Confidence)
	}

	return analyst.ReceiveMessage(analysis)
}
