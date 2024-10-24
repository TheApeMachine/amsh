package mastercomputer

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/openai/openai-go"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/utils"
)

// Sequencer modified to support scoped conversation contexts for each worker.
type Sequencer struct {
	ctx        context.Context
	cancel     context.CancelFunc
	executor   *Executor
	workers    map[string][]*Worker
	activeTeam string
	output     *Output
}

func NewSequencer() *Sequencer {
	sequencer := &Sequencer{
		workers: make(map[string][]*Worker),
		output:  NewOutput(),
	}
	sequencer.executor = NewExecutor(sequencer)
	return sequencer
}

func (sequencer *Sequencer) Initialize() {
	sequencer.ctx, sequencer.cancel = context.WithCancel(context.Background())
	v := viper.GetViper()
	toolset := NewToolset()

	for _, role := range []string{"prompt", "reasoner", "researcher", "planner", "actor"} {
		for _, wrkr := range []string{utils.NewName(), utils.NewName(), utils.NewName()} {
			worker := NewWorker(sequencer.ctx, wrkr, toolset.Assign(role), sequencer.executor, role)
			system := v.GetString("ai.prompt.system")
			system = strings.ReplaceAll(system, "{name}", wrkr)
			system = strings.ReplaceAll(system, "{role}", v.GetString("ai.prompt."+role+".role"))
			worker.system = system
			worker.user = fmt.Sprintf("User prompt for worker %s", wrkr)
			worker.Initialize()
			sequencer.workers[role] = append(sequencer.workers[role], worker)
		}
	}
}

func (sequencer *Sequencer) Start() {
	// Start each worker in sequence with scoped context
	for role, workers := range sequencer.workers {
		for _, worker := range workers {
			log.Printf("Starting worker %s of role %s\n", worker.name, role)
			worker.Start()
		}
	}
}

func (sequencer *Sequencer) MessageHandler(message string) {
	// Use OpenAI to determine the appropriate role
	role := sequencer.RoleClassifier(message)

	if workers, exists := sequencer.workers[role]; exists {
		// Assign the message to the first available worker of that role.
		worker := workers[0]
		worker.buffer.AddMessage(openai.UserMessage(message))
		log.Printf("Assigned message to worker %s of role %s\n", worker.name, role)
		worker.Start()
	} else {
		log.Printf("No worker found for role: %s\n", role)
	}
}

func (sequencer *Sequencer) RoleClassifier(message string) string {
	prompt := fmt.Sprintf(`
        Analyze the following message and determine the most appropriate role to handle it:
        
        Message: "%s"

        Roles to choose from: ["prompt", "reasoner", "researcher", "planner", "actor"]

        Please respond with the most appropriate role based on the content of the message.
    `, message)

	params := openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		Model:       openai.F(openai.ChatModelGPT4oMini),
		Temperature: openai.Float(0.0), // Set low temperature for deterministic responses
	}

	response, err := sequencer.executor.executeCompletion(params)
	if err != nil {
		log.Printf("Error determining role: %s", err.Error())
		return "actor" // Default role in case of error
	}

	if len(response.Choices) > 0 {
		return strings.ToLower(strings.TrimSpace(response.Choices[0].Message.Content))
	}

	return "actor" // Default role if response is empty
}
