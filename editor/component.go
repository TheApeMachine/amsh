package editor

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type Component interface {
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	Focus()
	Blur()
	View() string
	SetSize(width, height int)
}

type TextareaComponent struct {
	textarea textarea.Model
}

func (buffer *Buffer) adjustComponentSizes(width, height int) {
	for _, component := range buffer.components {
		component.SetSize(width, height)
	}
}

func (t *TextareaComponent) Init() tea.Cmd {
	return textarea.Blink
}

func (t *TextareaComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := t.textarea.Update(msg)
	t.textarea = newModel
	return t, cmd // Return the wrapper itself as a tea.Model
}

func (t *TextareaComponent) View() string {
	return t.textarea.View()
}

func (t *TextareaComponent) Focus() {
	t.textarea.Focus()
}

func (t *TextareaComponent) Blur() {
	t.textarea.Blur()
}

func (t *TextareaComponent) SetSize(width, height int) {
	t.textarea.SetWidth(width)
	t.textarea.SetHeight(height)
}
