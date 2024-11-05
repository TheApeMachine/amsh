package features

import (
	"errors"
	"fmt"
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
	currentDir := errnie.SafeMust(func() (string, error) {
		return os.Getwd()
	})

	fmt.Printf("Current directory: %s\n", currentDir)
	if files, err := os.ReadDir(currentDir); err == nil {
		for _, file := range files {
			fmt.Printf("File: %s\n", file.Name())
		}
	}

	fp := filepicker.New()
	fp.ShowHidden = true
	fp.Height = 20
	fp.CurrentDirectory = currentDir
	
	// Enable directory listing
	fp.DirAllowed = true
	fp.FileAllowed = true
	
	return &Browser{
		model: fp,
	}
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
		browser.model.Height = msg.Height - 2
	case clearErrorMsg:
		browser.err = nil
	}

	// Handle file picker updates
	var cmd tea.Cmd
	browser.model, cmd = browser.model.Update(msg)

	// Did the user select a file?
	if didSelect, path := browser.model.DidSelectFile(msg); didSelect {
		browser.selected = path
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

func (browser *Browser) Selected() string {
	return browser.selected
}
