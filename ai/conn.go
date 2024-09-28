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
Conn is a wrapper around the OpenAI API connection.
It facilitates the interface between the local behavior and state, and the AI models
provided by OpenAI.
*/
type Conn struct {
	client *openai.Client
	gemini *genai.Client
}

/*
NewConn sets up a connection to the OpenAI API.
*/
func NewConn() *Conn {
	return &Conn{
		client: openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		gemini: NewGeminiConn(),
	}
}

/*
NewGeminiConn sets up a connection to the Gemini API.
*/
func NewGeminiConn() *genai.Client {
	ctx := context.Background()
	var (
		err    error
		gemini *genai.Client
	)

	if gemini, err = genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY"))); err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	return gemini
}

/*
NewLocalConn sets up a connection to a local LLM.
*/
func NewLocalConn() *Conn {
	config := openai.DefaultConfig("lm-studio")
	config.BaseURL = "http://localhost:1234/v1"

	return &Conn{
		client: openai.NewClientWithConfig(config),
	}
}

/*
WithClient sets the OpenAI client for the connection.
*/
func (c *Conn) WithClient(client *openai.Client) *Conn {
	c.client = client
	return c
}
