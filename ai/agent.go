package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

/*
Agent is a configurable wrapper around an AI model, which uses composable
prompt templates to induce a specific type of behavior to perform.
Multiple Agents can be constructed and they should be able to communicate
and coordinate to form a cohesive AI Team of experts.
*/
type Agent struct {
	conn  *Conn
	role  RoleType
	tools []Tool
	name  string
}

/*
NewAgent dynamically constructs an expert Agent designed to perform one or
more tasks to achieve an overall goal.
*/
func NewAgent(conn *Conn, role RoleType, tools []Tool, name string) *Agent {
	return &Agent{
		conn:  conn,
		role:  role,
		tools: tools,
		name:  name,
	}
}

/* 
CreateChatCompletion sends a request to the OpenAI API for a chat completion.
This method is crucial for enabling the Agent to interact with the AI model,
allowing it to generate responses based on its role and available tools.
By including the agent's tools in the request, we enable the AI to utilize
these tools when formulating its response, enhancing its capabilities.
*/
func (a *Agent) CreateChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage) (openai.ChatCompletionResponse, error) {
	return a.conn.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT4TurboPreview,
		Messages: messages,
		Tools:    a.getToolDefinitions(),
	})
}

/*
getToolDefinitions prepares the tool definitions for the OpenAI API request.
This method is important because it allows the Agent to inform the AI model
about the tools it has access to. By doing this, we enable the AI to make
informed decisions about when and how to use these tools in its responses,
effectively extending the AI's capabilities beyond simple text generation.
*/
func (a *Agent) getToolDefinitions() []openai.Tool {
	var tools []openai.Tool
	for _, tool := range a.tools {
		tools = append(tools, tool.Definition())
	}
	return tools
}

/* 
ExecuteTask performs a task and returns a structured output.
This method is designed to leverage the capabilities of the AI model to execute a given task.
It constructs a prompt using various templates and instructions, ensuring that the AI has all the 
necessary context and guidelines to perform the task effectively. By including role-specific templates, 
task content, and instructions, the method ensures that the AI's response is structured and relevant to the task at hand.
The method then sends this prompt to the OpenAI API and processes the response, attempting to parse it into a structured format.
If parsing fails, it returns the raw response, ensuring that the caller always receives some form of result.
This approach allows the Agent to handle a wide range of tasks dynamically, making it a versatile tool for various applications.
*/
func (a *Agent) ExecuteTask(ctx context.Context, task string) (map[string]interface{}, error) {
	prompt := NewPrompt().
		AddRoleTemplate(a.role).
		AddContent("task", task).
		AddInstructions().
		AddContent("task_template", task). // Use AddContent instead of AddTaskTemplate
		Build()

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: prompt},
		{Role: openai.ChatMessageRoleUser, Content: task},
	}

	response, err := a.CreateChatCompletion(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to execute task: %w", err)
	}

	fmt.Printf("Raw AI Response for %s:\n%s\n", a.name, response.Choices[0].Message.Content)

	result, err := parseStructuredOutput(response.Choices[0].Message.Content)
	if err != nil {
		// If parsing fails, return the raw response as the result
		return map[string]interface{}{
			"result": response.Choices[0].Message.Content,
		}, nil
	}

	// Ensure the result always has a "result" key
	if _, ok := result["result"]; !ok {
		result["result"] = result
	}

	return result, nil
}

/*
parseStructuredOutput attempts to parse the raw output from the AI model into a structured format.
This method is crucial for extracting meaningful data from the AI's responses, making it easier to
process and utilize in subsequent steps of the workflow. By parsing the output as JSON, we can
reliably access and manipulate the structured data, ensuring that the AI's response is properly
understood and utilized.
*/
func parseStructuredOutput(output string) (map[string]interface{}, error) {
	// Trim any leading or trailing whitespace
	output = strings.TrimSpace(output)

	// If the output doesn't start with '{', try to find the start of the JSON
	if !strings.HasPrefix(output, "{") {
		jsonStart := strings.Index(output, "{")
		if jsonStart != -1 {
			output = output[jsonStart:]
		}
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w\nOutput: %s", err, output)
	}
	return result, nil
}
