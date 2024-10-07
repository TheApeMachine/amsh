package ai

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/errnie"
	"golang.org/x/exp/rand"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

/*
Conn encapsulates connections to various AI services.
This struct allows for a unified interface to interact with different AI providers,
enabling easy switching between services or using multiple services concurrently.
*/
type Conn struct {
	client *openai.Client
	gemini *genai.Client
	local  *openai.Client
}

/*
NewConn initializes a new Conn with OpenAI and Gemini clients.
This function assumes that the necessary API keys are set in the environment variables.
Example:

	conn := NewConn()
*/
func NewConn() *Conn {
	return &Conn{
		client: openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		gemini: NewGeminiConn(),
		local:  NewLocalConn(),
	}
}

func (conn *Conn) Next(ctx context.Context, prompt *Prompt) chan string {
	out := make(chan string)
	active := make([]string, 0)

	if conn.client != nil {
		active = append(active, "openai")
	}

	if conn.gemini != nil {
		active = append(active, "gemini")
	}

	if conn.local != nil {
		active = append(active, "local")
	}

	go func() {
		defer close(out)

		// Generate a random integer between 0 and the amount of active services.
		rand.Seed(uint64(time.Now().UnixNano()))
		randomIndex := rand.Intn(len(active))
		selectedService := active[randomIndex]

		switch selectedService {
		case "openai":
			conn.nextOpenAI(ctx, out, prompt)
		case "gemini":
			conn.nextGemini(ctx, out, prompt)
		case "local":
			conn.nextLocal(ctx, out, prompt)
		}
	}()

	return out
}

func (conn *Conn) nextOpenAI(ctx context.Context, out chan string, prompt *Prompt) {
	var (
		response openai.ChatCompletionResponse
		err      error
	)

	if response, err = conn.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: strings.Join(prompt.System, "\n\n"),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: strings.Join(prompt.User, "\n\n"),
			},
		},
	}); errnie.Error(err) != nil {
		return
	}

	out <- response.Choices[0].Message.Content
}

func (conn *Conn) nextGemini(ctx context.Context, out chan string, prompt *Prompt) {
	model := conn.gemini.GenerativeModel("gemini-1.5-flash")
	iter := model.GenerateContentStream(ctx, genai.Text(
		strings.Join([]string{
			strings.Join(prompt.System, "\n\n"),
			strings.Join(prompt.User, "\n\n"),
		}, "\n\n"),
	))

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

func (conn *Conn) nextLocal(ctx context.Context, out chan string, prompt *Prompt) {
	conn.local.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: strings.Join(prompt.System, "\n\n"),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: strings.Join(prompt.User, "\n\n"),
			},
		},
	})
}

/*
NewGeminiConn establishes a connection to the Gemini API.
This function is separated to allow for potential future customization of the Gemini client.
*/
func NewGeminiConn() *genai.Client {
	ctx := context.Background()
	var (
		err    error
		gemini *genai.Client
	)

	if gemini, err = genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY"))); err != nil {
		// Fatal log is used here as the application cannot function without a valid Gemini client
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	return gemini
}

/*
NewLocalConn sets up a connection to a local LLM.
This function allows for testing or development with a local language model,
providing flexibility in the AI backend used.
Example:

	localConn := NewLocalConn()
*/
func NewLocalConn() *openai.Client {
	config := openai.DefaultConfig("lm-studio")
	config.BaseURL = "http://localhost:1234/v1"

	return openai.NewClientWithConfig(config)
}

/*
WithClient allows for custom OpenAI client configuration.
This method enables runtime modification of the OpenAI client,
useful for testing or dynamically changing API endpoints.
Example:

	customClient := openai.NewClient("custom-api-key")
	conn.WithClient(customClient)
*/
func (c *Conn) WithClient(client *openai.Client) *Conn {
	c.client = client
	return c
}
