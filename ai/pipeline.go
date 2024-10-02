package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/spf13/viper"
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
	context := strings.ReplaceAll(agent.user, "{prompt}", chunk.Prompt)
	context = strings.ReplaceAll(context, "{context}", pipeline.history)
	context = strings.ReplaceAll(context, "{resources}", "No additional resources available.")

	chunk.System = agent.system
	chunk.User = context

	agentResponse := fmt.Sprintf(
		"\n\nDuring iteration %d of %d, agent %s (a %s) responded with:\n\n",
		pipeline.currentIteration,
		pipeline.maxIterations,
		agent.ID,
		agent.Type,
	)

	chunk.Color = reset + agent.Color

	fullResponse := ""
	for chnk := range agent.Generate(pipeline.ctx, context) {
		chunk.Response = chnk
		out <- chunk
		pipeline.history += fmt.Sprintf("%s\n", chunk.Response)
		fullResponse += chnk
	}

	agentResponse += fullResponse

	return agentResponse
}
