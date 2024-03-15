package widgets

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type Prompt struct {
	form *huh.Form
}

func NewPrompt() *Prompt {
	return &Prompt{
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewInput(),
			),
		),
	}
}

func (prompt *Prompt) Init() tea.Cmd {
	return prompt.Init()
}

func (prompt *Prompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	_, cmd := prompt.form.Update(msg)
	return prompt, cmd
}

func (prompt *Prompt) View() string {
	return prompt.form.View()
}
