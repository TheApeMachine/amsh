package system

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/provider"
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
	mu        sync.RWMutex
}

/*
NewProcessManager sets up the process manager, and the Agent that will act as the sequencer.
*/
func NewProcessManager(arch *Architecture) *ProcessManager {
	v := viper.GetViper()

	return &ProcessManager{
		arch:      arch,
		processes: make(map[string]string),
		agent: ai.NewAgent(
			utils.NewName(),
			"sequencer",
			v.GetString("ai.setups.marvin.system"),
			v.GetString("ai.setups.marvin.agents.sequencer.role"),
			ai.NewToolset().GetToolsForRole("sequencer"),
			provider.NewRandomProvider(map[string]string{
				"openai":    os.Getenv("OPENAI_API_KEY"),
				"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
				"google":    os.Getenv("GOOGLE_API_KEY"),
				"cohere":    os.Getenv("COHERE_API_KEY"),
			}),
		),
	}
}

/*
HandleProcess is the unified entry point for handling any process.
It handles the routing to appropriate teams and agents based on the process key.
*/
func (pm *ProcessManager) HandleProcess(ctx context.Context, userPrompt string) <-chan []byte {
	pm.agent.Task = userPrompt

	for evt := range pm.agent.ExecuteTaskStream() {
		switch pm.agent.GetState() {
		case types.StateDone:
			return
		}
	}

	// Create a channel to stream responses
	responseChan := make(chan []byte)

	// Start a goroutine to handle the process and stream responses
	go func() {
		defer close(responseChan)

		// Send to teamlead for processing
		teamlead.ReceiveMessage(processMsg)

		// Stream responses from the teamlead
		for response := range teamlead.ExecuteTaskStream() {
			responseChan <- pm.makeEvent(response)
		}
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

// RegisterProcess registers a new process with the manager
func (pm *ProcessManager) RegisterProcess(name, description string, teams []string, handler func(ctx context.Context, input interface{}) (interface{}, error)) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if process already exists
	if _, exists := pm.processes[name]; exists {
		return fmt.Errorf("process %s already registered", name)
	}

	// Register the new process
	pm.processes[name] = &Process{
		Name:        name,
		Description: description,
		Teams:       teams,
		Handler:     handler,
	}

	return nil
}

// StartProcess executes a registered process with the given input
func (pm *ProcessManager) StartProcess(ctx context.Context, name string, input interface{}) (interface{}, error) {
	pm.mu.RLock()
	process, exists := pm.processes[name]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("process %s not found", name)
	}

	// Execute the process handler
	return process.Handler(ctx, input)
}
