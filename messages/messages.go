package messages

import tea "github.com/charmbracelet/bubbletea"

// Define message types within the messages package
type SetFilenameMsg string
type FileSelectedMsg string
type OpenFileBrowserMsg struct{}
type SetActiveComponentMsg string

// Mode represents the buffer mode
type Mode int

const (
	NormalMode Mode = iota
	InsertMode
)

/*
StatusUpdateMsg is a message that is sent to update the status bar.
*/
type StatusUpdateMsg struct {
	Filename string
	Mode     Mode
}

/*
ComponentMsg is a message that is sent to a component in the buffer.
*/
type ComponentMsg struct {
	ComponentName string
	InnerMsg      tea.Msg
}
