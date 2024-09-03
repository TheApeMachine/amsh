package buffer

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type Component interface {
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type Model struct {
	filename   string
	mode       Mode
	components map[string]Component
	views      map[string]string
	mutex      sync.RWMutex
}

func New() *Model {
	return &Model{
		components: make(map[string]Component),
		views:      make(map[string]string),
	}
}

func (m *Model) RegisterComponent(name string, component Component) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.components[name] = component
}

func (m *Model) UpdateComponentView(name, view string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.views[name] = view
}

func (m *Model) Init() tea.Cmd {
	return nil
}
