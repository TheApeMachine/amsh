package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/goombaio/namegenerator"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

// Pipeline orchestrates the execution of steps and agents.
type Pipeline struct {
	ctx    context.Context
	conn   *Conn
	out    chan string
	agents []*Agent // List of agents who complete the contexts
	steps  []string // List of steps, each with its system message and contexts
	scenes []Scene
	scene  int
	action int
	agent  int
}

type Scene struct {
	System   string   // System message
	Contexts []string // Map of context name to agents' IDs
}

// NewPipeline initializes the pipeline with agents and steps.
func NewPipeline(ctx context.Context, conn *Conn) *Pipeline {
	errnie.Debug("NewPipeline")
	return &Pipeline{
		ctx:    ctx,
		conn:   conn,
		out:    make(chan string),
		agents: make([]*Agent, 0),
		steps:  viper.GetViper().GetStringSlice("ai.steps"),
	}
}

// Generate runs the pipeline over all outer steps and inner agent steps.
func (pipeline *Pipeline) Generate() <-chan string {
	pipeline.setScenes()
	spew.Dump(pipeline)

	go func() {
		defer close(pipeline.out)

		history := []string{}

		for {
			scene := pipeline.scenes[pipeline.scene]
			action := scene.Contexts[pipeline.action]
			agent := pipeline.agents[pipeline.agent]
			response := ""

			// Construct the historic context
			historicContext := "\n<details>\n  <summary>History</summary>\n\n"
			for _, entry := range history {
				historicContext += entry + "\n\n"
			}
			historicContext += "</details>\n\n---\n\n"

			if pipeline.action == 0 {
				pipeline.out <- "\n\n---\n\n" + scene.System + "\n---\n"
			}

			context := fmt.Sprintf("%s\n\nYOUR AGENT ID: %s\n\n%s\n", historicContext, agent.ID, action)
			pipeline.out <- context

			for chunk := range agent.Generate(pipeline.ctx, scene.System, context) {
				response += chunk
				pipeline.out <- chunk
			}

			// Add the new entry to the history
			historyEntry := fmt.Sprintf("**AGENT ID: %s**\n\n  %s", agent.ID, response)
			history = append(history, historyEntry)

			if shouldEnd := pipeline.next(); shouldEnd {
				break
			}
		}
	}()

	return pipeline.out
}

/*
next is used to move the story forward.
The logic is as follows:
- If the current agent is the last agent, we need to move to the next action.
- If the current action is the last action, we need to move to the next scene.
- If the current scene is the last scene, we need to end the story.
*/
func (pipeline *Pipeline) next() (shouldEnd bool) {
	errnie.Debug("\n\nNext: %d - %d - %d\n\n", pipeline.scene, pipeline.action, pipeline.agent)

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

	for i := 0; i < 2; i++ {
		nameGenerator := namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())
		pipeline.addAgent(NewAgent(pipeline.ctx, pipeline.conn, nameGenerator.Generate()))
	}
}

/*
addAgent is used to be able to dynamically add agents to the pipeline.
We need this so the story can have new characters enter the story.
*/
func (pipeline *Pipeline) addAgent(agent *Agent) {
	pipeline.agents = append(pipeline.agents, agent)
}
