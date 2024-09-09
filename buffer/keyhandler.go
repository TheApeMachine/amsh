package buffer

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/messages"
	"github.com/theapemachine/amsh/ui"
)

/*
KeyMapping maps the configuration for a key binding.
*/
type KeyMapping struct {
	Key     string
	Modes   []ui.Mode
	Command string
	Params  string
}

/*
KeyHandler manages the key bindings for the application in a way similar to vim.
*/
type KeyHandler struct {
	keyMap      map[string][]KeyMapping
	updateFn    func(tea.Msg) (tea.Model, tea.Cmd)
	currentMode ui.Mode
}

/*
NewKeyHandler creates a new key handler with the given current mode and update function.
*/
func NewKeyHandler(currentMode ui.Mode, updateFn func(tea.Msg) (tea.Model, tea.Cmd)) *KeyHandler {
	return &KeyHandler{
		keyMap:      make(map[string][]KeyMapping),
		updateFn:    updateFn,
		currentMode: currentMode,
	}
}

/*
Start the key handler, and return a channel that can be used to send key messages to the buffer.
*/
func (handler *KeyHandler) Start() chan tea.KeyMsg {
	// We create a channel to return to the buffer so it can send key messages to the key handler.
	in := make(chan tea.KeyMsg)

	buf := make([]string, 0)
	keyIndex := 0
	partialMatch := false
	timer := time.NewTimer(time.Millisecond * 500)
	timer.Stop()

	go func() {
		for {
			select {
			case key := <-in:
				// If we have ran the code below at least once, and we don't have a partial match,
				// it makes no more sense to continue checking.
				if keyIndex > 0 && !partialMatch {
					continue
				}

				keyIndex++
				buf = append(buf, key.String())

				// Check for a partial match, which will determine if we should reset the timer.
				// Resetting the timer will allow us to check for a full match later.
				if partialMatch = handler.hasPartialMatch(buf); partialMatch {
					timer.Reset(time.Millisecond * 500)
					continue
				}
			case <-timer.C:
				if partialMatch && len(buf) > 0 {
					handler.checkAndExecuteCommand(buf)
				}

				// Make sure we reset the partial check flag and the key index.
				buf = make([]string, 0)
				keyIndex = 0
				partialMatch = false
			}
		}
	}()

	return in
}

/*
hasPartialMatch checks if the buffer has a partial match for any key binding.
This is useful, since if there is not a partial match, there will never be a full match.
That means we can stop the process early.
*/
func (handler *KeyHandler) hasPartialMatch(buf []string) bool {
	for command := range handler.keyMap {
		logger.Debug("Checking command: %s", command)
		if strings.HasPrefix(command, strings.Join(buf, "")) {
			logger.Debug("Partial match found for command: %s", command)
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
	if mappings, ok := handler.keyMap[command]; ok && len(mappings) > 0 {
		mapping := mappings[0]
		logger.Info("Command recognized: %s, Executing: %s with params: %s", command, mapping.Command, mapping.Params)
		messages.NewFromString(mapping.Command, mapping.Params, handler.updateFn)
		return true
	}
	return false
}

/*
RegisterKeyBinding registers a new key binding with the key handler.
*/
func (handler *KeyHandler) RegisterKeyBinding(key string, modes []ui.Mode, command string, params string) {
	handler.keyMap[key] = append(handler.keyMap[key], KeyMapping{
		Key:     key,
		Modes:   modes,
		Command: command,
		Params:  params,
	})
}

/*
SetMode sets the current mode of the key handler.
*/
func (handler *KeyHandler) SetMode(mode ui.Mode) {
	logger.Debug("Setting keyhandler mode to: %T", mode)
	handler.currentMode = mode
}
