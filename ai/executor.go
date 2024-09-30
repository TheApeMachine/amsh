package ai

import (
	"context"
	"strings"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tweaker"
)

var parsers = map[string][]Parser{
	"director": {
		&Direction{},
	},
	"writer": {
		&Side{},
		&Character{},
		&Location{},
	},
	"editor": {
		&Edit{},
	},
	"producer": {
		&Extract{},
	},
}

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
	setup          map[string]interface{}
	agents         map[string][]*Agent
	pointer        Pointer
	history        []History
	crewHistory    []string
	historyContext string
}

func NewExecutor(
	ctx context.Context,
	conn *Conn,
) *Executor {
	errnie.Trace()

	return &Executor{
		ctx:    ctx,
		conn:   conn,
		setup:  tweaker.Setups(),
		agents: make(map[string][]*Agent),
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
			history := History{
				Agent:    agent.ID,
				Response: "",
			}

			out <- "\n\n---\n\n"
			out <- system
			out <- user

			for chunk := range agent.Generate(executor.ctx, system, user) {
				history.Response += chunk
				out <- chunk
			}

			out <- "\n\n---\n\n"

			// Parse the response
			if parser, ok := parsers[agent.Type]; ok {
				for _, p := range parser {
					if err := p.Parse(history.Response); err != nil {
						errnie.Error(err)
						continue
					}

					executor.crewHistory = append(executor.crewHistory, p.Markdown())
				}
			}

			executor.history = append(executor.history, history)
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

	// Extract flow and script dynamically from the config
	flowConfig := tweaker.Flow()
	sceneConfig := executor.setup["prompt"].(map[string]interface{})["script"].([]interface{})[executor.pointer.Scene].(map[string]interface{})
	actionConfig := sceneConfig["actions"].([]interface{})[executor.pointer.Action].(map[string]interface{})

	currentAgentType := flowConfig[executor.pointer.Flow]["agent"].(string)
	agent = executor.agents[currentAgentType][executor.pointer.Agent]

	// Retrieve the prompt templates dynamically
	system = tweaker.PromptTemplate("system")
	user = tweaker.PromptTemplate("user")

	// Replace placeholders with the current values from config
	system = strings.ReplaceAll(system, "<{name}>", executor.setup["name"].(string))
	system = strings.ReplaceAll(system, "<{prefix}>", tweaker.PromptPrefix())
	system = strings.ReplaceAll(system, "<{role}>", agent.Type)
	system = strings.ReplaceAll(system, "<{responsibilities}>", agent.Responsibilities)
	system = strings.ReplaceAll(system, "<{instructions}>", flowConfig[executor.pointer.Flow]["instructions"].(string))
	system = strings.ReplaceAll(system, "<{suffix}>", tweaker.PromptSuffix())

	if agent.Type == "crew" && len(executor.crewHistory) > 0 {
		user = strings.ReplaceAll(user, "<{context}>", strings.Join(executor.crewHistory, "\n\n"))
	} else if agent.Type == "worker" && len(executor.history) > 0 {
		user = strings.ReplaceAll(user, "<{context}>", executor.historyContext)
	} else {
		user = strings.ReplaceAll(user, "<{context}>", "No historical context to display.")
	}

	if agent.Type == "worker" {
		user = strings.ReplaceAll(user, "<{action}>", actionConfig["user"].(string))
	}

	// Append the current agent's response to the history context
	// for _, history := range executor.history {
	// 	executor.historyContext += fmt.Sprintf("| %s | %s |\n", history.Agent, history.Response)
	// }
	// executor.historyContext += fmt.Sprintf("| %s | %s |\n", agent.ID, "<agent response>")

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

	// Get the current agent type
	flowConfig := tweaker.Flow()
	currentAgentType := flowConfig[executor.pointer.Flow]["agent"].(string)

	// Check if all replicas of the current agent type have completed
	if executor.pointer.Agent < len(executor.agents[currentAgentType])-1 {
		executor.pointer.Agent++
		return false
	}

	// Reset agent pointer and move to the next agent type
	executor.pointer.Agent = 0

	// Check if there are more agent types in the current flow to process
	if executor.pointer.Flow < len(flowConfig)-1 {
		executor.pointer.Flow++
		return false
	}

	// Reset flow pointer and move to the next action
	executor.pointer.Flow = 0

	sceneConfig := executor.setup["prompt"].(map[string]interface{})["script"].([]interface{})[executor.pointer.Scene].(map[string]interface{})
	if executor.pointer.Action < len(sceneConfig["actions"].([]interface{}))-1 {
		executor.pointer.Action++
		return false
	}

	// Reset action pointer and move to the next scene
	executor.pointer.Action = 0

	if executor.pointer.Scene < len(executor.setup["prompt"].(map[string]interface{})["script"].([]interface{}))-1 {
		executor.pointer.Scene++
		return false
	}

	// End the simulation if all scenes have been processed
	return true
}
