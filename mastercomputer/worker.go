package mastercomputer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/format"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Worker struct {
	ctx         context.Context
	cancel      context.CancelFunc
	buffer      *data.Artifact
	memory      *ai.Memory
	status      string
	Function    *openai.FunctionDefinition
	state       WorkerState
	environment *Environment
}

func NewWorker() *Worker {
	errnie.Trace()

	return &Worker{
		status: "creating",
		Function: &openai.FunctionDefinition{
			Name:        "worker",
			Description: "Use to create a worker agent, which can become anything you can imagine, using the system and user prompt, and providing a toolset.",
			Strict:      true,
			Parameters: jsonschema.Definition{
				Type:                 jsonschema.Object,
				AdditionalProperties: false,
				Properties: map[string]jsonschema.Definition{
					"system": {
						Type:        jsonschema.String,
						Description: "The system prompt",
					},
					"user": {
						Type:        jsonschema.String,
						Description: "The user prompt",
					},
					"toolset": {
						Type:        jsonschema.String,
						Description: "The toolset to use, or 'none' if no toolset is required",
						Enum:        []string{"none", "system", "research", "code"},
					},
				},
				Required: []string{"system", "user", "toolset"},
			},
		},
	}
}

func (worker *Worker) Initialize(ctx context.Context, buffer *data.Artifact) *Worker {
	errnie.Trace()
	worker.status = "initializing"

	worker.ctx, worker.cancel = context.WithCancel(ctx)
	worker.buffer = buffer
	worker.memory = ai.NewMemory()
	worker.state = WaitingForPrompt

	if worker.buffer != nil {
		worker.memory.Add(utils.JoinWith("\n",
			utils.Muted("  [ORIGINAL USER PROMPT]"),
			"    "+worker.buffer.Peek("user"),
			utils.Muted("  [/ORGINAL USER PROMPT]"),
		))

		worker.memory.Add(utils.JoinWith("\n",
			utils.Muted("  [GUIDELINES]"),
			"    "+worker.buffer.Peek("guidelines"),
			utils.Muted("  [/GUIDELINES]"),
		))
	} else {
		fmt.Println("Warning: buffer is nil during initialization")
	}

	return worker
}

func (worker *Worker) Inspect() string {
	return strings.Join([]string{
		utils.Dark("  [WORKER]"),
		utils.Blue("    Status: ") + utils.Highlight(worker.status),
		utils.Muted("    [LOG]"),
		worker.memory.String(),
		utils.Muted("    [/LOG]"),
		utils.Dark("  [/WORKER]"),
	}, "\n\n")
}

func (worker *Worker) Run(ctx context.Context) *Worker {
	errnie.Trace()
	worker.status = "running"

	for _, strategy := range worker.ExtractReasoningStrategy(
		worker.withStrategy("reasoning_strategy"),
	) {
		errnie.Info("applying strategy: %s", strategy)
		worker.withStrategy(strategy)
	}

	worker.withStrategy("self_reflection")
	worker.withStrategy("")

	if worker.environment != nil {
		worker.ConnectToEnvironment(worker.environment)
	}

	worker.withStrategy("final")
	return worker
}

func (worker *Worker) ExtractReasoningStrategy(response openai.ChatCompletionMessage) []string {
	errnie.Trace()
	buf := strategymap["reasoning_strategy"]
	if response.Content == "" {
		fmt.Println("Warning: Reasoning strategy response content is empty")
		return []string{}
	}
	errnie.Error(json.Unmarshal([]byte(response.Content), &buf.Strategy))
	return strings.Split(buf.FinalAnswer(), ",")
}

func (worker *Worker) withStrategy(strategy string) openai.ChatCompletionMessage {
	errnie.Trace()

	var (
		selectedStrategy *format.Response
		instructions     string
	)

	if strategy != "" {
		errnie.Warn("selected strategy: %s", strategy)
		selectedStrategy = strategymap[strategy]
		instructions = utils.StrategyInstructions(strategy)
	}

	systemPrompt := worker.buffer.Peek("system")
	if systemPrompt == "" {
		fmt.Println("Warning: system prompt is empty")
	}

	fmt.Println(utils.JoinWith("\n\n",
		systemPrompt,
		worker.memory.String(),
		instructions,
	))

	return worker.printResponse(NewCompletion(worker.ctx).Execute(
		systemPrompt,
		utils.JoinWith("\n\n",
			stripansi.Strip(worker.memory.String()),
			instructions,
		),
		worker.buffer.Peek("toolset"),
		selectedStrategy,
	), strategy)
}

func (worker *Worker) printResponse(response openai.ChatCompletionResponse, strategy string) openai.ChatCompletionMessage {
	errnie.Trace()

	if len(response.Choices) == 0 {
		fmt.Println("Warning: No choices in AI response")
		return openai.ChatCompletionMessage{}
	}

	message := response.Choices[0].Message

	if len(message.ToolCalls) > 0 {
		for _, toolCall := range message.ToolCalls {
			worker.memory.ShortTerm = append(
				worker.memory.ShortTerm,
				fmt.Sprintf(
					"[%s (%s)]\n  %s\n[/%s]",
					toolCall.Function.Name,
					toolCall.Function.Arguments,
					worker.useTool(toolCall),
					toolCall.Function.Name,
				),
			)
		}
	}

	if strategy == "" {
		errnie.Warn("no strategy found")
		return message
	}

	if message.Content != "" {
		content := strategymap[strategy]

		if errnie.Error(json.Unmarshal([]byte(message.Content), content.Strategy)) != nil {
			return message
		}

		worker.memory.ShortTerm = append(worker.memory.ShortTerm, content.ToString())
		fmt.Println(content.ToString())
	}

	return message
}

func (worker *Worker) useTool(toolCall openai.ToolCall) string {
	errnie.Trace()

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return errnie.Error(err).Error()
	}

	utils.BeautifyToolCall(toolCall, args)

	switch toolCall.Function.Name {
	case "worker":
		return NewWorker().Initialize(worker.ctx, data.New(
			worker.buffer.Peek("origin"),
			"prompt",
			"task",
			[]byte(args["user"].(string)),
		)).Run(worker.ctx).Inspect()
	case "environment":
		env := NewEnvironment()
		err := env.Start(worker.ctx, args["shell"].(string))
		if err != nil {
			return err.Error()
		}
		worker.environment = env
		worker.ConnectToEnvironment(env)
		return ""
	}

	return "unknown tool " + toolCall.Function.Name
}

func (worker *Worker) ConnectToEnvironment(env *Environment) error {
	errnie.Trace()
	worker.environment = env

	go func() {
		for {
			select {
			case <-worker.ctx.Done():
				return
			default:
				envOutput := env.ReadOutputBuffer()
				if envOutput != "" {
					response := worker.handleEnvironmentOutput(envOutput)
					if err := env.WriteToStdin(response); err != nil {
						fmt.Println("Error writing to environment stdin:", err)
					}
				}
				time.Sleep(100 * time.Millisecond) // Adjust polling interval as needed
			}
		}
	}()
	return nil
}

func (worker *Worker) handleEnvironmentOutput(envOutput string) string {
	errnie.Trace()

	response := NewCompletion(worker.ctx).Execute(
		worker.buffer.Peek("system"),
		utils.JoinWith("\n\n", stripansi.Strip(envOutput)),
		"none",
		strategymap["environment_interaction"],
	)

	if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
		fmt.Println("Warning: Empty AI response received. Returning fallback response.")
		return "echo 'Fallback command executed'"
	}

	return response.Choices[0].Message.Content
}

var strategymap = map[string]*format.Response{
	"reasoning_strategy":      format.NewResponse("reasoning_strategy", format.NewReasoningStrategy()),
	"self_reflection":         format.NewResponse("self_reflection", format.NewSelfReflection()),
	"chain_of_thought":        format.NewResponse("chain_of_thought", format.NewChainOfThought()),
	"tree_of_thought":         format.NewResponse("tree_of_thought", format.NewTreeOfThought()),
	"first_principles":        format.NewResponse("first_principles", format.NewFirstPrinciplesReasoning()),
	"step_by_step_execution":  format.NewResponse("step_by_step_execution", format.NewStepByStepExecution()),
	"pros_and_cons_analysis":  format.NewResponse("pros_and_cons_analysis", &format.ProsAndConsAnalysis{}),
	"divide_and_conquer":      format.NewResponse("divide_and_conquer", format.NewDivideAndConquer()),
	"hypothesis_testing":      format.NewResponse("hypothesis_testing", format.NewHypothesisTesting()),
	"risk_assessment":         format.NewResponse("risk_assessment", format.NewRiskAssessment()),
	"backwards_reasoning":     format.NewResponse("backwards_reasoning", format.NewBackwardReasoning()),
	"roleplay_simulation":     format.NewResponse("roleplay_simulation", format.NewRolePlaySimulation()),
	"scenario_analysis":       format.NewResponse("scenario_analysis", format.NewScenarioAnalysis()),
	"counterfactual_thinking": format.NewResponse("counterfactual", format.NewCounterfactualThinking()),
	"final":                   format.NewResponse("final", format.NewFinal()),
	"environment_interaction": format.NewResponse("environment_interaction", format.NewEnvironmentInteraction()),
}
