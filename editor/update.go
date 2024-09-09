package editor

import (
	"bufio"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/components"
	"github.com/theapemachine/amsh/logger"
	"github.com/theapemachine/amsh/lsp"
	"github.com/theapemachine/amsh/messages"
	"github.com/theapemachine/amsh/textarea"
	"github.com/theapemachine/amsh/ui"
)

func (model *Model) handleLSPResponses(responseCh chan lsp.LSPResponse, errorCh chan lsp.LSPError) {
	for {
		select {
		case response := <-responseCh:
			logger.Debug("Received LSP response: %v", response)
			// Handle different types of responses here
		case err := <-errorCh:
			logger.Error("Received LSP error: %v", err)
		}
	}
}

/*
Update handles all incoming messages for the editor component.
This method is part of the tea.Model interface and is responsible for updating the editor state
based on various events such as key presses, file selection, and window size changes.
It delegates to specific handlers based on the current editing mode and message type.
*/
func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		model.inputs[model.focus].Update(msg)
	case messages.Message[ui.Mode]:
		if !messages.ShouldProcessMessage(model.state, msg.Context) {
			return model, nil
		}

		model.mode = msg.Data
		model.inputs[model.focus].Update(msg)
	case messages.Message[[]int]:
		if !messages.ShouldProcessMessage(model.state, msg.Context) {
			return model, nil
		}

		model.handleWindowSizeMsg(msg)
	case messages.Message[string]:
		if !messages.ShouldProcessMessage(model.state, msg.Context) {
			return model, nil
		}

		switch msg.Type {
		case messages.MessageEditor:
			switch msg.Data {
			case "next":
				if model.focus < len(model.inputs)-1 {
					model.focus++
				} else {
					model.addTextarea()
				}
			}
		case messages.MessageOpenFile:
			if model.err = model.loadFile(msg.Data); model.err != nil {
				logger.Log("Error opening file: %v", model.err)
				cmds = append(cmds, func() tea.Msg {
					return messages.NewMessage(
						messages.MessageError, model.err, messages.All,
					)
				})
			}
		}
	}

	return model, tea.Batch(cmds...)
}

/*
handleWindowSizeMsg handles window resizing messages.
*/
func (model *Model) handleWindowSizeMsg(msg messages.Message[[]int]) {
	model.width, model.height = msg.Data[0], msg.Data[1]
	model.resizeTextareas()
}

/*
addTextarea adds a new textarea to the editor.
*/
func (model *Model) addTextarea() *textarea.Model {
	logger.Debug("Adding textarea to editor")
	model.inputs = append(model.inputs, textarea.New(model.width, model.height))
	model.resizeTextareas()
	model.focus = len(model.inputs) - 1
	model.inputs[model.focus].Focus()
	return model.inputs[model.focus]
}

/*
loadFile loads a file into the editor, creating a new textarea for it.
*/
func (model *Model) loadFile(path string) error {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var content []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Set the file content in the editor
	if len(model.inputs) == 0 {
		model.focus = 0
		added := model.addTextarea()
		added.SetValue(strings.Join(content, "\n"))
	}

	model.state = components.Active
	model.currentFile = path
	return nil
}

/*
resizeTextareas resizes all textareas based on the current width and height.
*/
func (model *Model) resizeTextareas() {
	for _, input := range model.inputs {
		input.SetWidth(model.width / len(model.inputs))
		input.SetHeight(model.height)
	}
}
