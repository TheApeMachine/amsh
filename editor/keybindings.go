package editor

import (
	tea "github.com/charmbracelet/bubbletea"
)

type KeyHandler struct {
	editor       *Model
	spacePressed bool
}

func NewKeyHandler(editor *Model) *KeyHandler {
	return &KeyHandler{
		editor:       editor,
		spacePressed: false,
	}
}

func (kh *KeyHandler) Handle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch kh.editor.mode {
	case NormalMode:
		return kh.handleNormalMode(msg)
	case InsertMode:
		return kh.handleInsertMode(msg)
	}
	return kh.editor, nil
}

func (kh *KeyHandler) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "i":
		kh.editor.mode = InsertMode
		return kh.editor, kh.editor.sendStatusUpdate()
	case "esc":
		kh.editor.mode = NormalMode
		return kh.editor, kh.editor.sendStatusUpdate()
	case "h", "left", "j", "down", "k", "up", "l", "right":
		return kh.editor.Update(msg)
	case "q":
		return kh.editor, tea.Quit
	case "tab":
		kh.editor.activeEditor = (kh.editor.activeEditor + 1) % len(kh.editor.components)
		return kh.editor, nil
	case " ":
		kh.spacePressed = true
	case ",":
		if kh.spacePressed {
			// Open file browser (not implemented yet)
			kh.spacePressed = false
		}
	default:
		kh.spacePressed = false
	}
	return kh.editor, nil
}

func (kh *KeyHandler) handleInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		kh.editor.mode = NormalMode
		return kh.editor, kh.editor.sendStatusUpdate()
	default:
		// Pass the key to the active textarea
		newTextarea, cmd := kh.editor.components[kh.editor.activeEditor].Update(msg)
		kh.editor.components[kh.editor.activeEditor] = newTextarea
		kh.editor.syncContent(kh.editor.activeEditor)
		return kh.editor, cmd
	}
}

// Remove the Init method as it's likely part of the buffer package now
