package ai

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/types"
)

// AgentState represents the current state of an agent
type AgentState string

const (
	StateIdle      AgentState = "idle"
	StateThinking  AgentState = "thinking"
	StateWorking   AgentState = "working"
	StateWaiting   AgentState = "waiting"
	StateReviewing AgentState = "reviewing"
	StateDone      AgentState = "done"
)

// Agent represents an AI agent that can perform tasks and communicate with other agents
type Agent struct {
	role         string
	systemPrompt string
	tools        map[string]types.Tool
	provider     provider.Provider
	buffer       *Buffer
	state        AgentState
}

// NewAgent creates a new agent with integrated reasoning and learning
func NewAgent(role, systemPrompt string, tools map[string]types.Tool) *Agent {
	var toolNames []string

	for key := range tools {
		toolNames = append(toolNames, key)
	}

	if role != "orchestrator" {
		log.Info("Creating new agent", "role", role, "systemPrompt", systemPrompt, "tools", strings.Join(toolNames, ", "))
	}

	return &Agent{
		role:         role,
		systemPrompt: systemPrompt,
		tools:        tools,
		buffer:       NewBuffer().AddMessage("system", systemPrompt),
		provider: provider.NewRandomProvider(map[string]string{
			"openai":    os.Getenv("OPENAI_API_KEY"),
			"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
			"google":    os.Getenv("GEMINI_API_KEY"),
			"cohere":    os.Getenv("COHERE_API_KEY"),
		}),
		state: StateIdle,
	}
}

func (agent *Agent) Execute(prompt string) <-chan provider.Event {
	if agent.state != StateIdle {
		log.Error("Agent is not idle", "role", agent.role)
		return nil
	}

	out := make(chan provider.Event)
	ctx := context.Background()

	composedPrompt := []string{}

	if agent.role != "orchestrator" {
		composedPrompt = append(composedPrompt, "You have the following tools available to you:")

		for key, value := range agent.tools {
			_ = key
			composedPrompt = append(composedPrompt, value.Description())
		}

		composedPrompt = append(composedPrompt, "You will be able to iterate on your response until you are finished. When you are finished, just say 'task complete'.")
	}

	composedPrompt = append(composedPrompt, prompt)
	agent.buffer.AddMessage("user", strings.Join(composedPrompt, "\n\n"))
	log.Info("Executing agent", "role", agent.role, "prompt", strings.Join(composedPrompt, "\n\n"))

	go func() {
		defer close(out)
		agent.state = StateWorking
		agent.buffer.AddMessage("user", prompt)
		iteration := 0

		for agent.state == StateWorking {
			iteration++
			agent.buffer.AddMessage("assistant", "[CURRENT ITERATION: "+strconv.Itoa(iteration)+"]\n\nYour previous work is shown above.\n\n")
			// Send the agent's messages to the provider to generate new output
			var accumulator string
			for event := range agent.provider.Generate(ctx, agent.buffer.GetMessages()) {
				accumulator += event.Content
				out <- event
			}

			if agent.role == "orchestrator" {
				agent.state = StateDone
				break
			}

			if isTaskComplete(accumulator) {
				log.Info("Task complete", "role", agent.role)
				agent.state = StateDone
			}

			agent.buffer.AddMessage("assistant", accumulator)
			// Parse and check for tool call instructions
			if shouldCallTool(accumulator) {
				toolResult, err := agent.callTool(ctx, accumulator)
				if err != nil {
					log.Error("Tool call failed", "error", err)
					out <- provider.Event{
						Type:    provider.EventError,
						Content: err.Error(),
					}
				} else {
					// Log and buffer the tool's output as part of the agent's reasoning
					agent.buffer.AddMessage("tool_output", toolResult)
					out <- provider.Event{Content: toolResult}
				}
			}
		}
	}()

	return out
}

// shouldCallTool checks if a specific tool call is indicated in the agent's output
func shouldCallTool(content string) bool {
	// Example logic to detect tool call instruction in output
	return strings.Contains(content, `"tool_call"`)
}

// isTaskComplete checks if the agent believes the task is complete
func isTaskComplete(content string) bool {
	// Example logic to determine if the task is complete
	return strings.Contains(strings.ToLower(content), `"task complete"`)
}

// callTool executes the tool based on the agent's parsed instruction from its output
func (agent *Agent) callTool(ctx context.Context, content string) (string, error) {
	// Extract any items between Markdown json blocks
	toolCalls := extractToolCalls(content)

	var results []string

	for _, toolCall := range toolCalls {
		// Parse tool name and arguments from content
		toolName, args, err := parseToolCall(toolCall)
		if err != nil {
			log.Error("Failed to parse tool call", "error", err)
			continue
		}

		// Fetch the tool from the agent's available tools
		tool, ok := agent.tools[toolName]
		if !ok {
			log.Error("Tool not found", "tool", toolName)
			continue
		}

		// Execute the tool and return the result
		toolResult, err := tool.Execute(ctx, args)
		if err != nil {
			log.Error("Failed to execute tool", "tool", toolName, "error", err)
			continue
		}

		results = append(results, toolName+" results:\n\n"+toolResult)
	}

	return strings.Join(results, "\n\n"), nil
}

// extractToolCalls extracts any items between Markdown json blocks
func extractToolCalls(content string) []string {
	// Use regex to find all json blocks in the content
	re := regexp.MustCompile("```json(.*?)```")
	return re.FindAllString(content, -1)
}

// parseToolCall parses the tool call details from the agent's output content
func parseToolCall(content string) (toolName string, args map[string]interface{}, err error) {
	// Standardized structure for tool calls
	var toolCall struct {
		Tool      string                 `json:"tool"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	// Attempt to unmarshal the content directly into the standard structure
	if err := json.Unmarshal([]byte(content), &toolCall); err != nil {
		return "", nil, errors.New("failed to parse tool call: invalid JSON structure")
	}

	return toolCall.Tool, toolCall.Arguments, nil
}

func (agent *Agent) GetState() AgentState {
	return agent.state
}
