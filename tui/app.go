// tui/app.go

package tui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/theapemachine/amsh/tui/core"
	"golang.org/x/term"
)

type App struct {
	width    int
	height   int
	oldState *term.State
	queue    *core.Queue
	context  *core.Context
	running  bool
}

func New() *App {
	return &App{
		context: core.NewContext(
			core.NewQueue(100),
		),
	}
}

func (app *App) Initialize() *App {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigChan
		app.cleanupAndExit()
	}()

	app.flipMode() // Switch to raw mode.
	app.clearScreen()

	// Get terminal size
	app.width, app.height, _ = term.GetSize(int(os.Stdout.Fd()))
	app.running = true
	return app
}

func (app *App) Run() {
	defer app.flipMode() // Ensure terminal is restored when exiting

	for app.running {
		app.context.Run()
	}
}

// flipMode toggles the terminal between raw and cooked mode.
func (app *App) flipMode() {
	var err error
	if app.oldState == nil {
		if app.oldState, err = term.MakeRaw(int(os.Stdin.Fd())); err != nil {
			fmt.Println("Error entering raw mode:", err)
			os.Exit(1)
		}
		return
	}

	term.Restore(int(os.Stdin.Fd()), app.oldState)
	app.oldState = nil
	fmt.Println("\nTerminal restored")
}

// cleanupAndExit gracefully exits the application.
func (app *App) cleanupAndExit() {
	app.running = false
	app.flipMode()
	fmt.Println("Exiting...")
	os.Exit(0)
}

// clearScreen clears the terminal screen using ANSI escape codes.
func (app *App) clearScreen() {
	fmt.Print("\033[H\033[2J")
	os.Stdout.Sync() // Flush output
}
