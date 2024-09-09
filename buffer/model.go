package buffer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
	"github.com/theapemachine/amsh/ui"
)

/*
Model is the model for the buffer. It is responsible for managing the component state and views.
It acts as a central hub for all components, coordinating their interactions and rendering.
The use of a mutex ensures thread-safe access to the shared state, which is crucial for concurrent operations.
*/
type Model struct {
	components []tea.Model
	width      int
	height     int
	path       string
	mode       ui.Mode
	keyHandler *KeyHandler
	cmdChan    chan tea.KeyMsg
}

/*
New creates a new buffer model.
It initializes the components map and sets the default active component to "filebrowser".
This factory function ensures that every new buffer instance starts with a consistent initial state.
*/
func New(path string, width, height int) *Model {
	return &Model{
		components: make([]tea.Model, 0),
		path:       path,
		width:      width,
		height:     height,
		mode:       ui.ModeNormal,
	}
}

/*
Init initializes the buffer model. It initializes all components and returns a command to be executed.
*/
func (model *Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	model.keyHandler = NewKeyHandler(model.mode, model.Update)
	model.cmdChan = model.keyHandler.Start()

	model.LoadKeyMappings()

	for _, component := range model.components {
		cmds = append(cmds, component.Init())
	}

	if model.path != "" {
		cmds = append(cmds, func() tea.Msg {
			return messages.NewMessage(
				messages.MessageOpenFile, model.path, messages.All,
			)
		})
	}

	return tea.Batch(cmds...)
}

/*
RegisterComponents registers one or more component with the buffer, which exposes the Update
method of the tea.Model interface that each component must implement.
*/
func (model *Model) RegisterComponents(name string, components ...tea.Model) {
	model.components = append(model.components, components...)
}

func (model *Model) RegisterKeyBinding(key string, modes []ui.Mode, command string, params string) {
	model.keyHandler.RegisterKeyBinding(key, modes, command, params)
}

func (model *Model) LoadKeyMappings() {
	mappings := viper.Get("keymap.mapping").([]interface{})
	for _, m := range mappings {
		mapping := m.(map[string]interface{})
		key := mapping["key"].(string)
		modes := []ui.Mode{ui.ModeFromString(mapping["modes"].(string))}
		command := mapping["command"].(string)
		params := ""
		if p, ok := mapping["params"]; ok {
			params = p.(string)
		}

		logger.Info("Loaded key mapping: Key: %s, Command: %s, Params: %s", key, command, params)
		model.RegisterKeyBinding(key, modes, command, params)
	}
	logger.Info("Loaded %d key mappings", len(mappings))
}
