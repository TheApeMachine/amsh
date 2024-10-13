package berrt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/TylerBrock/colorjson"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/pkoukk/tiktoken-go"
	"github.com/theapemachine/amsh/container"
)

var Dark = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#666666")).Render
var Muted = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#999999")).Render
var Highlight = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#EEEEEE")).Render
var Blue = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#6E95F7")).Render
var Red = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#F7746D")).Render
var Yellow = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#F7B96D")).Render
var Green = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#06C26F")).Render

func PrettyJSON(v any) string {
	f := colorjson.NewFormatter()
	f.Indent = 2
	s, _ := f.Marshal(v)
	return string(s)
}

type ErrorAI struct {
	ctx              context.Context
	initialSystem    string
	initialUser      string
	runner           *container.Runner
	conversationLog  []openai.ChatCompletionMessageParamUnion
	maxContextTokens int
	tokenCounts      []int64 // Store token counts for each message
	done             bool
}

func NewErrorAI(message, stacktrace, snippet string) *ErrorAI {
	runner, err := container.NewRunner()
	if err != nil {
		fmt.Printf("Error creating runner: %v\n", err)
	}

	return &ErrorAI{
		ctx: context.Background(),
		initialSystem: `
		You are a helpful assistant that helps me debug my code...
		You are currently inside an isolated environment that has the code I need to debug.
		This means that when you see file paths, they are relative to /tmp/workspace.
		Example: /Users/theapemachine/go/src/github.com/theapemachine/amsh/mastercomputer/worker.go
		Becomes: /tmp/workspace/amsh/mastercomputer/worker.go

		Start by cloning the repository:

		cd /tmp/workspace
		git clone git@github.com:TheApeMachine/amsh.git
		cd amsh

		Make sure to use the ssh method to clone the repository, not https. Your SSH key is set up in the container.

		You may have to do some git config stuff first, your username and email are:

		git config --global user.name "marvinnotafan"
		git config --global user.email "development@fanfactory.nl"

		You should always create a new git branch before making any changes, using the: aibugfix/<branchname> convention.
		You should also open a PR early, before making any changes, and keep it updated as you work.
		Each time you push to the PR, you will receive a code review, which can be used to guide your work.

		If you write any tests, which is a good idea, you should use Goconvey. You can run the tests with:
		
		go test -v ./...

		REMEMBER: 
		1. Not all Linux command give an output, in such cases you should just check if the command was successful manually.
		2. Always know where you are in the filesystem, and what you current working directory contains.
		3. Using tree can be very helpful to get an overview of the filesystem.
		4. Install any dependencies you might need using the package manager, remember to update it first.
		5. Don't keep trying the same action over and over again, that's not how this works, switch strategy when you get stuck.
		
		You need to think, and reason your way through the problem, so after each tool call, and after each command execution,
		you should think about what you just did, and what the next best step is, and verify that you are on the right track.
		`,
		initialUser: fmt.Sprintf(
			"I am getting the following error:\n\n%s\n\nWith this stacktrace:\n\n%s\n\nAnd relevant snippet:\n\n%s\n\n",
			message, stacktrace, snippet,
		),
		runner:           runner,
		maxContextTokens: 128000, // Adjust based on the model's actual limit
	}
}

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

func makeTool(name, description string, schema openai.FunctionParameters) openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String(name),
			Description: openai.String(description),
			Parameters:  openai.F(schema),
		}),
	}
}

type Completion struct {
	ctx    context.Context
	client *openai.Client
}

func NewCompletion(ctx context.Context) *Completion {
	return &Completion{
		ctx:    ctx,
		client: openai.NewClient(),
	}
}

func (completion *Completion) Execute(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	response, err := completion.client.Chat.Completions.New(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("error from OpenAI API: %w", err)
	}

	if response == nil {
		return nil, errors.New("received nil response from OpenAI")
	}

	return response, nil
}

func (ai *ErrorAI) Execute() {
	ctx, cancel := context.WithCancel(ai.ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println(Red("\nReceived interrupt signal. Shutting down..."))
		cancel()
	}()

	defer ai.cleanup()
	// if err := ai.prepareWorkspace(); err != nil {
	// 	fmt.Printf("Error preparing workspace: %v\n", err)
	// 	return
	// }

	in, out, err := ai.startContainer(ctx)
	if err != nil {
		log.Error("Error starting container", "error", err)
		return
	}
	defer in.Close()
	defer out.Close()

	// Initialize conversation log with initial messages
	ai.conversationLog = []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(ai.initialSystem),
		openai.UserMessage(ai.initialUser),
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if ai.done {
				return
			}

			params := ai.getParamsWithManagedContext()

			log.Info("executing completion...")

			response, err := ai.executeCompletion(ctx, params)
			if err != nil {
				log.Error("Error executing completion", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}

			log.Info("processing response...")
			ai.processResponse(response)
			time.Sleep(1 * time.Second)
		}
	}
}

func (ai *ErrorAI) executeCompletion(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	completion := NewCompletion(ctx)
	response, err := completion.Execute(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("error from OpenAI API: %w", err)
	}

	if response == nil {
		return nil, errors.New("received nil response from OpenAI")
	}

	return response, nil
}

func (ai *ErrorAI) getParamsWithManagedContext() openai.ChatCompletionNewParams {
	// Truncate conversation to fit within token limit
	messages := ai.truncateConversation()

	return openai.ChatCompletionNewParams{
		Messages: openai.F(messages),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        openai.F("error_analysis"),
					Description: openai.F("Analyze the error and provide a plan to resolve it"),
					Schema:      openai.F(GenerateSchema[ErrorAnalysis]()),
					Strict:      openai.Bool(true),
				}),
			},
		),
		Tools: openai.F([]openai.ChatCompletionToolParam{
			makeTool(
				"bash_command",
				"Execute a bash command in the container.",
				openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"command": map[string]string{
							"type":        "string",
							"description": "The bash command to execute",
						},
					},
					"required": []string{"command"},
				},
			),
			makeTool(
				"browser",
				"Use a fully functional chrome browser to browse the web.",
				openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"url": map[string]string{
							"type":        "string",
							"description": "The URL to browse",
						},
						"javascript": map[string]string{
							"type":        "string",
							"description": "The javascript to execute in the development console. Must be a function that returns a value.",
						},
					},
					"required": []string{"url", "javascript"},
				},
			),
		}),
		Model:       openai.F(openai.ChatModelGPT4o),
		Temperature: openai.Float(0.0),
	}
}

func (ai *ErrorAI) truncateConversation() []openai.ChatCompletionMessageParamUnion {
	maxTokens := ai.maxContextTokens - 500 // Reserve tokens for response
	totalTokens := 0
	var truncatedMessages []openai.ChatCompletionMessageParamUnion

	// Start from the most recent message
	for i := len(ai.conversationLog) - 1; i >= 0; i-- {
		msg := ai.conversationLog[i]
		messageTokens := ai.estimateTokens(msg)
		if totalTokens+messageTokens <= maxTokens {
			truncatedMessages = append([]openai.ChatCompletionMessageParamUnion{msg}, truncatedMessages...)
			totalTokens += messageTokens
		} else {
			break
		}
	}

	return truncatedMessages
}

func (ai *ErrorAI) processResponse(response *openai.ChatCompletion) {
	if response.Usage.CompletionTokens > 0 {
		ai.updateTokenCounts(response.Usage)
	}

	userMessage, err := ai.extractAndPrintResponse(response)
	if err != nil {
		log.Error(err)
		return
	}

	if userMessage != nil {
		ai.updateConversationLog(userMessage)
	}

	toolMessage := ai.handleToolCalls(response)
	if toolMessage != nil {
		ai.updateConversationLog(toolMessage)
	}
}

func (ai *ErrorAI) extractAndPrintResponse(response *openai.ChatCompletion) (openai.ChatCompletionMessageParamUnion, error) {
	if response == nil || len(response.Choices) == 0 {
		log.Error("No response from OpenAI")
		return nil, nil
	}

	content := response.Choices[0].Message.Content
	if content == "" {
		return nil, nil
	}

	reasoning := ErrorAnalysis{}
	if err := json.Unmarshal([]byte(content), &reasoning); err != nil {
		return nil, err
	}

	for _, step := range reasoning.Steps {
		fmt.Println(Red("Thought:"), Highlight(step.Thought))
		fmt.Println(Green("Action:"), Highlight(step.Action))
	}

	fmt.Println(Blue("Plan:\n\n"), Highlight(reasoning.Plan))

	ai.done = reasoning.Done
	return openai.AssistantMessage(content), nil
}

func (ai *ErrorAI) handleToolCalls(response *openai.ChatCompletion) openai.ChatCompletionMessageParamUnion {
	if response == nil || len(response.Choices) == 0 {
		log.Error("No response from OpenAI")
		return nil
	}

	message := response.Choices[0].Message
	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		log.Error("No tool calls in response")
		return nil
	}

	for _, toolCall := range message.ToolCalls {
		var args map[string]any
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			log.Error("Error unmarshalling tool call arguments", "error", err)
			continue
		}

		switch toolCall.Function.Name {
		case "bash_command":
			fmt.Println(Dark("Executing command"), Muted(args["command"].(string)))
			commandStr := args["command"].(string)
			commandParts := strings.Fields(commandStr) // Split command into arguments
			buf := ai.runner.ExecuteCommand(ai.ctx, commandParts)

			if len(buf) == 0 {
				buf = []byte("Command executed successfully, but no output was produced. Check if the command was successful manually.")
			}

			fmt.Println(Muted(string(buf)))
			ai.updateConversationLog(message)
			return openai.ToolMessage(toolCall.ID, string(buf))
		case "browser":
			log.Info(Dark("Browsing"), Muted(args["url"].(string)))

			browser := NewBrowser()
			output, err := browser.Run(args)

			if err != nil {
				log.Error(err)
				return openai.ToolMessage(toolCall.ID, err.Error())
			}

			ai.updateConversationLog(message)
			return openai.ToolMessage(toolCall.ID, string(output))
		}
	}

	log.Warn("No tool calls matched")
	return nil
}

func (ai *ErrorAI) updateTokenCounts(usage openai.CompletionUsage) {
	ai.tokenCounts = append(ai.tokenCounts, usage.TotalTokens)
}

func (ai *ErrorAI) updateConversationLog(message openai.ChatCompletionMessageParamUnion) {
	ai.conversationLog = append(ai.conversationLog, message)
}

func (ai *ErrorAI) cleanup() {
	log.Info("Stopping container...")
	if err := ai.runner.StopContainer(context.Background()); err != nil {
		log.Error("Error stopping container", "error", err)
	}
}

func (ai *ErrorAI) prepareWorkspace() error {
	log.Info("copying files to /tmp/workspace...")
	if err := os.RemoveAll("/tmp/workspace/amsh"); err != nil {
		return err
	}

	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error("error walking", "error", err)
			return err
		}

		ignoreDirs := []string{"logs", "tmp", "node_modules"}
		if info.IsDir() {
			for _, ignoreDir := range ignoreDirs {
				if info.Name() == ignoreDir {
					log.Warn("skipping directory", "directory", path)
					return filepath.SkipDir
				}
			}
		}

		relPath, err := filepath.Rel(".", path)
		if err != nil {
			log.Error("error getting relative path", "error", err)
			return err
		}

		destPath := filepath.Join("/tmp/workspace/amsh", relPath)
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		log.Info("copying file", "source", path, "destination", destPath)
		cmd := exec.Command("cp", path, destPath)
		return cmd.Run()
	})
}

func (ai *ErrorAI) startContainer(ctx context.Context) (in io.WriteCloser, out io.ReadCloser, err error) {
	imageName := "berrt:latest"
	cmd := []string{"/bin/bash"}
	username := "debug-user"
	customMessage := "Debug environment ready. Use /tmp/workspace as your working directory. It should already have the code waiting for you."

	builder, err := container.NewBuilder()
	if err != nil {
		return nil, nil, fmt.Errorf("error creating builder: %w", err)
	}
	if err := builder.BuildImage(ctx, "./container", imageName); err != nil {
		return nil, nil, fmt.Errorf("error building image: %w", err)
	}

	in, out, err = ai.runner.RunContainer(ctx, imageName, cmd, username, customMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("error running container: %w", err)
	}

	return
}

type ErrorAnalysis struct {
	Steps []Step `json:"steps" jsonschema_description:"The steps to take to resolve the error"`
	Plan  string `json:"plan" jsonschema_description:"The plan to resolve the error"`
	Done  bool   `json:"done" jsonschema_description:"Whether the error has been resolved"`
}

type Step struct {
	Thought string `json:"thought" jsonschema_description:"The thought process to take to resolve the error"`
	Action  string `json:"action" jsonschema_description:"The action to take to resolve or further analyze the error"`
}

func (ai *ErrorAI) estimateTokens(msg openai.ChatCompletionMessageParamUnion) int {
	content := ""
	role := ""
	switch m := msg.(type) {
	case openai.ChatCompletionUserMessageParam:
		content = m.Content.String()
		role = "user"
	case openai.ChatCompletionAssistantMessageParam:
		content = m.Content.String()
		role = "assistant"
	case openai.ChatCompletionSystemMessageParam:
		content = m.Content.String()
		role = "system"
	case openai.ChatCompletionToolMessageParam:
		content = m.Content.String()
		role = "function"
	}

	// Use tiktoken-go to estimate tokens
	encoding, err := tiktoken.EncodingForModel("gpt-4o-mini")
	if err != nil {
		log.Error("Error getting encoding", "error", err)
		return 0
	}

	tokensPerMessage := 4 // As per OpenAI's token estimation guidelines

	numTokens := tokensPerMessage
	numTokens += len(encoding.Encode(content, nil, nil))
	if role == "user" || role == "assistant" || role == "system" || role == "function" {
		numTokens += len(encoding.Encode(role, nil, nil))
	}

	return numTokens
}
