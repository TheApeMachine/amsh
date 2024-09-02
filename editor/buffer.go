package editor

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Buffer struct {
	width              int
	height             int
	keyboardMgr        *KeyboardManager
	help               help.Model
	components         []Component
	focus              int
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
	spacePressed       bool
	filePaths          [2]string // Store file paths for both editors
}

type Mode int

const (
	NormalMode Mode = iota
	InsertMode
)

func NewBuffer(filename string) *Buffer {
	buffer := &Buffer{
		components:         make([]Component, 1), // Start with one editor
		help:               help.New(),
		keyboardMgr:        NewKeyboardManager(),
		width:              80, // Set a default width
		height:             24, // Set a default height
		filename:           filename,
		mode:               NormalMode,
		suggestionList:     list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		suggestionsVisible: false,
		focus:              0,
		fileBrowser:        NewFileBrowser(),
		isFileBrowserMode:  filename == "",
		activeEditor:       0,
		spacePressed:       false,
		filePaths:          [2]string{filename, ""},
	}

	if filename == "" {
		buffer.components = []Component{buffer.fileBrowser}
	} else {
		textarea := newTextarea()
		wrappedTextarea := &TextareaComponent{textarea: textarea}
		buffer.components[0] = wrappedTextarea
		content, err := os.ReadFile(filename)
		if err == nil {
			buffer.content = string(content)
			buffer.lastSaved = buffer.content
			buffer.components[0].(*TextareaComponent).textarea.SetValue(buffer.content)
		}
	}

	buffer.keyboardMgr.UpdateKeybindings(len(buffer.components))
	buffer.sizeInputs()

	buffer.components[0].Focus()

	// Start the auto-save goroutine
	go buffer.autoSave()

	return buffer
}

func (buffer *Buffer) autoSave() {
	for {
		time.Sleep(500 * time.Millisecond) // Check every 500ms
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
		switch msg := msg.(type) {
		case FileSelectedMsg:
			buffer.isFileBrowserMode = false
			content, err := os.ReadFile(msg.Path)
			if err == nil {
				buffer.content = string(content)
				buffer.lastSaved = buffer.content
				buffer.filePaths[buffer.activeEditor] = msg.Path
				
				// Update both editors if they have the same file open
				for i, component := range buffer.components {
					if textareaComponent, ok := component.(*TextareaComponent); ok {
						if buffer.filePaths[i] == msg.Path {
							textareaComponent.textarea.SetValue(buffer.content)
						}
					}
				}
				
				buffer.filename = msg.Path
			}
			buffer.keyboardMgr.UpdateKeybindings(len(buffer.components))
			buffer.SetSize(msg.Width, msg.Height)
			buffer.focusActiveEditor()
			return buffer, nil
		}
		newFileBrowser, cmd := buffer.fileBrowser.Update(msg)
		buffer.fileBrowser = newFileBrowser.(*FileBrowser)
		return buffer, cmd
	}

	var cmds []tea.Cmd

	for i, component := range buffer.components {
		newModel, cmd := component.Update(msg)
		buffer.components[i] = newModel.(Component)
		cmds = append(cmds, cmd)

		// Sync content if the same file is open in both editors
		if textareaComponent, ok := newModel.(*TextareaComponent); ok {
			newContent := textareaComponent.textarea.Value()
			if buffer.filePaths[0] == buffer.filePaths[1] && buffer.filePaths[0] != "" {
				for j, otherComponent := range buffer.components {
					if j != i {
						if otherTextareaComponent, ok := otherComponent.(*TextareaComponent); ok {
							otherTextareaComponent.textarea.SetValue(newContent)
						}
					}
				}
			}
			if i == buffer.activeEditor {
				buffer.content = newContent
			}
		}
	}

	// Update content after each keystroke
	buffer.saveMutex.Lock()
	buffer.content = buffer.components[0].(*TextareaComponent).textarea.Value()
	buffer.saveMutex.Unlock()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch buffer.mode {
		case NormalMode:
			switch msg.String() {
			case "i":
				buffer.mode = InsertMode
			case "/":
				buffer.suggestionsVisible = true
				return buffer, buffer.updateSuggestions
			case buffer.keyboardMgr.KeyMap.quit.Help().Key:
				return buffer, tea.Quit
			case "tab":
				if len(buffer.components) < 2 {
					buffer.addEditor()
				}
				buffer.activeEditor = (buffer.activeEditor + 1) % len(buffer.components)
				buffer.focusActiveEditor()
			case " ":
				buffer.spacePressed = true
			case ",":
				if buffer.spacePressed {
					buffer.openFileBrowserForActiveEditor()
					buffer.spacePressed = false
				}
			default:
				buffer.spacePressed = false
			}
		case InsertMode:
			buffer.spacePressed = false
			switch msg.String() {
			case "esc":
				buffer.mode = NormalMode
			}
		}
	case tea.WindowSizeMsg:
		buffer.SetSize(msg.Width, msg.Height)
	}

	// Update suggestion list
	newSuggestionList, cmd := buffer.suggestionList.Update(msg)
	buffer.suggestionList = newSuggestionList
	cmds = append(cmds, cmd)

	return buffer, tea.Batch(cmds...)
}

func (buffer *Buffer) updateSuggestions() tea.Msg {
	// Implement suggestion logic here
	// This is where you'd query your suggestion source
	suggestions := []string{"suggestion1", "suggestion2", "suggestion3"}
	return SuggestionsMsg{Suggestions: suggestions}
}

type SuggestionsMsg struct {
	Suggestions []string
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

	// Display current mode
	modeStr := "NORMAL"
	if buffer.mode == InsertMode {
		modeStr = "INSERT"
	}
	s.WriteString(fmt.Sprintf("-- %s --\n", modeStr))

	help := buffer.help.ShortHelpView([]key.Binding{
		buffer.keyboardMgr.KeyMap.next,
		buffer.keyboardMgr.KeyMap.prev,
		buffer.keyboardMgr.KeyMap.add,
		buffer.keyboardMgr.KeyMap.remove,
		buffer.keyboardMgr.KeyMap.quit,
	})

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

	// Display suggestions if active
	if buffer.mode == NormalMode && buffer.suggestionsVisible {
		s.WriteString("\nSuggestions:\n")
		s.WriteString(buffer.suggestionList.View())
	}

	return fmt.Sprintf("%s\n%s\n\n%s", s.String(), lipgloss.JoinHorizontal(lipgloss.Top, views...), help)
}

func (buffer *Buffer) addEditor() {
	if len(buffer.components) < 2 {
		textarea := newTextarea()
		wrappedTextarea := &TextareaComponent{textarea: textarea}
		buffer.components = append(buffer.components, wrappedTextarea)
		
		// If the first editor has a file open, open the same file in the second editor
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
}

func (buffer *Buffer) openFileBrowserForActiveEditor() {
	buffer.isFileBrowserMode = true
	buffer.fileBrowser = NewFileBrowser()
	buffer.fileBrowser.SetSize(buffer.width, buffer.height)
}
