package editor

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Define message types within the editor package
type SetFilenameMsg string
type StatusUpdateMsg struct {
	Filename string
	Mode     Mode
}

// Mode represents the editor mode
type Mode int

const (
	NormalMode Mode = iota
	InsertMode
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		newModel, newCmd := m.handleKeyMsg(msg)
		if newCmd != nil {
			cmd = tea.Batch(cmd, newCmd)
		}
		return newModel, cmd

	case SetFilenameMsg:
		m.filename = string(msg)
		return m, m.sendStatusUpdate()

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	default:
		return m, nil
	}
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.mode {
	case NormalMode:
		if msg.String() == "i" {
			m.mode = InsertMode
			return m, m.sendStatusUpdate()
		}
	case InsertMode:
		if msg.String() == "esc" {
			m.mode = NormalMode
			return m, m.sendStatusUpdate()
		}
		newContent, newCursor := m.handleTextInput(msg)
		if newContent != m.content {
			m.content = newContent
			m.cursor = newCursor
			cmd = m.sendStatusUpdate()
		}
	}

	return m, cmd
}

func (m *Model) handleTextInput(msg tea.KeyMsg) (string, int) {
	// Implement text input handling logic here
	return m.content, m.cursor
}

func (m *Model) sendStatusUpdate() tea.Cmd {
	return func() tea.Msg {
		return StatusUpdateMsg{
			Filename: m.filename,
			Mode:     m.mode,
		}
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.sizeInputs()
}
