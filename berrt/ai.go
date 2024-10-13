package berrt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/TylerBrock/colorjson"
	"github.com/charmbracelet/lipgloss"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
)

var Dark = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#666666")).Render
var Muted = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#999999")).Render
var Highlight = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#EEEEEE")).Render
var Blue = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#6E95F7")).Render
var Red = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#F7746D")).Render
var Yellow = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#F7B96D")).Render
var Green = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#06C26F")).Render

var done = false

func PrettyJSON(v any) string {
	f := colorjson.NewFormatter()
	f.Indent = 2
	s, _ := f.Marshal(v)
	return string(s)
}

type ErrorAI struct {
	ctx    context.Context
	System string
	User   string
}

func NewErrorAI() *ErrorAI {
	return &ErrorAI{
		ctx: context.Background(),
		System: `
		You are a helpful assistant that helps me debug my code.
		You are given an error message, a stacktrace, and a snippet of code, and you have access to a few tools to help you debug the error.
		When using tools, you should be careful with using them in ways that produce a lot of output, as it may overflow the message context window.
		You should always start with highly specific queries and commands, and not search for single, generic words.
		Always think step-by-step before taking an action.
		`,
		User: "I am getting the following error:\n\n",
	}
}

func GenerateSchema[T any]() interface{} {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

func GetParams(
	system, user string,
	schema openai.ResponseFormatJSONSchemaJSONSchemaParam,
	toolset []openai.ChatCompletionToolParam,
) openai.ChatCompletionNewParams {
	return openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
			openai.UserMessage(user),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schema),
			},
		),
		Tools:       openai.F(toolset),
		Seed:        openai.Int(0),
		Model:       openai.F(openai.ChatModelGPT4oMini),
		Temperature: openai.Float(0.0),
	}
}

/*
makeTool reduces some of the boilerplate code for creating a tool.
*/
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

func (completion *Completion) Execute(ctx context.Context, params openai.ChatCompletionNewParams) *openai.ChatCompletion {
	response, err := completion.client.Chat.Completions.New(ctx, params)
	if err != nil {
		_ = fmt.Errorf("%w", err)
		return nil
	}

	if response == nil {
		err = errors.New("no response from OpenAI")
		_ = fmt.Errorf("%w", err)
		return nil
	}

	return response
}

func (ai *ErrorAI) Execute(message, stacktrace, snippet string) {
	params := GetParams(
		ai.System, strings.Join([]string{ai.User, message, stacktrace, snippet}, "\n\n"),
		openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        openai.F("error_analysis"),
			Description: openai.F("Analyze the error and provide a plan to resolve it"),
			Schema:      openai.F(GenerateSchema[ErrorAnalysis]()),
			Strict:      openai.Bool(true),
		},
		[]openai.ChatCompletionToolParam{
			makeTool(
				"tree",
				"Get a tree of the current project directory to inspect the file structure.",
				openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]string{
							"type":        "string",
							"description": "The path to the directory to get the tree of",
						},
					},
					"required": []string{"path"},
				},
			),
			makeTool(
				"get_file_contents",
				"Get the contents of a file to inspect the code.",
				openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]string{
							"type":        "string",
							"description": "The path to the file to get the contents of",
						},
					},
					"required": []string{"path"},
				},
			),
			makeTool(
				"ack",
				"Search the codebase using ack.",
				openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"ack_arguments": map[string]string{
							"type":        "string",
							"description": "The ack arguments to execute",
						},
					},
					"required": []string{"ack_arguments"},
				},
			),
			makeTool(
				"sed",
				"Replace code with your fixes using sed.",
				openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"sed_arguments": map[string]string{
							"type":        "string",
							"description": "The sed arguments to execute",
						},
					},
					"required": []string{"sed_arguments"},
				},
			),
			makeTool(
				"git",
				"Use git to make sure your changes are isolated on a new branch and don't affect other code.",
				openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"git_arguments": map[string]string{
							"type":        "string",
							"description": "The git arguments to execute",
						},
					},
					"required": []string{"git_arguments"},
				},
			),
		},
	)

	for {
		if done {
			os.Exit(0)
		}

		response := NewCompletion(ai.ctx).Execute(ai.ctx, params)
		completionMessage := ai.printResponse(response)
		params.Messages.Value = append(
			params.Messages.Value, openai.UserMessage(ai.handleToolCalls(&completionMessage, params)),
		)
	}
}

func (ai *ErrorAI) printResponse(response *openai.ChatCompletion) openai.ChatCompletionMessage {
	if response == nil || len(response.Choices) == 0 {
		return openai.ChatCompletionMessage{}
	}

	var reasoning map[string]any
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &reasoning); err != nil {
		_ = fmt.Errorf("%w", err)
		return response.Choices[0].Message
	}

	done = reasoning["done"].(bool)
	fmt.Println(PrettyJSON(reasoning))
	return response.Choices[0].Message
}

func (ai *ErrorAI) handleToolCalls(
	message *openai.ChatCompletionMessage, params openai.ChatCompletionNewParams,
) string {
	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		return message.Content
	}

	var (
		args map[string]interface{}
		out  string
	)

	params.Messages.Value = append(params.Messages.Value, message)
	for _, toolCall := range message.ToolCalls {
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); fmt.Errorf("%w", err) != nil {
			out = "error unmarshalling arguments"
		}

		switch toolCall.Function.Name {
		case "tree":
			fmt.Println("$ tree", args["path"])
			// Call the shell command to get the tree
			tree, err := exec.Command("/usr/bin/tree", args["path"].(string)).Output()
			if err != nil {
				out = fmt.Errorf("%w", err).Error()
				params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
				return out
			}

			fmt.Println(string(tree))
			out = string(tree)
			params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
		case "get_file_contents":
			fmt.Println("$ cat", args["path"])
			// Call the shell command to get the file contents
			contents, err := os.ReadFile(args["path"].(string))
			if err != nil {
				out = fmt.Errorf("%w", err).Error()
				params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
				return out
			}

			fmt.Println(string(contents))
			out = string(contents)
			params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
		case "ack":
			fmt.Println("$ ack", args["ack_arguments"])
			// Call the shell command to search the codebase
			ack, err := exec.Command("ack", args["ack_arguments"].(string), "--ignore-dir={logs,.git,node_modules}").Output()
			if err != nil {
				out = fmt.Errorf("%w", err).Error()
				params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
				return out
			}

			fmt.Println(string(ack))
			out = string(ack)
			params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
		case "sed":
			// Call the shell command to replace the code
			sed, err := exec.Command("sed", args["sed_arguments"].(string)).Output()
			if err != nil {
				out = fmt.Errorf("%w", err).Error()
				params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
				return out
			}

			fmt.Println(string(sed))
			out = string(sed)
			params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
		case "git":
			fmt.Println("$ git", args["git_arguments"])

			// Call the shell command to run git
			git, err := exec.Command("git", args["git_arguments"].(string)).Output()
			if err != nil {
				out = fmt.Errorf("%w", err).Error()
				params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
				return out
			}

			fmt.Println(string(git))
			out = string(git)
			params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
		}
	}

	return out
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

func checkWithHuman(command string) (string, bool) {
	fmt.Printf("Executing: %s\n", command)
	fmt.Println("Continue? (y/n)")
	var response string
	fmt.Scanln(&response)
	if response != "y" {
		// Get user input to send to the AI
		fmt.Println("Feedback:")
		var feedback string
		fmt.Scanln(&feedback)
		command = fmt.Sprintf("%s %s", command, feedback)
		return command, false
	}

	return command, true
}
