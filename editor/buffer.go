package editor

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Buffer struct {
	width              int
	height             int
	components         []Component
	filename           string
	content            string
	lastSaved          string
	saveMutex          sync.Mutex
	mode               Mode
	suggestionList     list.Model
	suggestionsVisible bool
	fileBrowser        *FileBrowser
	isFileBrowserMode  bool
	activeEditor       int
	filePaths          [2]string
	activeComponent    int
	statusBar          *StatusBar
	keyHandler         *KeyHandler
}

func NewBuffer(filename string) *Buffer {
	buffer := &Buffer{
		components:         make([]Component, 1),
		width:              80,
		height:             24,
		filename:           filename,
		mode:               NormalMode,
		suggestionList:     list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		suggestionsVisible: false,
		fileBrowser:        NewFileBrowser(),
		isFileBrowserMode:  filename == "",
		activeEditor:       0,
		filePaths:          [2]string{filename, ""},
		activeComponent:    0,
	}

	buffer.statusBar = NewStatusBar(buffer)
	buffer.keyHandler = NewKeyHandler(buffer)

	buffer.initializeComponents(filename)
	buffer.sizeInputs()
	buffer.components[0].Focus()

	go buffer.autoSave()

	return buffer
}

func (buffer *Buffer) initializeComponents(filename string) {
	if filename == "" {
		buffer.components = []Component{buffer.fileBrowser}
	} else {
		textarea := newTextarea()
		wrappedTextarea := &TextareaComponent{textarea: textarea}
		buffer.components[0] = wrappedTextarea
		buffer.loadContent(filename)
	}
}

func (buffer *Buffer) loadContent(filename string) {
	content, err := os.ReadFile(filename)
	if err == nil {
		buffer.content = string(content)
		buffer.lastSaved = buffer.content
		buffer.components[0].(*TextareaComponent).textarea.SetValue(buffer.content)
	}
}

func (buffer *Buffer) autoSave() {
	for {
		time.Sleep(500 * time.Millisecond)
		buffer.saveMutex.Lock()
		if buffer.content != buffer.lastSaved && buffer.filename != "" {
			err := os.WriteFile(buffer.filename, []byte(buffer.content), 0644)
			if err == nil {
				buffer.lastSaved = buffer.content
			}
		}
		buffer.saveMutex.Unlock()
	}
}

func (buffer *Buffer) Init() tea.Cmd {
	return textarea.Blink
}

func (buffer *Buffer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if buffer.isFileBrowserMode {
		return buffer.updateFileBrowser(msg)
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return buffer.keyHandler.Handle(msg)
	case tea.WindowSizeMsg:
		buffer.SetSize(msg.Width, msg.Height)
	}

	buffer.statusBar.Update()
	return buffer, cmd
}

func (buffer *Buffer) updateFileBrowser(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FileSelectedMsg:
		return buffer.handleFileSelected(msg)
	}
	newFileBrowser, cmd := buffer.fileBrowser.Update(msg)
	buffer.fileBrowser = newFileBrowser.(*FileBrowser)
	return buffer, cmd
}

func (buffer *Buffer) handleFileSelected(msg FileSelectedMsg) (tea.Model, tea.Cmd) {
	buffer.isFileBrowserMode = false
	buffer.loadContent(msg.Path)
	buffer.filePaths[buffer.activeEditor] = msg.Path

	for i, component := range buffer.components {
		if textareaComponent, ok := component.(*TextareaComponent); ok {
			if buffer.filePaths[i] == msg.Path {
				textareaComponent.textarea.SetValue(buffer.content)
			}
		}
	}

	buffer.filename = msg.Path
	buffer.SetSize(msg.Width, msg.Height)
	buffer.focusActiveEditor()
	return buffer, nil
}

func (buffer *Buffer) syncContent(index int, newModel tea.Model) {
	if textareaComponent, ok := newModel.(*TextareaComponent); ok {
		newContent := textareaComponent.textarea.Value()
		if buffer.filePaths[0] == buffer.filePaths[1] && buffer.filePaths[0] != "" {
			for j, otherComponent := range buffer.components {
				if j != index {
					if otherTextareaComponent, ok := otherComponent.(*TextareaComponent); ok {
						otherTextareaComponent.textarea.SetValue(newContent)
					}
				}
			}
		}
		if index == buffer.activeEditor {
			buffer.content = newContent
		}
	}
}

func (buffer *Buffer) updateContent() {
	buffer.saveMutex.Lock()
	buffer.content = buffer.components[0].(*TextareaComponent).textarea.Value()
	buffer.saveMutex.Unlock()
}

func (buffer *Buffer) SetSize(width, height int) {
	buffer.width = width
	buffer.height = height
	if buffer.isFileBrowserMode {
		buffer.fileBrowser.SetSize(width, height)
	} else {
		buffer.sizeInputs()
	}
}

func (buffer *Buffer) sizeInputs() {
	availableWidth := buffer.width / len(buffer.components)
	availableHeight := buffer.height - helpHeight

	for i := range buffer.components {
		buffer.components[i].SetSize(availableWidth, availableHeight)
	}
}

func (buffer *Buffer) View() string {
	if buffer.isFileBrowserMode {
		return buffer.fileBrowser.View()
	}

	var s strings.Builder

	s.WriteString(buffer.statusBar.Render())
	s.WriteString("\n")

	views := buffer.getComponentViews()

	if buffer.mode == NormalMode && buffer.suggestionsVisible {
		s.WriteString("\nSuggestions:\n")
		s.WriteString(buffer.suggestionList.View())
	}

	return fmt.Sprintf("%s\n%s", s.String(), lipgloss.JoinHorizontal(lipgloss.Top, views...))
}

func (buffer *Buffer) getComponentViews() []string {
	var views []string
	for i, component := range buffer.components {
		view := component.View()
		if i == buffer.activeEditor {
			view = focusedBorderStyle.Render(view)
		} else {
			view = blurredBorderStyle.Render(view)
		}
		views = append(views, view)
	}
	return views
}

func (buffer *Buffer) addEditor() {
	if len(buffer.components) < 2 {
		textarea := newTextarea()
		wrappedTextarea := &TextareaComponent{textarea: textarea}
		buffer.components = append(buffer.components, wrappedTextarea)

		if buffer.filePaths[0] != "" {
			content, err := os.ReadFile(buffer.filePaths[0])
			if err == nil {
				wrappedTextarea.textarea.SetValue(string(content))
				buffer.filePaths[1] = buffer.filePaths[0]
			}
		}

		buffer.sizeInputs()
	}
}

func (buffer *Buffer) focusActiveEditor() {
	for i, component := range buffer.components {
		if i == buffer.activeEditor {
			component.Focus()
		} else {
			component.Blur()
		}
	}
	buffer.activeComponent = buffer.activeEditor
}

func (buffer *Buffer) openFileBrowserForActiveEditor() {
	buffer.isFileBrowserMode = true
	buffer.fileBrowser = NewFileBrowser()
	buffer.fileBrowser.SetSize(buffer.width, buffer.height)
}

func (buffer *Buffer) updateSuggestions() tea.Msg {
	suggestions := []string{"suggestion1", "suggestion2", "suggestion3"}
	return SuggestionsMsg{Suggestions: suggestions}
}

type SuggestionsMsg struct {
	Suggestions []string
}

// Add this method to the Buffer struct
func (buffer *Buffer) switchEditor() {
	if len(buffer.components) < 2 {
		buffer.addEditor()
	}
	buffer.activeEditor = (buffer.activeEditor + 1) % len(buffer.components)
	buffer.activeComponent = buffer.activeEditor
	buffer.focusActiveEditor()
}
