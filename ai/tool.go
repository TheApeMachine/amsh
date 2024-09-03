package ai

import (
	"io"

	"github.com/sashabaranov/go-openai"
)

/*
Tool is an interface objects can implement if they want to provide additional
capabilities to AI Agents. This allows the Agent to do things like use a
headless browser to use the Internet, make HTTP requests to interface with
external APIs or services, or relay information to amsh internals.
*/
type Tool interface {
	io.ReadWriteCloser
	Definition() openai.Tool
}
