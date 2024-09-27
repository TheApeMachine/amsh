package ai

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
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
	prompt *Prompt
	stream *openai.ChatCompletionStream
	step   int
	out    chan string
	err    error
}

/*
NewAgent dynamically constructs an expert Agent designed to perform one or
more tasks to achieve an overall goal.
*/
func NewAgent(ctx context.Context, conn *Conn, prompt *Prompt) *Agent {
	return &Agent{
		ctx:    ctx,
		conn:   conn,
		prompt: prompt,
		step:   0,
		out:    make(chan string),
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
	ctx context.Context, step int,
) <-chan string {
	go func() {
		defer close(agent.out)
		agent.NextGemini()
	}()

	return agent.out
}

func (agent *Agent) NextGemini() {
	model := agent.conn.gemini.GenerativeModel("gemini-1.5-flash")
	iter := model.GenerateContentStream(agent.ctx, genai.Text(agent.prompt.systems[agent.step]+agent.prompt.contexts[agent.step].Responses[agent.step]))

	for {
		var resp *genai.GenerateContentResponse

		if resp, agent.err = iter.Next(); agent.err != nil {
			errnie.Error(agent.err.Error())
			return
		}

		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					agent.out <- fmt.Sprintf("%s", part)
				}
			}
		}
	}
}

func (agent *Agent) NextOpenAI() {
	if agent.stream, agent.err = agent.conn.client.CreateChatCompletionStream(agent.ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: agent.prompt.systems[agent.step],
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: agent.prompt.contexts[agent.step].Responses[agent.step],
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

		agent.out <- response.Choices[0].Delta.Content
	}
}
