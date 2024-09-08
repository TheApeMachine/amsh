package textarea

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyMap is the key bindings for different actions within the textarea.
type KeyMap struct {
	CharacterBackward       key.Binding
	CharacterForward        key.Binding
	DeleteAfterCursor       key.Binding
	DeleteBeforeCursor      key.Binding
	DeleteCharacterBackward key.Binding
	DeleteCharacterForward  key.Binding
	DeleteWordBackward      key.Binding
	DeleteWordForward       key.Binding
	InsertNewline           key.Binding
	LineEnd                 key.Binding
	LineNext                key.Binding
	LinePrevious            key.Binding
	LineStart               key.Binding
	Paste                   key.Binding
	WordBackward            key.Binding
	WordForward             key.Binding
	InputBegin              key.Binding
	InputEnd                key.Binding

	UppercaseWordForward  key.Binding
	LowercaseWordForward  key.Binding
	CapitalizeWordForward key.Binding

	TransposeCharacterBackward key.Binding
}

// DefaultKeyMap is the default set of key bindings for navigating and acting
// upon the textarea.
var DefaultKeyMap = KeyMap{
	CharacterForward:           key.NewBinding(key.WithKeys("right", "ctrl+f"), key.WithHelp("right", "character forward")),
	CharacterBackward:          key.NewBinding(key.WithKeys("left", "ctrl+b"), key.WithHelp("left", "character backward")),
	WordForward:                key.NewBinding(key.WithKeys("alt+right", "alt+f"), key.WithHelp("alt+right", "word forward")),
	WordBackward:               key.NewBinding(key.WithKeys("alt+left", "alt+b"), key.WithHelp("alt+left", "word backward")),
	LineNext:                   key.NewBinding(key.WithKeys("down", "ctrl+n"), key.WithHelp("down", "next line")),
	LinePrevious:               key.NewBinding(key.WithKeys("up", "ctrl+p"), key.WithHelp("up", "previous line")),
	DeleteWordBackward:         key.NewBinding(key.WithKeys("alt+backspace", "ctrl+w"), key.WithHelp("alt+backspace", "delete word backward")),
	DeleteWordForward:          key.NewBinding(key.WithKeys("alt+delete", "alt+d"), key.WithHelp("alt+delete", "delete word forward")),
	DeleteAfterCursor:          key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "delete after cursor")),
	DeleteBeforeCursor:         key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "delete before cursor")),
	InsertNewline:              key.NewBinding(key.WithKeys("enter", "ctrl+m"), key.WithHelp("enter", "insert newline")),
	DeleteCharacterBackward:    key.NewBinding(key.WithKeys("backspace", "ctrl+h"), key.WithHelp("backspace", "delete character backward")),
	DeleteCharacterForward:     key.NewBinding(key.WithKeys("delete", "ctrl+d"), key.WithHelp("delete", "delete character forward")),
	LineStart:                  key.NewBinding(key.WithKeys("home", "ctrl+a"), key.WithHelp("home", "line start")),
	LineEnd:                    key.NewBinding(key.WithKeys("end", "ctrl+e"), key.WithHelp("end", "line end")),
	Paste:                      key.NewBinding(key.WithKeys("ctrl+v"), key.WithHelp("ctrl+v", "paste")),
	InputBegin:                 key.NewBinding(key.WithKeys("alt+<", "ctrl+home"), key.WithHelp("alt+<", "input begin")),
	InputEnd:                   key.NewBinding(key.WithKeys("alt+>", "ctrl+end"), key.WithHelp("alt+>", "input end")),
	CapitalizeWordForward:      key.NewBinding(key.WithKeys("alt+c"), key.WithHelp("alt+c", "capitalize word forward")),
	LowercaseWordForward:       key.NewBinding(key.WithKeys("alt+l"), key.WithHelp("alt+l", "lowercase word forward")),
	UppercaseWordForward:       key.NewBinding(key.WithKeys("alt+u"), key.WithHelp("alt+u", "uppercase word forward")),
	TransposeCharacterBackward: key.NewBinding(key.WithKeys("ctrl+t"), key.WithHelp("ctrl+t", "transpose character backward")),
}

func (model *Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		model.easyMotion()
	}

	return model, nil
}

func (model *Model) handleInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, model.KeyMap.DeleteAfterCursor):
		model.col = clamp(model.col, 0, len(model.value[model.row]))
		if model.col >= len(model.value[model.row]) {
			model.mergeLineBelow(model.row)
			break
		}
		model.deleteAfterCursor()
	case key.Matches(msg, model.KeyMap.DeleteBeforeCursor):
		model.col = clamp(model.col, 0, len(model.value[model.row]))
		if model.col <= 0 {
			model.mergeLineAbove(model.row)
			break
		}
		model.deleteBeforeCursor()
	case key.Matches(msg, model.KeyMap.DeleteCharacterBackward):
		model.col = clamp(model.col, 0, len(model.value[model.row]))
		if model.col <= 0 {
			model.mergeLineAbove(model.row)
			break
		}
		if len(model.value[model.row]) > 0 {
			model.value[model.row] = append(model.value[model.row][:max(0, model.col-1)], model.value[model.row][model.col:]...)
			model.SetCursor(model.col - 1)
		}
	case key.Matches(msg, model.KeyMap.DeleteCharacterForward):
		if len(model.value[model.row]) > 0 && model.col < len(model.value[model.row]) {
			model.value[model.row] = append(model.value[model.row][:model.col], model.value[model.row][model.col+1:]...)
		}
		if model.col >= len(model.value[model.row]) {
			model.mergeLineBelow(model.row)
			break
		}
	case key.Matches(msg, model.KeyMap.DeleteWordBackward):
		if model.col <= 0 {
			model.mergeLineAbove(model.row)
			break
		}
		model.deleteWordLeft()
	case key.Matches(msg, model.KeyMap.DeleteWordForward):
		model.col = clamp(model.col, 0, len(model.value[model.row]))
		if model.col >= len(model.value[model.row]) {
			model.mergeLineBelow(model.row)
			break
		}
		model.deleteWordRight()
	case key.Matches(msg, model.KeyMap.InsertNewline):
		if model.MaxHeight > 0 && len(model.value) >= model.MaxHeight {
			return model, nil
		}
		model.col = clamp(model.col, 0, len(model.value[model.row]))
		model.splitLine(model.row, model.col)
	case key.Matches(msg, model.KeyMap.LineEnd):
		model.CursorEnd()
	case key.Matches(msg, model.KeyMap.LineStart):
		model.CursorStart()
	case key.Matches(msg, model.KeyMap.CharacterForward):
		model.characterRight()
	case key.Matches(msg, model.KeyMap.LineNext):
		model.CursorDown()
	case key.Matches(msg, model.KeyMap.WordForward):
		model.wordRight()
	case key.Matches(msg, model.KeyMap.Paste):
		return model, Paste
	case key.Matches(msg, model.KeyMap.CharacterBackward):
		model.characterLeft(false)
	case key.Matches(msg, model.KeyMap.LinePrevious):
		model.CursorUp()
	case key.Matches(msg, model.KeyMap.WordBackward):
		model.wordLeft()
	case key.Matches(msg, model.KeyMap.InputBegin):
		model.moveToBegin()
	case key.Matches(msg, model.KeyMap.InputEnd):
		model.moveToEnd()
	case key.Matches(msg, model.KeyMap.LowercaseWordForward):
		model.lowercaseRight()
	case key.Matches(msg, model.KeyMap.UppercaseWordForward):
		model.uppercaseRight()
	case key.Matches(msg, model.KeyMap.CapitalizeWordForward):
		model.capitalizeRight()
	case key.Matches(msg, model.KeyMap.TransposeCharacterBackward):
		model.transposeLeft()

	default:
		model.insertRunesFromUserInput(msg.Runes)
	}

	return model, nil
}
