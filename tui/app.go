package tui

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/tui/features"
)

const (
	width         = 96
	height        = 24
	columnWidth   = 30
	initialInputs = 2
	maxInputs     = 6
	minInputs     = 1
	helpHeight    = 5
)

type App struct {
	width    int
	height   int
	screens  map[string][]tea.Model
	screen   string
	textarea *features.TextArea // Keep track of the textarea model
}

func NewApp() *App {
	textarea := features.NewTextarea()
	return &App{
		screens: map[string][]tea.Model{
			"splash":  {features.NewSplash(width, height)},
			"editor":  {textarea},
			"browser": {features.NewBrowser()},
		},
		screen:   "splash",
		textarea: textarea, // Store the textarea model
	}
}

func (app *App) Init() tea.Cmd {
	var cmds []tea.Cmd

	for _, model := range app.screens[app.screen] {
		cmds = append(cmds, model.Init())
	}

	return tea.Batch(cmds...)
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Update current screen's models and collect their commands
	for _, model := range app.screens[app.screen] {
		var cmd tea.Cmd
		_, cmd = model.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return app, tea.Quit
		case "ctrl+b":
			app.screen = "browser"
			// Initialize the browser when switching to it
			cmds = append(cmds, app.screens["browser"][0].Init())
		case "ctrl+e":
			app.screen = "editor"
			// Initialize the editor and set focus when switching to it
			cmds = append(cmds, app.screens["editor"][0].Init(), app.textarea.Focus())
		}
	case tea.WindowSizeMsg:
		app.width, app.height = msg.Width, msg.Height
		// Propagate the window size message to the current screen's models
		for _, models := range app.screens {
			for _, model := range models {
				_, cmd := model.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	// Check if a file was selected in the browser
	if app.screen == "browser" {
		browser := app.screens["browser"][0].(*features.Browser)
		if browser.Selected() != "" {
			// Load the file content into the textarea
			content, err := os.ReadFile(browser.Selected())
			if err != nil {
				log.Printf("Error reading file: %v", err)
			} else {
				app.textarea.SetValue(string(content))
				app.screen = "editor"                     // Switch to the editor
				cmds = append(cmds, app.textarea.Focus()) // Set focus on the textarea
			}
		}
	}

	return app, tea.Batch(cmds...)
}

func (app *App) View() string {
	return app.screens[app.screen][0].View()
}
