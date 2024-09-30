package ai

import (
	"context"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
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
