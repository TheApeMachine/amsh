package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/goombaio/namegenerator"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/crew"
	"github.com/theapemachine/amsh/errnie"
)

/*
PipelineState encapsulates the current state of the pipeline.
*/
type PipelineState struct {
	Agents  []*AgentState
	History []string
	Scenes  []Scene
	Scene   int
	Action  int
	Agent   int
}

/*
Pipeline orchestrates the execution of scenes and agents.
*/
type Pipeline struct {
	ctx      context.Context
	conn     *Conn
	out      chan string
	crew     *crew.Crew
	agents   []*Agent
	history  []string
	scenes   []Scene
	scene    int
	action   int
	agent    int
	loglevel string
	err      error
	mutex    sync.Mutex
}

/*
Scene represents a distinct part of the story with its own context.
*/
type Scene struct {
	System   string
	Contexts []string
}

/*
NewPipeline initializes the pipeline with agents and scenes.
*/
func NewPipeline(ctx context.Context, conn *Conn) *Pipeline {
	errnie.Debug("NewPipeline")
	return &Pipeline{
		ctx:      ctx,
		conn:     NewConn(),
		out:      make(chan string),
		crew:     crew.NewCrew(ctx),
		agents:   make([]*Agent, 0),
		history:  make([]string, 0),
		loglevel: viper.GetString("loglevel"),
	}
}

/*
Generate runs the pipeline over all scenes and actions.
*/
func (pipeline *Pipeline) Generate() <-chan string {
	pipeline.setScenes()

	if pipeline.loglevel == "debug" {
		spew.Dump(pipeline)
	}

	go func() {
		defer close(pipeline.out)

		for {
			if pipeline.shouldEnd() {
				break
			}

			pipeline.executeCurrentAction()
		}
	}()

	return pipeline.out
}

/*
executeCurrentAction handles the execution of the current action.
*/
func (pipeline *Pipeline) executeCurrentAction() {
	pipeline.mutex.Lock()
	defer pipeline.mutex.Unlock()

	scene := pipeline.scenes[pipeline.scene]
	actionPrompt := scene.Contexts[pipeline.action]
	agent := pipeline.agents[pipeline.agent]

	// Build the context for the agent
	context := fmt.Sprintf("%s\n\n%s", pipeline.handleHistory(), actionPrompt)

	// Generate the agent's response
	pipeline.out <- agent.Color + "\n"

	responseChan := agent.Generate(pipeline.ctx, scene.System, context)
	response := ""

	for chunk := range responseChan {
		response += chunk
		pipeline.out <- chunk
	}

	pipeline.out <- reset + "\n"

	// Update history
	historyEntry := fmt.Sprintf("**AGENT ID: %s**\n\n%s", agent.ID, response)
	pipeline.history = append(pipeline.history, historyEntry)

	// Flow control
	if pipeline.Flow(response) {
		return
	}
}

/*
handleHistory manages the historical context of the story.
*/
func (pipeline *Pipeline) handleHistory() string {
	historicalContext := "\n<details>\n  <summary>History</summary>\n\n"
	historicalContext += fmt.Sprintf("\n\nSCENE: %d - ACTION: %d\n\n", pipeline.scene, pipeline.action)
	for _, entry := range pipeline.history {
		historicalContext += entry + "\n\n"
	}
	historicalContext += "</details>\n\n---\n\n"

	return historicalContext
}

/*
Flow controls the progression of the story.
*/
func (pipeline *Pipeline) Flow(action string) (shouldEnd bool) {
	flowDecision, err := pipeline.crew.Flow.Decide(action)
	if err != nil {
		errnie.Error(err.Error())
		return true
	}

	if flowDecision.Repeat {
		if flowDecision.Scope == "action" {
			pipeline.agent = 0
		} else if flowDecision.Scope == "scene" {
			if pipeline.scene > 0 {
				pipeline.scene--
			}
		}
	} else {
		pipeline.next()
	}

	return false
}

/*
next progresses the story to the next step.
*/
func (pipeline *Pipeline) next() {
	if pipeline.agent < len(pipeline.agents)-1 {
		pipeline.agent++
		return
	}

	pipeline.agent = 0

	if pipeline.action < len(pipeline.scenes[pipeline.scene].Contexts)-1 {
		pipeline.action++
		return
	}

	pipeline.action = 0

	if pipeline.scene < len(pipeline.scenes)-1 {
		pipeline.scene++
		pipeline.history = make([]string, 0)
		return
	}

	// Set a flag or handle end of simulation
}

/*
shouldEnd determines if the simulation should end.
*/
func (pipeline *Pipeline) shouldEnd() bool {
	// Implement logic to determine if the simulation should end
	return false
}

/*
Save persists the current state of the pipeline to a file.
*/
func (pipeline *Pipeline) Save() {
	state := PipelineState{
		Agents:  make([]*AgentState, 0),
		History: pipeline.history,
		Scenes:  pipeline.scenes,
		Scene:   pipeline.scene,
		Action:  pipeline.action,
		Agent:   pipeline.agent,
	}

	for _, agent := range pipeline.agents {
		state.Agents = append(state.Agents, agent.Save())
	}

	jsonData, err := json.Marshal(state)
	if err != nil {
		errnie.Error(err.Error())
		return
	}

	os.WriteFile("pipeline.json", jsonData, 0644)
}

/*
Load restores the pipeline state from a file.
*/
func (pipeline *Pipeline) Load() {
	data, err := os.ReadFile("pipeline.json")
	if err != nil {
		errnie.Error(err.Error())
		return
	}

	state := PipelineState{}
	if err := json.Unmarshal(data, &state); err != nil {
		errnie.Error(err.Error())
		return
	}

	pipeline.agents = make([]*Agent, 0)
	for _, agentState := range state.Agents {
		agent := NewAgent(pipeline.ctx, pipeline.conn, agentState.ID, agentState.Color)
		agent.Load()
		pipeline.agents = append(pipeline.agents, agent)
	}

	pipeline.history = state.History
	pipeline.scenes = state.Scenes
	pipeline.scene = state.Scene
	pipeline.action = state.Action
	pipeline.agent = state.Agent
}

/*
setScenes initializes the scenes and agents for the pipeline.
*/
func (pipeline *Pipeline) setScenes() {
	pipeline.scenes = make([]Scene, 0)
	steps := viper.GetStringSlice("ai.steps")

	for _, step := range steps {
		userSteps := viper.GetStringSlice("ai.contexts." + step)
		contexts := make([]string, 0)

		for _, userStep := range userSteps {
			contexts = append(contexts, viper.GetString("ai.prompts."+userStep))
		}

		scene := Scene{
			System:   viper.GetString("ai.systems." + step),
			Contexts: contexts,
		}

		pipeline.scenes = append(pipeline.scenes, scene)
	}

	// Initialize agents
	numAgents := 1 // Adjust as needed
	for i := 0; i < numAgents; i++ {
		seed := time.Now().UTC().UnixNano()
		nameGenerator := namegenerator.NewNameGenerator(seed)
		agentID := nameGenerator.Generate()
		agentColor := Colors[i%len(Colors)]
		agent := NewAgent(pipeline.ctx, pipeline.conn, agentID, agentColor)
		pipeline.agents = append(pipeline.agents, agent)
	}
}
