package buffer

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

/*
Component is an interface that must be implemented by all components that are registered with the buffer.
The buffer is responsible for updating the component views and sending messages to the components.
This interface ensures a consistent structure for all components, allowing the buffer to manage them uniformly.
*/
type Component interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

/*
Model is the model for the buffer. It is responsible for managing the component state and views.
It acts as a central hub for all components, coordinating their interactions and rendering.
The use of a mutex ensures thread-safe access to the shared state, which is crucial for concurrent operations.
*/
type Model struct {
	components      map[string]tea.Model
	activeComponent string
	mutex           sync.RWMutex
	width           int
	height          int
}

/*
New creates a new buffer model.
It initializes the components map and sets the default active component to "filebrowser".
This factory function ensures that every new buffer instance starts with a consistent initial state.
*/
func New() *Model {
	return &Model{
		components:      make(map[string]tea.Model),
		activeComponent: "filebrowser",
	}
}

/*
RegisterComponent registers a component with the buffer, which should establish a
two-way communication channel, so components can exchange messages through the Buffer.
This method is crucial for the modular design of the application, allowing new components
to be added dynamically and ensuring they can participate in the message-passing system.
*/
func (m *Model) RegisterComponent(name string, component Component) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.components[name] = component
}

/*
Init initializes all registered components.
This method is part of the tea.Model interface and is called when the application starts.
It ensures that all components are properly initialized and ready to handle updates and render views.
*/
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.components["filebrowser"].Init(),
		m.components["editor"].Init(),
		m.components["statusbar"].Init(),
	)
}

/*
SetActiveComponent changes the currently active component.
This method is used to switch focus between different parts of the application,
allowing for a dynamic and interactive user interface.
*/
func (m *Model) SetActiveComponent(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.activeComponent = name
}
