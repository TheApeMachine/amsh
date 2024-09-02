package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/editor"
	"golang.org/x/term"
)

var (
	path string
)

var rootCmd = &cobra.Command{
	Use:   "amsh",
	Short: "A brief description of your application",
	Long:  roottxt,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		p := tea.NewProgram(
			newModel(path),
			tea.WithAltScreen(),
		)

		if _, err := p.Run(); err != nil {
			fmt.Println("Error while running program:", err)
			os.Exit(1)
		}

		return nil
	},
}

type model struct {
	fileBrowser *editor.FileBrowser
	buffer      *editor.Buffer
	activeView  string
}

func newModel(initialPath string) model {
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))
	
	m := model{
		fileBrowser: editor.NewFileBrowser(),
		activeView:  "fileBrowser",
	}
	
	m.fileBrowser.SetSize(w, h)
	
	if initialPath != "" {
		m.buffer = editor.NewBuffer(initialPath)
		m.buffer.SetSize(w, h)
		m.activeView = "buffer"
	}
	
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.fileBrowser != nil {
			m.fileBrowser.SetSize(msg.Width, msg.Height)
		}
		if m.buffer != nil {
			m.buffer.SetSize(msg.Width, msg.Height)
		}
	case editor.FileSelectedMsg:
		if m.buffer == nil {
			m.buffer = editor.NewBuffer(msg.Path)
		} else {
			var newBuffer tea.Model
			newBuffer, cmd = m.buffer.Update(msg)
			if updatedBuffer, ok := newBuffer.(*editor.Buffer); ok {
				m.buffer = updatedBuffer
			} else {
				// Handle the case where the type assertion fails
				// This shouldn't happen in normal circumstances
				fmt.Println("Error: Failed to update buffer")
				return m, tea.Quit
			}
		}
		m.buffer.SetSize(msg.Width, msg.Height)
		m.activeView = "buffer"
		return m, cmd
	}

	if m.activeView == "fileBrowser" {
		newFileBrowser, newCmd := m.fileBrowser.Update(msg)
		m.fileBrowser = newFileBrowser.(*editor.FileBrowser)
		cmd = newCmd
	} else if m.activeView == "buffer" {
		var newBuffer tea.Model
		newBuffer, cmd = m.buffer.Update(msg)
		if updatedBuffer, ok := newBuffer.(*editor.Buffer); ok {
			m.buffer = updatedBuffer
		} else {
			// Handle the case where the type assertion fails
			fmt.Println("Error: Failed to update buffer")
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m model) View() string {
	if m.activeView == "fileBrowser" {
		return m.fileBrowser.View()
	}
	return m.buffer.View()
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Path to open")
}

const roottxt = `amsh v0.0.1
A minimal shell and vim-like text editor written in Go, with integrated A.I. capabilities.
Different from other A.I. integrations, it uses multiple A.I. models that engage independently
in conversation with each other and the user, improving the developer experience and providing
a more human-like interaction.
`
