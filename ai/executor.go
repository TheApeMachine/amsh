package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tweaker"
)

type Pointer struct {
	Flow   int
	Scene  int
	Action int
	Agent  int
}

type History struct {
	Agent    string
	Response string
}

type Executor struct {
	ctx            context.Context
	conn           *Conn
	setup          tweaker.Setup
	template       tweaker.Template
	agents         map[string][]*Agent
	pointer        Pointer
	history        []History
	historyContext string
}

func NewExecutor(
	ctx context.Context,
	conn *Conn,
	setup tweaker.Setup,
	template tweaker.Template,
) *Executor {
	errnie.Trace()

	return &Executor{
		ctx:      ctx,
		conn:     conn,
		setup:    setup,
		template: template,
		agents:   make(map[string][]*Agent),
		pointer: Pointer{
			Scene:  0,
			Action: 0,
			Agent:  0,
		},
		history: make([]History, 0),
	}
}

func (executor *Executor) Initialize() {
	errnie.Trace()
}

/*
Generate the pipeline.
*/
func (executor *Executor) Generate() <-chan string {
	errnie.Trace()

	out := make(chan string)

	go func() {
		defer close(out)

		shouldEnd := false

		for {
			if shouldEnd {
				break
			}

			agent, system, user := executor.compile()

			for chunk := range agent.Generate(executor.ctx, system, user) {
				out <- chunk
			}

			shouldEnd = executor.next()
		}
	}()

	return out
}

/*
compile the templates.
*/
func (executor *Executor) compile() (agent *Agent, system, user string) {
	errnie.Trace()

	flow := executor.setup.Flow[executor.pointer.Flow]
	scene := executor.setup.Prompt.Script[executor.pointer.Scene]
	action := scene.Actions[executor.pointer.Action]
	agent = executor.agents[flow.Agent][executor.pointer.Agent]

	system = executor.template.System
	user = executor.template.User

	// Replace the placeholders with the setup values.
	system = strings.ReplaceAll(system, "<{name}>", executor.setup.Name)
	system = strings.ReplaceAll(system, "<{prefix}>", executor.setup.Prompt.Prefix)
	system = strings.ReplaceAll(system, "<{role}>", agent.Type)
	system = strings.ReplaceAll(system, "<{responsibilities}>", agent.Responsibilities)
	system = strings.ReplaceAll(system, "<{instructions}>", flow.Instructions)
	system = strings.ReplaceAll(system, "<{suffix}>", executor.setup.Prompt.Suffix)
	user = strings.ReplaceAll(user, "<{context}>", executor.historyContext)
	user = strings.ReplaceAll(user, "<{action}>", action.User)

	// Build the context history table.
	history := ""
	for _, response := range executor.history {
		history += fmt.Sprintf("| %s | %s |\n", response.Agent, response.Response)
	}

	return agent, system, user
}

/*
Add an agent to the executor.
*/
func (executor *Executor) AddAgent(agent *Agent) {
	errnie.Trace()

	executor.agents[agent.Type] = append(executor.agents[agent.Type], agent)
}

/*
next is used to move the story forward.
The logic is as follows:
- If the current agent is the last agent, we need to move to the next action.
- If the current action is the last action, we need to move to the next scene.
- If the current scene is the last scene, we need to end the story.
*/
func (executor *Executor) next() (shouldEnd bool) {
    errnie.Trace()

    // Move to the next agent
    if executor.pointer.Agent < len(executor.agents["flow"])-1 {
        executor.pointer.Agent++
        return false
    }

    // Reset agent and move to the next action
    executor.pointer.Agent = 0

    if executor.pointer.Action < len(executor.setup.Prompt.Script[executor.pointer.Scene].Actions)-1 {
        executor.pointer.Action++
        return false
    }

    // Reset action and move to the next scene
    executor.pointer.Action = 0

    if executor.pointer.Scene < len(executor.setup.Prompt.Script)-1 {
        executor.pointer.Scene++
        return false
    }

    // End the simulation when all scenes are processed
    return true
}
