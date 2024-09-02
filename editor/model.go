package editor

import (
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/statusbar"
	"github.com/theapemachine/amsh/textarea"
)

type Model struct {
	width, height int
	components    []textarea.Model
	filename      string
	content       string
	lastSaved     string
	saveMutex     sync.Mutex
	mode          Mode
	activeEditor  int
	filePaths     [2]string
	StatusBar     statusbar.Model
}

func NewModel(filename string) *Model {
	m := &Model{
		components:   make([]textarea.Model, 1),
		width:        80,
		height:       24,
		filename:     filename,
		mode:         NormalMode,
		activeEditor: 0,
		filePaths:    [2]string{filename, ""},
	}

	m.StatusBar = statusbar.New()
	m.initializeComponents(filename)
	m.sizeInputs()
	m.components[0].Focus()

	go m.autoSave()

	return m
}

func (m *Model) initializeComponents(filename string) {
	m.components[0] = textarea.New()
	m.loadContent(filename)
}

func (m *Model) loadContent(filename string) {
	content, err := os.ReadFile(filename)
	if err == nil {
		m.content = string(content)
		m.lastSaved = m.content
		m.components[0].SetValue(m.content)
	}
}

func (m *Model) sizeInputs() {
	availableWidth := m.width / len(m.components)
	availableHeight := m.height - 1 // Reserve 1 line for the status bar

	for i := range m.components {
		m.components[i].SetSize(availableWidth, availableHeight)
	}
}

func (m *Model) autoSave() {
	for {
		time.Sleep(500 * time.Millisecond)
		m.saveMutex.Lock()
		if m.content != m.lastSaved && m.filename != "" {
			err := os.WriteFile(m.filename, []byte(m.content), 0644)
			if err == nil {
				m.lastSaved = m.content
			}
		}
		m.saveMutex.Unlock()
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.sizeInputs()
}

// Add this method to implement tea.Model interface
func (m *Model) Init() tea.Cmd {
	return nil
}
