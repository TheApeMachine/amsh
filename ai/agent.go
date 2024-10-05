package ai

import (
	"context"
	"fmt"
	"io"

	"github.com/google/generative-ai-go/genai"
	openai "github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/errnie"
	"google.golang.org/api/iterator"
)

type Agent struct {
	ctx    context.Context
	conn   *Conn
	ID     string
	Type   string
	system string
	user   string
	Color  string
}

func NewAgent(ctx context.Context, conn *Conn, ID, Type, system, user, color string) *Agent {
	return &Agent{
		ctx:    ctx,
		conn:   conn,
		ID:     ID,
		Type:   Type,
		system: system,
		user:   user,
		Color:  color,
	}
}

func (agent *Agent) Generate(ctx context.Context, user string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		// Retrieve relevant knowledge
		relevantKnowledge, err := agent.RetrieveKnowledge(user)
		if err != nil {
			fmt.Printf("Error retrieving knowledge: %v\n", err)
		}

		// Modify user prompt
		userPrompt := fmt.Sprintf("%s\n\nRelevant Knowledge:\n%s", user, relevantKnowledge)

		agent.NextOpenAI(agent.system, userPrompt, out)
	}()

	return out
}

func (agent *Agent) GenerateEmbedding(text string) ([]float32, error) {
	// Use OpenAI API to generate embeddings
	embedding, err := agent.conn.client.CreateEmbeddings(agent.ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	})
	if err != nil {
		return nil, err
	}
	// Convert embedding to []float32
	var vector []float32
	for _, v := range embedding.Data[0].Embedding {
		vector = append(vector, float32(v))
	}
	return vector, nil
}

func (agent *Agent) StoreKnowledge(content string) error {
	return nil
}

func (agent *Agent) RetrieveKnowledge(prompt string) (string, error) {
	return "", nil
}

/*
NextLocal handles the local LLM interaction.
*/
func (agent *Agent) NextLocal(system, user string, out chan string) {
	request := openai.ChatCompletionRequest{
		Model: "lmstudio-community/Meta-Llama-3.1-8B-Instruct-GGUF",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		Stream: true,
	}

	stream, err := agent.conn.local.CreateChatCompletionStream(agent.ctx, request)
	if err != nil {
		errnie.Error(err)
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			errnie.Error(err)
			break
		}

		if chunk := response.Choices[0].Delta.Content; chunk != "" {
			out <- chunk
		}
	}
}

/*
NextOpenAI handles the OpenAI API interaction.
*/
func (agent *Agent) NextOpenAI(system, user string, out chan string) {
	request := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		Stream: true,
	}

	stream, err := agent.conn.client.CreateChatCompletionStream(agent.ctx, request)
	if err != nil {
		errnie.Error(err)
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			errnie.Error(err)
			break
		}

		if chunk := response.Choices[0].Delta.Content; chunk != "" {
			out <- chunk
		}
	}
}

/*
ChatCompletion generates a single, complete response from the OpenAI API.
*/
func (agent *Agent) ChatCompletion(system, user string) (string, error) {
	request := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	}

	response, err := agent.conn.client.CreateChatCompletion(agent.ctx, request)
	if err != nil {
		errnie.Error(err)
		return "", err
	}

	content := response.Choices[0].Message.Content
	return content, nil
}

func (agent *Agent) NextGemini(system, user string, out chan string) {
	model := agent.conn.gemini.GenerativeModel("gemini-1.5-flash")
	iter := model.GenerateContentStream(agent.ctx, genai.Text(system+"\n\n"+user))

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			errnie.Error(err)
			break
		}

		for _, candidate := range resp.Candidates {
			for _, part := range candidate.Content.Parts {
				if formatted := fmt.Sprintf("%s", part); formatted != "" {
					out <- formatted
				}
			}
		}
	}
}
