package ai

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/memory"
	"github.com/theapemachine/amsh/errnie"
)

type Chunk struct {
	Iteration int    `json:"iteration"`
	Prompt    string `json:"prompt"`
	Agent     string `json:"agent"`
	AgentType string `json:"agent_type"`
	System    string `json:"system"`
	User      string `json:"user"`
	Response  string `json:"response"`
	Color     string `json:"color"`
}

type Pipeline struct {
	ctx              context.Context
	conn             *Conn
	agents           map[string]*Agent
	resourceManager  *ResourceManager
	history          string
	approachHistory  []string
	currentIteration int
	maxIterations    int
	agentResponses   map[string]string
	semanticMemory   *memory.SemanticMemory
}

func NewPipeline(ctx context.Context, conn *Conn) *Pipeline {
	return &Pipeline{
		ctx:              ctx,
		conn:             conn,
		agents:           make(map[string]*Agent),
		resourceManager:  NewResourceManager(),
		approachHistory:  make([]string, 0),
		currentIteration: 0,
	}
}

func (pipeline *Pipeline) Initialize() *Pipeline {
	errnie.Trace()

	agentTypes := []string{
		"reasoner",
		"verifier",
		"learning",
		"metacognition",
		"prompt_engineer",
		"context_manager",
	}

	for idx, agentType := range agentTypes {
		ID := namegenerator.NewNameGenerator(time.Now().UnixNano()).Generate()
		system := viper.GetString(fmt.Sprintf("ai.%s.system", agentType))
		user := viper.GetString(fmt.Sprintf("ai.%s.user", agentType))

		if system == "" || user == "" {
			errnie.Error(fmt.Errorf("missing configuration for %s agent", agentType))
			continue
		}

		// In Pipeline.Initialize()
		pipeline.semanticMemory = memory.NewSemanticMemory()
		err := pipeline.semanticMemory.CreateCollection(context.Background(), "knowledge_base", 1536) // Assuming 1536-dimensional embeddings
		if err != nil {
			log.Fatalf("Failed to create collection: %v", err)
		}

		pipeline.agents[agentType] = NewAgent(
			pipeline.ctx,
			pipeline.conn,
			ID,
			agentType,
			system,
			user,
			Colors[idx%len(Colors)],
		)
	}

	return pipeline
}

func (pipeline *Pipeline) Generate(prompt string, iterations int) <-chan Chunk {
	errnie.Trace()
	pipeline.maxIterations = iterations

	out := make(chan Chunk)

	go func() {
		defer close(out)

		pipelineOrder := viper.GetStringSlice("pipeline")
		finalAnswerReached := false

		for pipeline.currentIteration < iterations && !finalAnswerReached {
			pipeline.currentIteration++
			chunk := Chunk{}
			chunk.Prompt = prompt
			chunk.Iteration = pipeline.currentIteration
			chunk.Response = ""
			iterationResources := make([]string, 0)

			for _, agentType := range pipelineOrder {
				agent, exists := pipeline.agents[agentType]
				if !exists {
					errnie.Error(fmt.Errorf("agent %s not found", agentType))
					continue
				}

				chunk.Agent = agent.ID
				chunk.AgentType = agent.Type

				agentResponse := pipeline.runAgent(agent, chunk, out)
				if strings.Contains(strings.ToUpper(agentResponse), "FINAL ANSWER: TRUE") {
					finalAnswerReached = true
					break
				}

				// We don't need to run any more agents if we've reached the max iterations.
				if pipeline.currentIteration == iterations {
					break
				}

				// Extract [RESEARCH] tag without requiring quotes
				re := regexp.MustCompile(`(?m)\[RESEARCH\]\s*(.+)$`)
				matches := re.FindAllStringSubmatch(agentResponse, -1)
				for _, match := range matches {
					if len(match) > 1 {
						searchQuery := strings.TrimSpace(match[1])
						iterationResources = append(iterationResources, fmt.Sprintf(`[RESEARCH] %s`, searchQuery))
					}
				}

				// Extract [PROGRAM] tag
				if strings.Contains(agentResponse, "[PROGRAM]") {
					iterationResources = append(iterationResources, "[PROGRAM]")
				}

				if agent.Type == "context_manager" {
					// Extract markdown content
					re := regexp.MustCompile("(?s)```markdown(.*?)```")
					match := re.FindStringSubmatch(agentResponse)
					if len(match) > 0 {
						pipeline.history = match[1]
					}

					// Process all collected resources after context_manager runs
					for _, resource := range iterationResources {
						pipeline.useResource(resource)
					}
				}
			}
		}
	}()

	return out
}

func (pipeline *Pipeline) useResource(resource string) {
	errnie.Trace()
	fmt.Printf("useResource called with: %s\n", resource)

	// Regular expression to match [RESEARCH] tags without requiring quotes
	re := regexp.MustCompile(`(?m)\[RESEARCH\]\s*(.+)$`)
	matches := re.FindAllStringSubmatch(resource, -1)

	var outputs []string
	for _, match := range matches {
		if len(match) > 1 {
			query := strings.TrimSpace(match[1])
			fmt.Printf("Processing research query: %s\n", query)
			if query != "" {
				output := pipeline.resourceManager.UseResource("RESEARCH", query)
				outputs = append(outputs, output)
			}
		}
	}

	// Handle [PROGRAM] tag if present
	if strings.Contains(resource, "[PROGRAM]") {
		output := pipeline.resourceManager.UseResource("PROGRAM", "")
		outputs = append(outputs, output)
	}

	// Combine all outputs and add to history
	combinedOutput := strings.Join(outputs, "\n\n")
	pipeline.history += combinedOutput + "\n"
}

func (pipeline *Pipeline) runAgent(agent *Agent, chunk Chunk, out chan<- Chunk) string {
	var context string
	if agent.Type == "prompt_engineer" {
		// Collect original prompts and feedback
		originalPrompts := pipeline.getOriginalPrompts()
		agentFeedback := pipeline.getAgentFeedback()

		// Replace placeholders in the agent's user prompt
		context = strings.ReplaceAll(agent.user, "{original_prompts}", originalPrompts)
		context = strings.ReplaceAll(context, "{agent_feedback}", agentFeedback)
	} else {
		context := strings.ReplaceAll(agent.user, "{prompt}", chunk.Prompt)
		context = strings.ReplaceAll(context, "{context}", pipeline.history)
		context = strings.ReplaceAll(context, "{resources}", "No additional resources available.")
	}

	chunk.System = agent.system
	chunk.User = context

	if agent.Type == "prompt_engineer" {
		// Collect the response and update agent prompts
		fullResponse := ""
		for chnk := range agent.Generate(pipeline.ctx, context) {
			chunk.Response = chnk
			out <- chunk
			pipeline.history += fmt.Sprintf("%s\n", chunk.Response)
			fullResponse += chnk
		}

		// After receiving the full response, parse and update prompts
		pipeline.updateAgentPrompts(fullResponse)
	}

	agentResponse := fmt.Sprintf(
		"\n\nDuring iteration %d of %d, agent %s (a %s) responded with:\n\n",
		pipeline.currentIteration,
		pipeline.maxIterations,
		agent.ID,
		agent.Type,
	)

	_ = agentResponse

	chunk.Color = reset + agent.Color

	fullResponse := ""
	for chnk := range agent.Generate(pipeline.ctx, context) {
		chunk.Response = chnk
		out <- chunk
		pipeline.history += fmt.Sprintf("%s\n", chunk.Response)
		fullResponse += chnk
	}

	// Store the agent's response for feedback if applicable
	if agent.Type == "verifier" || agent.Type == "learning" || agent.Type == "metacognition" {
		pipeline.agentResponses[agent.Type] = fullResponse
	}

	// Existing code...
	return fullResponse
}

func (pipeline *Pipeline) getOriginalPrompts() string {
	var builder strings.Builder
	for _, agent := range pipeline.agents {
		if agent.Type != "prompt_engineer" {
			builder.WriteString(fmt.Sprintf("### %s\n", agent.Type))
			builder.WriteString("**System Prompt:**\n")
			builder.WriteString("```\n" + agent.system + "\n```\n")
			builder.WriteString("**User Prompt:**\n")
			builder.WriteString("```\n" + agent.user + "\n```\n")
		}
	}
	return builder.String()
}

func (pipeline *Pipeline) getAgentFeedback() string {
	var builder strings.Builder
	feedbackAgents := []string{"verifier", "learning", "metacognition"}
	for _, agentType := range feedbackAgents {
		if feedback, exists := pipeline.agentResponses[agentType]; exists {
			builder.WriteString(fmt.Sprintf("### %s Feedback\n", strings.Title(agentType)))
			builder.WriteString(feedback + "\n\n")
		}
	}
	return builder.String()
}

func (pipeline *Pipeline) updateAgentPrompts(response string) {
	optimizedPrompts := parseOptimizedPrompts(response)

	for agentName, prompts := range optimizedPrompts {
		if agent, exists := pipeline.agents[agentName]; exists {
			agent.system = prompts["system"]
			agent.user = prompts["user"]
		}
	}
}

func parseOptimizedPrompts(response string) map[string]map[string]string {
	optimizedPrompts := make(map[string]map[string]string)

	// Use regular expressions to parse the response
	re := regexp.MustCompile("\\*\\*([^*]+)\\*\\*:\\s+- \\*\\*System Prompt\\*\\*:\\s+```yaml\\s+(.+?)\\s+```\\s+- \\*\\*User Prompt\\*\\*:\\s+```yaml\\s+(.+?)\\s+```(?s)")
	matches := re.FindAllStringSubmatch(response, -1)

	for _, match := range matches {
		agentName := match[1]
		systemPrompt := match[2]
		userPrompt := match[3]
		optimizedPrompts[agentName] = map[string]string{
			"system": systemPrompt,
			"user":   userPrompt,
		}
	}

	return optimizedPrompts
}
