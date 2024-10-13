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
	"slices"
	"syscall"
	"time"

	"github.com/TylerBrock/colorjson"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
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
	toolMessages     []openai.ChatCompletionToolMessageParam
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
		ctx:           context.Background(),
		initialSystem: "You are a helpful assistant that helps me debug my code...",
		initialUser: fmt.Sprintf(
			"I am getting the following error:\n\n%s\n\nWith this stacktrace:\n\n%s\n\nAnd relevant snippet:\n\n%s\n\n",
			message, stacktrace, snippet,
		),
		runner:           runner,
		maxContextTokens: 4000, // Adjust based on the model's actual limit
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
		spew.Dump(params)
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
		fmt.Println("\nReceived interrupt signal. Shutting down...")
		cancel()
	}()

	defer ai.cleanup()
	if err := ai.prepareWorkspace(); err != nil {
		fmt.Printf("Error preparing workspace: %v\n", err)
		return
	}

	in, out, err := ai.startContainer(ctx)
	if err != nil {
		fmt.Printf("Error starting container: %v\n", err)
		return
	}
	defer in.Close()
	defer out.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if ai.done {
				return
			}

			params := ai.getParamsWithManagedContext()

			fmt.Println("Executing completion")
			response, err := ai.executeCompletion(ctx, params)
			if err != nil {
				fmt.Printf("Error executing completion: %v\n", err)
				time.Sleep(5 * time.Second)
				continue
			}

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
	return response, nil
}

func (ai *ErrorAI) getParamsWithManagedContext() openai.ChatCompletionNewParams {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(ai.initialSystem),
		openai.UserMessage(ai.initialUser),
	}

	messages = append(messages, ai.truncateConversation()...)

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
						"path": map[string]string{
							"type":        "string",
							"description": "The bash command to execute",
						},
					},
					"required": []string{"path"},
				},
			),
		}),
		Model:       openai.F(openai.ChatModelGPT4oMini),
		Temperature: openai.Float(0.0),
	}
}

func (ai *ErrorAI) truncateConversation() []openai.ChatCompletionMessageParamUnion {
	var truncatedMessages []openai.ChatCompletionMessageParamUnion
	remainingTokens := ai.maxContextTokens - ai.getTotalTokens()

	for i := len(ai.conversationLog) - 1; i >= 0; i-- {
		messageTokens := ai.estimateTokens(ai.conversationLog[i])
		if remainingTokens >= messageTokens {
			truncatedMessages = append([]openai.ChatCompletionMessageParamUnion{ai.conversationLog[i]}, truncatedMessages...)
			remainingTokens -= messageTokens
		} else {
			break
		}
	}

	return truncatedMessages
}

func (ai *ErrorAI) processResponse(response *openai.ChatCompletion) {
	ai.updateTokenCounts(response.Usage)
	fmt.Println("Printing response")
	ai.updateConversationLog(ai.extractAndPrintResponse(response))
	fmt.Println("Handling tool calls")
	ai.updateToolMessages(ai.handleToolCalls(response))
}

func (ai *ErrorAI) extractAndPrintResponse(response *openai.ChatCompletion) openai.ChatCompletionUserMessageParam {
	if response == nil || len(response.Choices) == 0 {
		fmt.Println("No response from OpenAI")
		return openai.UserMessage("Looking for a response...").(openai.ChatCompletionUserMessageParam)
	}

	var reasoning map[string]any
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &reasoning); err != nil {
		fmt.Printf("Error unmarshalling response: %v\n", err)
		return openai.UserMessage(fmt.Sprintf("Error unmarshalling response: %v", err)).(openai.ChatCompletionUserMessageParam)
	}

	ai.done = reasoning["done"].(bool)
	fmt.Println(PrettyJSON(reasoning))
	return openai.UserMessage(response.Choices[0].Message.Content).(openai.ChatCompletionUserMessageParam)
}

func (ai *ErrorAI) handleToolCalls(response *openai.ChatCompletion) openai.ChatCompletionToolMessageParam {
	if response == nil || len(response.Choices) == 0 {
		fmt.Println("No response from OpenAI")
		return openai.ChatCompletionToolMessageParam{}
	}

	message := response.Choices[0].Message
	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		fmt.Println("No tool calls in response")
		return openai.ChatCompletionToolMessageParam{}
	}

	for _, toolCall := range message.ToolCalls {
		fmt.Println(PrettyJSON(toolCall))
		var args map[string]any
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			fmt.Printf("Error unmarshalling tool call arguments: %v\n", err)
			continue
		}

		switch toolCall.Function.Name {
		case "bash_command":
			fmt.Printf("Executing command: %s\n", args["path"])
			buf, err := ai.runner.ExecuteCommand(ai.ctx, []string{args["path"].(string)})
			if err != nil {
				fmt.Printf("Error executing command: %v\n", err)
				continue
			}
			fmt.Println(string(buf))
			return openai.ToolMessage(toolCall.ID, string(buf))
		}
	}

	fmt.Println("No tool calls matched")
	return openai.ChatCompletionToolMessageParam{}
}

func (ai *ErrorAI) getTotalTokens() int {
	total := 0
	for _, msg := range ai.conversationLog {
		total += ai.estimateTokens(msg)
	}
	return total
}

func (ai *ErrorAI) estimateTokens(msg openai.ChatCompletionMessageParamUnion) int {
	return 0
}

func (ai *ErrorAI) updateTokenCounts(usage openai.CompletionUsage) {
	ai.tokenCounts = append(ai.tokenCounts, usage.PromptTokens)
}

func (ai *ErrorAI) updateConversationLog(message openai.ChatCompletionMessageParamUnion) {
	ai.conversationLog = append(ai.conversationLog, message)
}

func (ai *ErrorAI) updateToolMessages(message openai.ChatCompletionToolMessageParam) {
	ai.toolMessages = append(ai.toolMessages, message)
}

func (ai *ErrorAI) cleanup() {
	fmt.Println("Stopping container...")
	if err := ai.runner.StopContainer(context.Background()); err != nil {
		fmt.Printf("Error stopping container: %v\n", err)
	}
}

func (ai *ErrorAI) prepareWorkspace() error {
	fmt.Println("Copying files to /tmp/workspace...")
	if err := os.RemoveAll("/tmp/workspace"); err != nil {
		return err
	}

	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ignoreDirs := []string{"logs", "tmp", "node_modules", ".git", "src-tauri", "youi", "frontend", "ui"}
		if info.IsDir() && slices.Contains(ignoreDirs, info.Name()) {
			fmt.Printf("Skipping %s\n", path)
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(".", path)
		if err != nil {
			return err
		}

		destPath := filepath.Join("/tmp/workspace", relPath)
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

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
		return nil, nil, fmt.Errorf("Error creating builder: %w", err)
	}
	if err := builder.BuildImage(ctx, "./container", imageName); err != nil {
		return nil, nil, fmt.Errorf("Error building image: %w", err)
	}

	in, out, err = ai.runner.RunContainer(ctx, imageName, cmd, username, customMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("Error running container: %w", err)
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
