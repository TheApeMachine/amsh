package buffer

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/ui"
)

/*
KeyBinding maps a series of keys within a mode and time span to a command
with parameters, which can be scheduled onto the Update queue.
*/
type KeyBinding struct {
	Keys    []string
	Modes   []ui.Mode
	Command string
	Params  []string
}

/*
KeyHandler is a management structure to handle key bindings. It is responsible for
registering key bindings, and checking for key presses to execute the appropriate command.
It is inspired by the way vim handles key bindings, and uses a similar approach to
matching key presses to registered key bindings.
*/
type KeyHandler struct {
	keyMap   map[string][]KeyBinding
	mode     ui.Mode
	buf      []string
	lastTime time.Time
}

/*
NewKeyHandler creates a new key handler with the given current mode and update function.
*/
func NewKeyHandler(mode ui.Mode) *KeyHandler {
	return &KeyHandler{
		keyMap:   make(map[string][]KeyBinding),
		mode:     mode,
		buf:      []string{},
		lastTime: time.Now(),
	}
}

/*
Start the key handler, and return a channel that can be used to send key messages to the buffer.
*/
func (handler *KeyHandler) Handle(msg tea.KeyMsg) tea.Cmd {
	currentTime := time.Now()
	key := msg.String()

	// Clear buffer if more than 500ms have passed since last key press
	if currentTime.Sub(handler.lastTime) > 500*time.Millisecond {
		handler.buf = []string{}
	}

	handler.buf = append(handler.buf, key)
	handler.lastTime = currentTime

	if handler.checkAndExecuteCommand(handler.buf) {
		cmd := handler.executeCommand(strings.Join(handler.buf, ""))
		handler.buf = []string{}
		return cmd
	}

	if !handler.hasPartialMatch(handler.buf) {
		handler.buf = []string{}
	}

	return nil
}

/*
hasPartialMatch checks if the buffer has a partial match for any key binding.
This is useful, since if there is not a partial match, there will never be a full match.
That means we can stop the process early.
*/
func (handler *KeyHandler) hasPartialMatch(buf []string) bool {
	prefix := strings.Join(buf, "")
	for command := range handler.keyMap {
		if strings.HasPrefix(command, prefix) {
			return true
		}
	}
	return false
}

/*
checkAndExecuteCommand checks if the buffer has a full match for any key binding and executes the command if it does.
*/
func (handler *KeyHandler) checkAndExecuteCommand(buf []string) bool {
	command := strings.Join(buf, "")
	_, ok := handler.keyMap[command]
	return ok
}

/*
RegisterKeyBinding registers a new key binding with the key handler.
*/
func (handler *KeyHandler) RegisterKeyBinding(key string, modes []ui.Mode, command string, params string) {
	handler.keyMap[key] = append(handler.keyMap[key], KeyBinding{
		Keys:    []string{key},
		Modes:   modes,
		Command: command,
		Params:  []string{params},
	})
}

/*
SetMode sets the current mode of the key handler.
*/
func (handler *KeyHandler) SetMode(mode ui.Mode) {
	logger.Debug("Setting keyhandler mode to: %T", mode)
	handler.mode = mode
}

func (handler *KeyHandler) executeCommand(command string) tea.Cmd {
	if mappings, ok := handler.keyMap[command]; ok && len(mappings) > 0 {
		mapping := mappings[0]
		logger.Info("Command recognized: %s, Executing: %s with params: %v", command, mapping.Command, mapping.Params)
		return func() tea.Msg {
			return data.New(
				"KeyHandler",
				mapping.Command,
				strings.Join(mapping.Params, ","),
				[]byte{},
			)
		}
	}
	return nil
}
