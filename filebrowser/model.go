package filebrowser

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	list         list.Model
	currentPath  string
	selectedFile string
	width        int
	height       int
}

func New() *Model {
	m := &Model{
		currentPath: ".",
		width:       80,
		height:      24,
	}
	m.initList()
	return m
}

func (m *Model) initList() {
	items := m.getFileItems()
	m.list = list.New(items, list.NewDefaultDelegate(), m.width, m.height)
	m.list.Title = "File Browser"
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

func (m *Model) getFileItems() []list.Item {
	files, _ := os.ReadDir(m.currentPath)
	items := make([]list.Item, 0, len(files))
	for _, file := range files {
		items = append(items, fileItem{
			name:  file.Name(),
			path:  filepath.Join(m.currentPath, file.Name()),
			isDir: file.IsDir(),
		})
	}
	return items
}

type fileItem struct {
	name  string
	path  string
	isDir bool
}

func (i fileItem) Title() string       { return i.name }
func (i fileItem) Description() string { return i.path }
func (i fileItem) FilterValue() string { return i.name }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			return m.handleEnterKey()
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	case OpenFileBrowserMsg:
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return m.list.View()
}

func (m *Model) sendFileSelected() tea.Cmd {
	return func() tea.Msg {
		return FileSelectedMsg(m.selectedFile)
	}
}

// Add Init method to implement tea.Model interface
func (m Model) Init() tea.Cmd {
	return nil
}
