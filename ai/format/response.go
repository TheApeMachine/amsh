package format

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

var dark = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#666666")).Render
var muted = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#999999")).Render
var highlight = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#EEEEEE")).Render
var blue = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#6E95F7")).Render
var red = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#F7746D")).Render
var yellow = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#F7B96D")).Render
var green = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#06C26F")).Render

type Strategy interface {
	FinalAnswer() string
	Schema() (*jsonschema.Definition, error)
	ToString() string
}

type Response struct {
	Name     string
	Strategy Strategy
}

func NewResponse(name string, strategy Strategy) *Response {
	return &Response{
		Name:     name,
		Strategy: strategy,
	}
}

func (response *Response) FinalAnswer() string {
	return response.Strategy.FinalAnswer()
}

func (response *Response) Schema() *jsonschema.Definition {
	var (
		definition *jsonschema.Definition
		err        error
	)

	if definition, err = response.Strategy.Schema(); err != nil {
		errnie.Error(err)
		return nil
	}

	return definition
}

func (response *Response) ToString() string {
	return response.Strategy.ToString()
}
