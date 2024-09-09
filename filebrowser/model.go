package filebrowser

import (
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
)

/*
Model represents the state of the file browser component.
It manages the list of files/directories, current path, and selected file.
*/
type Model struct {
	filepicker  filepicker.Model
	currentPath string
	state       components.State
	width       int
	height      int
}

/*
New creates a new file browser model.
It initializes the model with the current directory and sets up the file list.
*/
func New(width, height int) *Model {
	path, _ := os.UserHomeDir()
	return &Model{
		filepicker:  filepicker.New(),
		currentPath: path,
		state:       components.Inactive,
		width:       width,
		height:      height,
	}
}

func (model *Model) Init() tea.Cmd {
	model.filepicker.CurrentDirectory = model.currentPath
	model.filepicker.AutoHeight = true
	return model.filepicker.Init()
}
