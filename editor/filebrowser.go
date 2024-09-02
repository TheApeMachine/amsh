package editor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type FileBrowser struct {
	list        list.Model
	currentPath string
	width       int
	height      int
}

func NewFileBrowser() *FileBrowser {
	fb := &FileBrowser{
		currentPath: ".",
		width:       80,
		height:      24,
	}
	fb.initList()
	return fb
}

func (f *FileBrowser) initList() {
	items := f.getFileItems()
	delegate := list.NewDefaultDelegate()
	f.list = list.New(items, delegate, f.width, f.height)
	f.list.Title = fmt.Sprintf("File Browser - %s", f.currentPath)
	f.list.SetShowStatusBar(false)
	f.list.SetFilteringEnabled(false)
	f.updateKeyMap()
}

func (f *FileBrowser) Init() tea.Cmd {
	return nil
}

func (f *FileBrowser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			return f, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if i, ok := f.list.SelectedItem().(fileBrowserItem); ok {
				if i.isDir {
					f.currentPath = i.path
					f.initList()
					return f, nil
				}
				return f, func() tea.Msg {
					return FileSelectedMsg{Path: i.path, Width: f.width, Height: f.height}
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("backspace", "left"))):
			if f.currentPath != "." {
				f.currentPath = filepath.Dir(f.currentPath)
				f.initList()
				return f, nil
			}
		}
	}

	f.list, cmd = f.list.Update(msg)
	return f, cmd
}

func (f *FileBrowser) View() string {
	return appStyle.Render(f.list.View())
}

func (f *FileBrowser) Focus() {}

func (f *FileBrowser) Blur() {}

func (f *FileBrowser) SetSize(width, height int) {
	f.width = width
	f.height = height
	f.list.SetSize(width, height)
}

func (f *FileBrowser) getFileItems() []list.Item {
	items := []list.Item{}
	files, err := os.ReadDir(f.currentPath)
	if err != nil {
		return items
	}
	if f.currentPath != "." {
		items = append(items, fileBrowserItem{
			path:  filepath.Dir(f.currentPath),
			isDir: true,
			title: "..",
		})
	}
	for _, file := range files {
		items = append(items, fileBrowserItem{
			path:  filepath.Join(f.currentPath, file.Name()),
			isDir: file.IsDir(),
			title: file.Name(),
		})
	}
	return items
}

func (f *FileBrowser) updateKeyMap() {
	f.list.KeyMap.Quit.SetEnabled(false)
	f.list.KeyMap.ForceQuit.SetEnabled(false)
	f.list.KeyMap.CloseFullHelp.SetEnabled(false)
	f.list.KeyMap.ShowFullHelp.SetEnabled(false)
	f.list.KeyMap.CancelWhileFiltering.SetEnabled(false)
}

type fileBrowserItem struct {
	path  string
	isDir bool
	title string
}

func (i fileBrowserItem) Title() string {
	if i.isDir {
		return fmt.Sprintf("📁 %s", i.title)
	}
	return fmt.Sprintf("📄 %s", i.title)
}

func (i fileBrowserItem) Description() string {
	return i.path
}

func (i fileBrowserItem) FilterValue() string {
	return i.title
}

type FileSelectedMsg struct {
	Path   string
	Width  int
	Height int
}
