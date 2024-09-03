package filebrowser

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theapemachine/amsh/messages"
)

/*
fileItem represents a file or directory in the file browser.
It implements the list.Item interface, allowing it to be used with the list bubble.
*/
type fileItem struct {
	path string
	info fs.FileInfo
}

// FilterValue returns the path of the file item, used for filtering in the list.
func (i fileItem) FilterValue() string { return i.path }

// Title returns the name of the file or directory.
func (i fileItem) Title() string { return i.info.Name() }

// Description returns either "Directory" for directories or the file size for files.
func (i fileItem) Description() string {
	if i.info.IsDir() {
		return "Directory"
	}
	return fmt.Sprintf("%d bytes", i.info.Size())
}

/*
Model represents the state of the file browser component.
It manages the list of files/directories, current path, and selected file.
*/
type Model struct {
	list         list.Model
	currentPath  string
	selectedFile string
	err          error
}

/*
New creates a new file browser model.
It initializes the model with the current directory and sets up the file list.
*/
func New() *Model {
	m := &Model{
		currentPath: ".",
	}
	m.initList()
	return m
}

// Init initializes the file browser model. It's part of the tea.Model interface.
func (m *Model) Init() tea.Cmd {
	return nil
}

/*
SetSize adjusts the size of the file browser list.
This method is crucial for responsive design, ensuring the file browser
adapts to window size changes.
*/
func (m *Model) SetSize(width, height int) {
	m.list.SetSize(width, height-1) // Leave space for status line
}

/*
initList initializes the file list with items from the current directory.
This method is called when creating a new model or changing directories.
*/
func (m *Model) initList() {
	items, err := m.getFilesInDirectory(m.currentPath)
	if err != nil {
		m.err = err
		return
	}

	delegate := list.NewDefaultDelegate()
	m.list = list.New(items, delegate, 0, 0)
	m.list.Title = "File Browser"
	m.list.SetShowHelp(false)
}

/*
getFilesInDirectory retrieves all files and directories in the given path.
It sorts the items to display directories first, then files, both in alphabetical order.
This method is crucial for populating the file browser with accurate and organized content.
*/
func (m *Model) getFilesInDirectory(path string) ([]list.Item, error) {
	var items []list.Item

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// Sort directories first, then files, both alphabetically
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir() == files[j].IsDir() {
			return files[i].Name() < files[j].Name()
		}
		return files[i].IsDir()
	})

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		items = append(items, fileItem{
			path: filepath.Join(path, file.Name()),
			info: info,
		})
	}

	return items, nil
}

/*
statusLine generates a string representing the current directory.
This provides context to the user about their location in the file system.
*/
func (m *Model) statusLine() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf("Current directory: %s", m.currentPath))
}

/*
sendFileSelected creates a command to send a FileSelectedMsg.
This is used when a file is selected to notify other components.
*/
func (m *Model) sendFileSelected() tea.Cmd {
	return func() tea.Msg {
		return messages.FileSelectedMsg(m.selectedFile)
	}
}
