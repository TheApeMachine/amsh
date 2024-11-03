package features

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/errnie"
)

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

type Browser struct {
	model    filepicker.Model
	selected string
	err      error
}

func NewBrowser() *Browser {
	fp := filepicker.New()
	fp.AllowedTypes = []string{} // Empty slice allows all file types
	fp.ShowHidden = true         // Show hidden files
	fp.Height = 20               // Set a reasonable height for the picker

	return &Browser{
		model: fp,
	}
}

func (browser *Browser) Init() tea.Cmd {
	// Set the current directory to start from
	currentDir := errnie.SafeMust(func() (string, error) {
		return os.Getwd()
	})

	browser.model.CurrentDirectory = currentDir
	return browser.model.Init()
}

func (browser *Browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return browser, tea.Quit
		}
	case clearErrorMsg:
		browser.err = nil
	}

	// Handle file picker updates
	var cmd tea.Cmd
	browser.model, cmd = browser.model.Update(msg)

	// Did the user select a disabled file?
	if didSelect, path := browser.model.DidSelectDisabledFile(msg); didSelect {
		browser.err = errors.New(path + " is not valid.")
		browser.selected = ""
		return browser, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return browser, cmd
}

func (browser *Browser) View() string {
	var s strings.Builder
	s.WriteString("\n  ")
	if browser.err != nil {
		s.WriteString(browser.model.Styles.DisabledFile.Render(browser.err.Error()))
	} else if browser.selected == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + browser.model.Styles.Selected.Render(browser.selected))
	}
	s.WriteString("\n\n" + browser.model.View() + "\n")
	return s.String()
}
