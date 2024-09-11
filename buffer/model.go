package buffer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
	"github.com/theapemachine/amsh/ui"
)

/*
Model represent a buffer that acts as a multiplexer for multiple components.
It is responsible for io, message routing, and rendering a composite view of
all active components.
*/
type Model struct {
	components []tea.Model
	width      int
	height     int
	path       string
	mode       ui.Mode
	keyHandler *KeyHandler
}

/*
New returns an instance of a buffer, and provides an entry point for the
BubbleTea application to start.
We take in the path from the command line, if it was provided, which opens
either the editor for a file, or the file browser for a directory.
Additionally we take in the width and height as reported by the terminal.
*/
func New(path string, width, height int) *Model {
	return &Model{
		components: make([]tea.Model, 0),
		path:       path,
		width:      width,
		height:     height,
		mode:       ui.ModeNormal,
		keyHandler: NewKeyHandler(ui.ModeNormal),
	}
}

/*
Init provides an initialization stage to prepare the buffer for use.
It returns one or more commands that will be pushed on the Update queue.
*/
func (model *Model) Init() tea.Cmd {
	var cmds []tea.Cmd
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

/*
RegisterKeyBinding registers a key binding for the given modes.
*/
func (model *Model) RegisterKeyBinding(key string, modes []ui.Mode, command string, params string) {
	model.keyHandler.RegisterKeyBinding(key, modes, command, params)
}

/*
LoadKeyMappings loads the key mappings from the configuration file.
*/
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
