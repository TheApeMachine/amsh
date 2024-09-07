package filebrowser

import (
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/ui"
)

/*
Model represents the state of the file browser component.
It manages the list of files/directories, current path, and selected file.
*/
type Model struct {
	filepicker   filepicker.Model
	currentPath  string
	selectedFile string
	active       bool
	layout       ui.LayoutPreference
	err          error
}

/*
New creates a new file browser model.
It initializes the model with the current directory and sets up the file list.
*/
func New() *Model {
	path, _ := os.UserHomeDir()
	return &Model{
		filepicker:  filepicker.New(),
		currentPath: path,
		active:      false,
		layout:      ui.Vertical,
	}
}

func (model *Model) Init() tea.Cmd {
	model.filepicker.CurrentDirectory = model.currentPath
	return model.filepicker.Init()
}
