package ai

import (
	"os"

	"github.com/sashabaranov/go-openai"
)

/*
Conn is a wrapper around the OpenAI API connection.
It facilitates the interface between the local behavior and state, and the AI models
provided by OpenAI.
*/
type Conn struct {
	client *openai.Client
}

/*
NewConn sets up a connection to the OpenAI API.
*/
func NewConn() *Conn {
	return &Conn{
		client: openai.NewClient(os.Getenv("OPENAI_API_KEY")),
	}
}

/*
WithClient sets the OpenAI client for the connection.
*/
func (c *Conn) WithClient(client *openai.Client) *Conn {
	c.client = client
	return c
}