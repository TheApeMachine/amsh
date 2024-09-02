package buffer

import (
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/messages"
)

type Model struct {
	filename  string
	content   string
	lastSaved string
	saveMutex sync.Mutex
	width     int
	height    int
	mode      messages.Mode
}

func New(filename string) *Model {
	m := &Model{
		filename: filename,
		mode:     messages.NormalMode,
	}
	m.loadContent()
	return m
}

// Add Init method to implement tea.Model interface
func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) loadContent() {
	content, err := os.ReadFile(m.filename)
	if err == nil {
		m.content = string(content)
		m.lastSaved = m.content
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) updateContent() {
	m.saveMutex.Lock()
	defer m.saveMutex.Unlock()
}

func (m *Model) sendStatusUpdate() tea.Cmd {
	return func() tea.Msg {
		return messages.StatusUpdateMsg{
			Filename: m.filename,
			Mode:     m.mode,
		}
	}
}
