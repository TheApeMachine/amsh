package filebrowser

import (
	"errors"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/logger"
)

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	EndSection := logger.StartSection("filebrowser.Update", "update")
	defer EndSection()

	logger.Debug("<- %T", msg)

	switch msg := msg.(type) {
	case clearErrorMsg:
		model.err = nil
	default:
		_ = msg
	}

	var cmd tea.Cmd
	model.filepicker, cmd = model.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := model.filepicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		model.selectedFile = path
	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := model.filepicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		model.err = errors.New(path + " is not valid.")
		model.selectedFile = ""
		return model, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return model, cmd
}
