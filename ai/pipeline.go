package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/goombaio/namegenerator"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

/*
colors is a list of ANSI escape codes for colored output.
*/
var colors = []string{
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[34m", // Blue
	"\033[35m", // Magenta
	"\033[36m", // Cyan
	"\033[40m", // Bright Green
	"\033[41m", // Bright Yellow
	"\033[42m", // Bright Blue
	"\033[43m", // Bright Magenta
	"\033[44m", // Bright Cyan
}

var reset = "\033[0m"

type PipelineState struct {
	Agents  []*AgentState
	History []string
	Scenes  []Scene
	Scene   int
	Action  int
	Agent   int
}

// Pipeline orchestrates the execution of steps and agents.
type Pipeline struct {
	ctx      context.Context
	conn     *Conn
	out      chan string
	crew     *Crew
	agents   []*Agent
	history  []string
	steps    []string
	scenes   []Scene
	scene    int
	action   int
	agent    int
	loglevel string
	err      error
}

type Scene struct {
	System   string
	Contexts []string
}

// NewPipeline initializes the pipeline with agents and steps.
func NewPipeline(ctx context.Context, conn *Conn) *Pipeline {
	errnie.Debug("NewPipeline")
	return &Pipeline{
		ctx:      ctx,
		conn:     conn,
		out:      make(chan string),
		crew:     NewCrew(ctx, conn),
		agents:   make([]*Agent, 0),
		history:  make([]string, 0),
		steps:    viper.GetViper().GetStringSlice("ai.steps"),
		loglevel: viper.GetViper().GetString("loglevel"),
	}
}

// Generate runs the pipeline over all outer steps and inner agent steps.
func (pipeline *Pipeline) Generate() <-chan string {
	pipeline.setScenes()

	if pipeline.loglevel == "debug" {
		spew.Dump(pipeline)
	}

	go func() {
		defer close(pipeline.out)

		for {
			scene := pipeline.scenes[pipeline.scene]
			action := scene.Contexts[pipeline.action]
			agent := pipeline.agents[pipeline.agent]
			response := ""

			// Construct the historic context
			historicContext := pipeline.handleHistory()

			if pipeline.action == 0 && pipeline.loglevel == "debug" {
				pipeline.out <- "\n\n---\n\n" + scene.System + "\n---\n"
			}

			context := fmt.Sprintf("%s\n\nYOUR AGENT ID: %s\n\n%s\n", historicContext, agent.ID, action)

			if pipeline.loglevel == "debug" {
				pipeline.out <- context
			}

			pipeline.out <- agent.color + "\n"

			script := pipeline.Write()

			for chunk := range agent.Generate(pipeline.ctx, script.Scene, script.Actions[pipeline.action]) {
				response += chunk
				pipeline.out <- chunk
			}

			pipeline.out <- reset + "\n"

			// Add the new entry to the history
			historyEntry := fmt.Sprintf("**AGENT ID: %s**\n\n  %s", agent.ID, response)
			pipeline.history = append(pipeline.history, historyEntry)

			if shouldEnd := pipeline.Flow(response); shouldEnd {
				break
			}
		}
	}()

	return pipeline.out
}

/*
handleHistory handles the history of the story.
*/
func (pipeline *Pipeline) handleHistory() (historicalContext string) {
	historicalContext = "\n<details>\n  <summary>History</summary>\n\n"
	historicalContext += fmt.Sprintf("\n\nSCENE: %d - ACTION: %d\n\n", pipeline.scene, pipeline.action)
	for _, entry := range pipeline.history {
		historicalContext += entry + "\n\n"
	}
	historicalContext += "</details>\n\n---\n\n"

	return
}

/*
Direct is used to direct the story.
*/
func (pipeline *Pipeline) Direct() (response *Direction) {
	if response, pipeline.err = pipeline.crew.Direct(
		strings.Join(pipeline.history, "\n\n"),
	); pipeline.err != nil {
		errnie.Error(pipeline.err.Error())
		return
	}

	return response
}

/*
Write the prompts for the agents to follow.
*/
func (pipeline *Pipeline) Write() (response *Script) {
	if response, pipeline.err = pipeline.crew.Write(
		strings.Join(pipeline.history, "\n\n"),
		pipeline.Direct(),
	); pipeline.err != nil {
		errnie.Error(pipeline.err.Error())
		return
	}

	return response
}

/*
Flow is used to control the flow of the story.
*/
func (pipeline *Pipeline) Flow(action string) (shouldEnd bool) {
	// Check if we need to continue or repeat the action.
	if flow, err := pipeline.crew.Flow(action); err != nil {
		errnie.Error(err.Error())
		return
	} else if flow.Flow == "repeat" {
		if flow.Scope == "action" {
			pipeline.agent = 0
		} else if flow.Scope == "scene" {
			pipeline.scene--
		}
	} else {
		if shouldEnd := pipeline.next(); shouldEnd {
			return true
		}
	}

	return
}

/*
Save the pipeline state to a file.
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

	json, err := json.Marshal(state)
	if err != nil {
		errnie.Error(err.Error())
		return
	}

	os.WriteFile("pipeline.json", json, 0644)
}

/*
Load the pipeline state from a file.
*/
func (pipeline *Pipeline) Load() {
	buf, err := os.ReadFile("pipeline.json")
	if err != nil {
		errnie.Error(err.Error())
		return
	}

	state := PipelineState{}
	if err := json.Unmarshal(buf, &state); err != nil {
		errnie.Error(err.Error())
		return
	}

	pipeline.agents = make([]*Agent, 0)
	for _, agent := range state.Agents {
		as := NewAgent(pipeline.ctx, pipeline.conn, agent.ID, agent.Color)
		as.Load()
		pipeline.addAgent(as)
	}

	pipeline.history = state.History
	pipeline.scenes = state.Scenes
	pipeline.scene = state.Scene
	pipeline.action = state.Action
	pipeline.agent = state.Agent
}

/*
next is used to move the story forward.
The logic is as follows:
- If the current agent is the last agent, we need to move to the next action.
- If the current action is the last action, we need to move to the next scene.
- If the current scene is the last scene, we need to end the story.
*/
func (pipeline *Pipeline) next() (shouldEnd bool) {
	if pipeline.loglevel == "debug" {
		errnie.Debug("\n\nNext: %d - %d - %d\n\n", pipeline.scene, pipeline.action, pipeline.agent)
	}

	// Move to the next agent
	if pipeline.agent < len(pipeline.agents)-1 {
		pipeline.agent++
		return false
	}

	// Reset agent and move to the next action
	pipeline.agent = 0

	if pipeline.action < len(pipeline.scenes[pipeline.scene].Contexts)-1 {
		pipeline.action++
		return false
	}

	// Reset action and move to the next scene
	pipeline.action = 0

	if pipeline.scene < len(pipeline.scenes)-1 {
		pipeline.scene++

		// Update the profile of all agents and reset the history.
		for _, agent := range pipeline.agents {
			pipeline.crew.UpdateProfile(agent)
		}

		pipeline.history = make([]string, 0)

		return false
	}

	// If the final scene has been processed, end the story
	return true
}

/*
setScenes sets the scene for the pipeline.
*/
func (pipeline *Pipeline) setScenes() {
	pipeline.scenes = make([]Scene, 0)

	for _, step := range pipeline.steps {
		userSteps := viper.GetViper().GetStringSlice("ai.contexts." + step)
		contexts := make([]string, 0)

		for _, userStep := range userSteps {
			contexts = append(contexts, viper.GetString("ai.prompts."+userStep))
		}

		scene := Scene{
			System:   viper.GetViper().GetString("ai.systems." + step),
			Contexts: contexts,
		}

		pipeline.scenes = append(pipeline.scenes, scene)
	}

	for i := 0; i < 1; i++ {
		nameGenerator := namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())
		pipeline.addAgent(
			NewAgent(
				pipeline.ctx,
				pipeline.conn,
				nameGenerator.Generate(),
				colors[len(pipeline.crew.agents)+i%len(colors)],
			),
		)
	}
}

/*
addAgent is used to be able to dynamically add agents to the pipeline.
We need this so the story can have new characters enter the story.
*/
func (pipeline *Pipeline) addAgent(agent *Agent) {
	pipeline.agents = append(pipeline.agents, agent)
}
