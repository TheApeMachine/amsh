package ai

import (
	"context"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/errnie"
)

/*
Agent is a configurable wrapper around an AI model, which uses composable
prompt templates to induce a specific type of behavior to performodel.
Multiple Agents can be constructed and they should be able to communicate
and coordinate to form a cohesive AI Team of experts.
*/
type Agent struct {
	ctx    context.Context
	conn   *Conn
	name   string
	role   string
	prompt *Prompt
	stream *openai.ChatCompletionStream
	err    error
}

/*
NewAgent dynamically constructs an expert Agent designed to perform one or
more tasks to achieve an overall goal.
*/
func NewAgent(ctx context.Context, conn *Conn, name string, role string, prompt *Prompt) *Agent {
	errnie.Debug("Creating agent for %s with name %s", role, name)
	return &Agent{
		ctx:    ctx,
		conn:   conn,
		name:   name,
		role:   role,
		prompt: prompt,
	}
}

/*
CreateChatCompletion sends a request to the OpenAI API for a chat completion.
This method is crucial for enabling the Agent to interact with the AI model,
allowing it to generate responses based on its role and available tools.
By including the agent's tools in the request, we enable the AI to utilize
these tools when formulating its response, enhancing its capabilities.
*/
func (agent *Agent) Generate(
	ctx context.Context, context string,
) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		system := agent.replacements(agent.makeSystem())
		user := agent.prompt.context

		if agent.stream, agent.err = agent.conn.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: system,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: user,
				},
			},
		}); agent.err != nil {
			errnie.Error(agent.err.Error())
			return
		}

		var response openai.ChatCompletionStreamResponse

		for {
			if response, agent.err = agent.stream.Recv(); agent.err != nil {
				errnie.Error(agent.err.Error())
				return
			}

			if response.Choices[0].Delta.Content == "" {
				continue
			}

			out <- response.Choices[0].Delta.Content
		}
	}()

	return out
}

func (agent *Agent) makeSystem() string {
	return strings.Join([]string{
		agent.prompt.system,
	}, "\n\n")
}

func (agent *Agent) replacements(prompt string) string {
	prompt = strings.ReplaceAll(prompt, "<{profile}>", agent.prompt.role)
	prompt = strings.ReplaceAll(prompt, "<{name}>", "`"+agent.name+"`")
	prompt = strings.ReplaceAll(prompt, "<{modules}>", agent.prompt.modules)

	return prompt
}
