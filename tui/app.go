package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tui/core"
)

/*
App is the main application struct. It manages the screens and the layout.
Everything else is handled by the features and flows through the command
Update pipeline.
*/
type App struct {
	*core.Manager
}

/*
NewApp initializes the application and sets up the default screens.
*/
func NewApp() *App {
	errnie.Info("Initializing application")

	return &App{
		Manager: core.NewManager(),
	}
}

/*
Init initializes the application.
*/
func (app *App) Init() tea.Cmd {
	errnie.Info("Initializing application")
	return app.Manager.Init()
}

/*
Update handles the application's update logic.
*/
func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return app.Manager.Update(msg)
}

/*
View returns the application's view.
*/
func (app *App) View() string {
	return app.Manager.View()
}
