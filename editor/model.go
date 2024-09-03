package editor

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/textarea"
)

const (
	initialInputs = 2
	maxInputs     = 6
	minInputs     = 1
	helpHeight    = 5
)

type Model struct {
	filename string
	content  string
	cursor   int
	mode     Mode
	width    int
	height   int
	help     help.Model
	inputs   []textarea.Model
	focus    int
}

func New(filename string) *Model {
	m := &Model{
		filename: filename,
		mode:     NormalMode,
		content:  "",
		cursor:   0,
		inputs:   make([]textarea.Model, initialInputs),
		help:     help.New(),
	}

	for i := 0; i < initialInputs; i++ {
		m.inputs[i] = textarea.New()
	}

	m.inputs[m.focus].Focus()
	return m
}

func (m *Model) Init() tea.Cmd {
	return m.inputs[m.focus].Blink()
}

func (m *Model) sizeInputs() {
	for i := range m.inputs {
		m.inputs[i].SetSize(m.width/len(m.inputs), m.height-helpHeight)
	}
}
