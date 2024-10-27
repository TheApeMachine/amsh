package system

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai"          // For Agent, Team, Toolset
	"github.com/theapemachine/amsh/ai/provider" // For provider package

	// For Role and other types
	"github.com/theapemachine/amsh/datalake"
	"github.com/theapemachine/amsh/utils"
)

/*
Architecture represents a configurable multi-agent system that coordinates processes
and integrations.
*/
type Architecture struct {
	Name           string              `json:"name"`
	Config         *ArchitectureConfig `json:"config"`
	Teams          map[string]*ai.Team `json:"teams"`
	Orchestrator   *ai.Agent           `json:"orchestrator"`
	LastUpdateTime time.Time           `json:"last_update_time"`

	ProcessManager  *ProcessManager  `json:"-"`
	agentManager    *ai.AgentManager `json:"-"`
	toolset         *ai.Toolset      `json:"-"`
	monitor         *Monitor         `json:"-"`
	workloadManager *WorkloadManager `json:"-"`
	storage         *datalake.Conn   `json:"-"`
	saveTicker      *time.Ticker     `json:"-"`
	mu              sync.RWMutex     `json:"-"`
}

/*
ArchitectureConfig holds the configuration for a specific architecture.
This matches the structure in the YAML config.
*/
type ArchitectureConfig struct {
	Name          string              `yaml:"name"`
	System        string              `yaml:"system"`
	Orchestration string              `yaml:"orchestration"`
	Processes     map[string]string   `yaml:"processes"`
	teams         map[string][]string `yaml:"teams"`
	Agents        map[string]string   `yaml:"agents"`
}

/*
NewArchitecture creates a new instance of the specified architecture.
*/
func NewArchitecture(ctx context.Context, name string) (*Architecture, error) {
	// Load architecture-specific configuration
	config, err := loadArchitectureConfig(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration for architecture %s: %w", name, err)
	}

	// Initialize storage connection
	storage := datalake.NewConn("architectures")

	// Initialize toolset first as it's needed by agents
	toolset := ai.NewToolset()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize toolset: %w", err)
	}

	// Initialize the agent manager
	agentManager := ai.GetAgentManager()

	// Create the base system
	arch := &Architecture{
		Name:         name,
		Config:       config,
		agentManager: agentManager,
		toolset:      toolset,
		monitor:      NewMonitor(),
		Teams:        make(map[string]*ai.Team),
		storage:      storage,
		saveTicker:   time.NewTicker(5 * time.Minute), // Save state every 5 minutes
	}

	processManager := NewProcessManager(arch)
	arch.ProcessManager = processManager

	// Try to restore previous state
	if err := arch.restoreState(); err != nil {
		// Log error but continue with fresh state
		fmt.Printf("Warning: Could not restore previous state: %v\n", err)
	}

	// Initialize workload manager
	workloadManager, err := NewWorkloadManager(ctx, arch)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize workload manager: %w", err)
	}
	arch.workloadManager = workloadManager

	// Initialize teams if not restored
	if len(arch.Teams) == 0 {
		if err := arch.initializeTeams(); err != nil {
			return nil, fmt.Errorf("failed to initialize teams: %w", err)
		}
	}

	// Start state persistence goroutine
	go arch.persistStateRoutine(ctx)

	return arch, nil
}

/*
persistStateRoutine periodically saves the architecture state to storage.
*/
func (arch *Architecture) persistStateRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-arch.saveTicker.C:
			if err := arch.saveState(); err != nil {
				fmt.Printf("Error saving architecture state: %v\n", err)
			}
		}
	}
}

/*
saveState persists the current state of the architecture to storage.
*/
func (arch *Architecture) saveState() error {
	arch.mu.Lock()
	defer arch.mu.Unlock()

	arch.LastUpdateTime = time.Now()

	// Open connection to state file
	key := fmt.Sprintf("architectures/%s/state.json", arch.Name)
	conn := datalake.NewConn(key)
	defer conn.Close()

	// Create encoder and encode directly to the connection
	if err := json.NewEncoder(conn).Encode(arch); err != nil {
		return fmt.Errorf("failed to encode architecture state: %w", err)
	}

	return nil
}

/*
restoreState attempts to restore the architecture state from storage.
*/
func (arch *Architecture) restoreState() error {
	// Open connection to state file
	key := fmt.Sprintf("architectures/%s/state.json", arch.Name)
	conn := datalake.NewConn(key)
	defer conn.Close()

	// Create decoder and decode directly from the connection
	if err := json.NewDecoder(conn).Decode(arch); err != nil {
		return fmt.Errorf("failed to decode architecture state: %w", err)
	}

	// Reconnect non-serialized components
	arch.reestablishConnections()

	return nil
}

/*
reestablishConnections reconnects the non-serialized components after state restoration.
*/
func (arch *Architecture) reestablishConnections() {
}

/*
ArchitectureConfig holds the configuration for a specific architecture.
This matches the structure in the YAML config.
*/
func loadArchitectureConfig(name string) (*ArchitectureConfig, error) {
	var config ArchitectureConfig
	configKey := fmt.Sprintf("ai.setups.%s", name)

	// Extract configuration from viper
	if err := viper.UnmarshalKey(configKey, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	config.Name = name
	return &config, nil
}

/*
initializeTeams creates and configures all teams defined in the architecture's configuration.
*/
func (arch *Architecture) initializeTeams() error {
	for teamName, members := range arch.Config.teams {
		team := ai.NewTeam(arch.toolset)

		for _, memberName := range members {
			v := viper.GetViper()

			agent := ai.NewAgent(
				utils.NewName(),
				"sequencer",
				v.GetString("ai.setups.marvin.system"),
				v.GetString("ai.setups.marvin.agents.sequencer.role"),
				ai.NewToolset().GetToolsForRole(memberName),
				provider.NewRandomProvider(map[string]string{
					"openai":    os.Getenv("OPENAI_API_KEY"),
					"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
					"google":    os.Getenv("GOOGLE_API_KEY"),
					"cohere":    os.Getenv("COHERE_API_KEY"),
				}),
			)

			arch.agentManager.RegisterAgent(agent)
			team.AddMember(agent)
		}

		arch.Teams[teamName] = team
	}

	return nil
}

// Add this method to the Architecture type
func (arch *Architecture) GetTeam(name string) *ai.Team {
	arch.mu.RLock()
	defer arch.mu.RUnlock()

	team, exists := arch.Teams[name]
	if !exists {
		return nil
	}
	return team
}
