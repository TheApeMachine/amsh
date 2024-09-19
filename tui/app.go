package tui

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

/*
App wraps the entire application so we can have a guarantee that we can always
restore the terminal to cooked mode.
*/
type App struct {
	oldState *term.State
	err      error
	width    int
	height   int
}

/*
New creates a new App.
*/
func New() *App {
	return &App{}
}

/*
Initialize sets up the application, and the recovery mechanisms.
*/
func (app *App) Initialize() *App {
	// Handle OS signals to ensure terminal restoration.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for range sigChan {
			app.flipMode()
		}
	}()

	// Flip to raw mode at initialization.
	app.flipMode()
	app.width, app.height, _ = term.GetSize(int(os.Stdout.Fd()))

	return app
}

/*
Run starts the main loop, which includes panic recovery.
*/
func (app *App) Run() *App {
	// Recover from panics and restore the terminal.
	defer func() {
		if r := recover(); r != nil {
			app.flipMode()
			os.Exit(1)
		}
	}()

	// Main application loop or logic here.
	// Placeholder for your main application logic
	for {
		
	}

	return app
}

func (app *App) Error() string {
	return app.err.Error()
}

/*
flipMode flips the terminal mode between raw and cooked.
*/
func (app *App) flipMode() (err error) {
	if app.oldState == nil {
		// Switch to raw mode
		if app.oldState, app.err = term.MakeRaw(int(os.Stdin.Fd())); app.err != nil {
			os.Exit(1)
		}

		return
	}

	// Restore terminal to cooked mode
	term.Restore(int(os.Stdin.Fd()), app.oldState)
	app.oldState = nil
	syscall.Write(int(os.Stdout.Fd()), []byte("terminal restored\n"))
	os.Exit(1)

	return
}
