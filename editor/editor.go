package editor

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/messages"
)

type Editor struct {
	textarea textarea.Model
	mode     messages.Mode
}

func NewEditor() *Editor {
	e := &Editor{
		textarea: textarea.New(),
		mode:     messages.NormalMode,
	}
	e.textarea.ShowLineNumbers = true
	return e
}

func (e *Editor) SetContent(content string) {
	e.textarea.SetValue(content)
}

func (e *Editor) GetContent() string {
	return e.textarea.Value()
}

func (e *Editor) SetSize(width, height int) {
	e.textarea.SetWidth(width)
	e.textarea.SetHeight(height)
}

func (e *Editor) Update(msg tea.Msg) (*Editor, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch e.mode {
		case messages.NormalMode:
			if msg.String() == "i" {
				e.mode = messages.InsertMode
				return e, nil
			}
		case messages.InsertMode:
			if msg.String() == "esc" {
				e.mode = messages.NormalMode
				return e, nil
			}
		}
	}

	e.textarea, cmd = e.textarea.Update(msg)
	return e, cmd
}

func (e *Editor) View() string {
	return e.textarea.View()
}

// Add Init method to implement tea.Model interface
func (e *Editor) Init() tea.Cmd {
	return nil
}