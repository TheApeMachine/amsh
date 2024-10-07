package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	errnie.Trace()

	return &Conn{
		client: openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		gemini: NewGeminiConn(),
		local:  NewLocalConn(),
	}
}

/*
Next is the core function that orchestrates the generation of responses from the AI services.
It randomly selects an active service and processes the request through that service.
The selected service is determined by the availability of the service and the presence of a valid API key.
The function returns a channel that emits Chunk objects, each containing a response from the AI service.
*/
func (conn *Conn) Next(ctx context.Context, prompt *Prompt, chunk Chunk) chan Chunk {
	errnie.Trace()

	out := make(chan Chunk, 128)
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

		errnie.Debug("SYSTEM:\n\n")
		errnie.Debug(strings.Join(prompt.System, "\n\n"))
		errnie.Debug("USER:\n\n")
		errnie.Debug(strings.Join(prompt.User, "\n\n"))

		switch selectedService {
		case "openai":
			conn.nextOpenAI(ctx, out, prompt, chunk)
		case "gemini":
			conn.nextGemini(ctx, out, prompt, chunk)
		case "local":
			conn.nextLocal(ctx, out, prompt, chunk)
		}
	}()

	return out
}

/*
nextOpenAI is a helper function that handles the generation of responses using the OpenAI service.
It constructs a ChatCompletionRequest with the provided prompt and sends it to the OpenAI API.
The response is then formatted and sent to the output channel.
*/
func (conn *Conn) nextOpenAI(ctx context.Context, out chan Chunk, prompt *Prompt, chunk Chunk) {
	errnie.Trace()

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
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
		Stream: true,
	}

	stream, err := conn.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()

		if errors.Is(err, io.EOF) {
			return
		}

		if err != nil {
			errnie.Error(err)
			return
		}

		chunk.Response = response.Choices[0].Delta.Content
		out <- chunk
	}
}

/*
nextGemini is a helper function that handles the generation of responses using the Gemini service.
It constructs a GenerativeModel with the specified model name and sends the prompt to the Gemini API.
The response is then formatted and sent to the output channel.
*/
func (conn *Conn) nextGemini(ctx context.Context, out chan Chunk, prompt *Prompt, chunk Chunk) {
	errnie.Trace()

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
					chunk.Response = formatted
					out <- chunk
				}
			}
		}
	}
}

/*
nextLocal is a helper function that handles the generation of responses using a local LLM.
It constructs a ChatCompletionRequest with the provided prompt and sends it to the local LLM API.
The response is then formatted and sent to the output channel.
*/
func (conn *Conn) nextLocal(ctx context.Context, out chan Chunk, prompt *Prompt, chunk Chunk) {
	errnie.Trace()

	var (
		response openai.ChatCompletionResponse
		err      error
	)

	if response, err = conn.local.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
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

	chunk.Response = response.Choices[0].Message.Content
	out <- chunk
}

/*
NewGeminiConn establishes a connection to the Gemini API.
This function is separated to allow for potential future customization of the Gemini client.
*/
func NewGeminiConn() *genai.Client {
	errnie.Trace()

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
	errnie.Trace()

	config := openai.DefaultConfig("lm-studio")
	config.BaseURL = "http://localhost:1234/v1"

	return nil
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
	errnie.Trace()

	c.client = client
	return c
}
