package editor

import (
	tea "github.com/charmbracelet/bubbletea"
)

type KeyHandler struct {
	buffer *Buffer
}

func NewKeyHandler(buffer *Buffer) *KeyHandler {
	return &KeyHandler{
		buffer: buffer,
	}
}

func (kh *KeyHandler) Handle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch kh.buffer.mode {
	case NormalMode:
		return kh.handleNormalMode(msg)
	case InsertMode:
		return kh.handleInsertMode(msg)
	}
	return kh.buffer, nil
}

func (kh *KeyHandler) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "i":
		kh.buffer.mode = InsertMode
	case "esc":
		kh.buffer.mode = NormalMode
	case "h", "left", "j", "down", "k", "up", "l", "right":
		newModel, cmd := kh.buffer.components[kh.buffer.activeComponent].Update(msg)
		if textareaComponent, ok := newModel.(*TextareaComponent); ok {
			kh.buffer.components[kh.buffer.activeComponent] = textareaComponent
		}
		return kh.buffer, cmd
	case "q":
		return kh.buffer, tea.Quit
	case "tab":
		kh.buffer.switchEditor()
	}
	return kh.buffer, nil
}

func (kh *KeyHandler) handleInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		kh.buffer.mode = NormalMode
		return kh.buffer, nil
	default:
		newModel, cmd := kh.buffer.components[kh.buffer.activeComponent].Update(msg)
		if textareaComponent, ok := newModel.(*TextareaComponent); ok {
			kh.buffer.components[kh.buffer.activeComponent] = textareaComponent
		}
		return kh.buffer, cmd
	}
}
