package buffer

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/theapemachine/amsh/logger" // Add this import
	"github.com/theapemachine/amsh/messages"
	"github.com/theapemachine/amsh/ui"
)

type KeyMapping struct {
	Key     string
	Modes   []ui.Mode
	Command string
	Params  string
}

type KeyHandler struct {
	keyMap      map[string][]KeyMapping
	updateFn    func(tea.Msg) (tea.Model, tea.Cmd)
	currentMode ui.Mode
}

func NewKeyHandler(updateFn func(tea.Msg) (tea.Model, tea.Cmd)) *KeyHandler {
	return &KeyHandler{
		keyMap:   make(map[string][]KeyMapping),
		updateFn: updateFn,
	}
}

/*
Start the key handler, and return a channel that can be used to send key messages to the buffer.
*/
func (handler *KeyHandler) Start() chan tea.KeyMsg {
	in := make(chan tea.KeyMsg)

	buf := make([]string, 0)
	timer := time.NewTimer(time.Millisecond * 500) // Increased timer duration
	timer.Stop()

	go func() {
		for {
			select {
			case key := <-in:
				buf = append(buf, key.String())
				logger.Debug("Key pressed: %s, Current buffer: %s", key.String(), strings.Join(buf, ""))

				// Check for command match immediately
				if handler.checkAndExecuteCommand(buf) {
					buf = make([]string, 0)
				} else {
					// Reset timer for potential multi-key commands
					timer.Reset(time.Millisecond * 500)
				}
			case <-timer.C:
				if len(buf) > 0 {
					logger.Debug("Timer expired, checking command: %s", strings.Join(buf, ""))
					handler.checkAndExecuteCommand(buf)
					buf = make([]string, 0)
				}
			}
		}
	}()

	return in
}

func (handler *KeyHandler) checkAndExecuteCommand(buf []string) bool {
	command := strings.Join(buf, "")
	if mappings, ok := handler.keyMap[command]; ok && len(mappings) > 0 {
		mapping := mappings[0]
		logger.Info("Command recognized: %s, Executing: %s with params: %s", command, mapping.Command, mapping.Params)
		messages.NewFromString(mapping.Command, mapping.Params, handler.updateFn)
		return true
	}
	return false
}

func (handler *KeyHandler) RegisterKeyBinding(key string, modes []ui.Mode, command string, params string) {
	mapping := KeyMapping{
		Key:     key,
		Modes:   modes,
		Command: command,
		Params:  params,
	}
	handler.keyMap[key] = append(handler.keyMap[key], mapping)
	logger.Info("Registered key binding: Key: %s, Command: %s, Params: %s", key, command, params)
}

func (handler *KeyHandler) ProcessKey(key tea.KeyMsg, model *Model) (tea.Model, tea.Cmd) {
	keyString := key.String()
	logger.Debug("Processing key: %s", keyString)

	if mappings, ok := handler.keyMap[keyString]; ok && len(mappings) > 0 {
		for _, mapping := range mappings {
			if mapping.Modes == nil || len(mapping.Modes) == 0 || contains(mapping.Modes, handler.currentMode) {
				logger.Info("Executing command for key: %s, Command: %s, Params: %s", keyString, mapping.Command, mapping.Params)
				messages.NewFromString(mapping.Command, mapping.Params, handler.updateFn)
			}
		}
	}

	// If no mapping found, return the model and key as is
	return model, func() tea.Msg { return key }
}

func contains(modes []ui.Mode, mode ui.Mode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}

// Add this method to set the current mode
func (handler *KeyHandler) SetMode(mode ui.Mode) {
	handler.currentMode = mode
}
