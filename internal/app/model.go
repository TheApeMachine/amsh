package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/editor"
	"golang.org/x/term"
)

type Model struct {
	FileBrowser *editor.FileBrowser
	Buffer      *editor.Buffer
	ActiveView  string
}

func NewModel(initialPath string) Model {
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))
	
	m := Model{
		FileBrowser: editor.NewFileBrowser(),
		ActiveView:  "fileBrowser",
	}
	
	m.FileBrowser.SetSize(w, h)
	
	if initialPath != "" {
		m.Buffer = editor.NewBuffer(initialPath)
		m.Buffer.SetSize(w, h)
		m.ActiveView = "buffer"
	}
	
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.FileBrowser != nil {
			m.FileBrowser.SetSize(msg.Width, msg.Height)
		}
		if m.Buffer != nil {
			m.Buffer.SetSize(msg.Width, msg.Height)
		}
	case editor.FileSelectedMsg:
		if m.Buffer == nil {
			m.Buffer = editor.NewBuffer(msg.Path)
		} else {
			var newBuffer tea.Model
			newBuffer, cmd = m.Buffer.Update(msg)
			if updatedBuffer, ok := newBuffer.(*editor.Buffer); ok {
				m.Buffer = updatedBuffer
			} else {
				fmt.Println("Error: Failed to update buffer")
				return m, tea.Quit
			}
		}
		m.Buffer.SetSize(msg.Width, msg.Height)
		m.ActiveView = "buffer"
		return m, cmd
	}

	if m.ActiveView == "fileBrowser" {
		newFileBrowser, newCmd := m.FileBrowser.Update(msg)
		m.FileBrowser = newFileBrowser.(*editor.FileBrowser)
		cmd = newCmd
	} else if m.ActiveView == "buffer" {
		var newBuffer tea.Model
		newBuffer, cmd = m.Buffer.Update(msg)
		if updatedBuffer, ok := newBuffer.(*editor.Buffer); ok {
			m.Buffer = updatedBuffer
		} else {
			fmt.Println("Error: Failed to update buffer")
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m Model) View() string {
	if m.ActiveView == "fileBrowser" {
		return m.FileBrowser.View()
	}
	return m.Buffer.View()
}