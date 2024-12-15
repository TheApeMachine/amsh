package features

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/errnie"
)

type clearErrorMsg struct{}
type FocusEditorMsg struct{}
type FileSelectedMsg struct {
	Path string
}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

type Browser struct {
	width    int
	model    filepicker.Model
	selected string
	err      error
}

func NewBrowser() *Browser {
	currentDir := errnie.SafeMust(func() (string, error) {
		return os.Getwd()
	})

	fp := filepicker.New()
	fp.ShowHidden = true
	fp.Height = 10
	fp.CurrentDirectory = currentDir

	// Enable directory listing
	fp.DirAllowed = true
	fp.FileAllowed = true

	return &Browser{
		model: fp,
	}
}

func (browser *Browser) Model() tea.Model {
	return browser
}

func (browser *Browser) Name() string {
	return "browser"
}

func (browser *Browser) Size() (int, int) {
	return browser.width, browser.model.Height
}

func (browser *Browser) Init() tea.Cmd {
	return browser.model.Init()
}

func (browser *Browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return browser, tea.Quit
		}
	case tea.WindowSizeMsg:
		// Enforce a strict maximum height of 50 lines to ensure we never exceed layout bounds
		// This leaves room for header, borders, and padding
		maxHeight := 20
		desiredHeight := msg.Height - 8 // Even more conservative padding

		if desiredHeight > maxHeight {
			browser.model.Height = maxHeight
		} else if desiredHeight < 3 {
			browser.model.Height = 3 // Minimum viable height
		} else {
			browser.model.Height = desiredHeight
		}
	case clearErrorMsg:
		browser.err = nil
	}

	// Handle file picker updates
	var cmd tea.Cmd
	browser.model, cmd = browser.model.Update(msg)

	// Did the user select a file?
	if didSelect, path := browser.model.DidSelectFile(msg); didSelect {
		fileInfo, err := os.Stat(path)
		if err == nil && !fileInfo.IsDir() {
			browser.selected = path
			return browser, tea.Batch(cmd, func() tea.Msg {
				return FileSelectedMsg{Path: path}
			})
		} else if fileInfo != nil && fileInfo.IsDir() {
			browser.err = errors.New("cannot open directories")
			return browser, tea.Batch(cmd, clearErrorAfter(2*time.Second))
		}
	}

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
	s.WriteString("\n")
	s.WriteString("  ")
	if browser.err != nil {
		s.WriteString(browser.model.Styles.DisabledFile.Render(browser.err.Error()))
	} else if browser.selected == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + browser.model.Styles.Selected.Render(browser.selected))
	}
	s.WriteString("\n" + browser.model.View())
	return s.String()
}

func (browser *Browser) Selected() string {
	return browser.selected
}
