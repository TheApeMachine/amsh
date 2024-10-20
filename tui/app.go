// File: tui/app.go

package tui

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tui/core"
	"golang.org/x/term"
)

type App struct {
	width    int
	height   int
	oldState *term.State
	context  *core.Context
	running  bool
	wg       *sync.WaitGroup
}

// New creates a new application.
func New() *App {
	return &App{
		wg: &sync.WaitGroup{},
	}
}

// Initialize sets up the application.
func (app *App) Initialize() *App {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigChan
		app.cleanupAndExit()
	}()

	app.flipMode() // Switch to raw mode.
	app.clearScreen()

	app.running = true

	app.width, app.height, _ = term.GetSize(int(os.Stdout.Fd()))
	app.context = core.NewContext(
		core.NewQueue(),
		app.width,
		app.height,
	)

	// Subscribe to app events
	go app.handleAppEvents()

	return app
}

// Run starts the main event loop.
func (app *App) Run() {
	errnie.Trace()
	app.wg.Add(1)
	go app.readLoop()
	app.wg.Wait()
}

// handleAppEvents listens for events on the 'app_event' topic.
func (app *App) handleAppEvents() {
	errnie.Trace()
	appSub := app.context.Queue.Subscribe("app_event")
	for artifact := range appSub {
		switch artifact.Peek("scope") {
		case "Quit":
			app.cleanupAndExit()
		default:
			errnie.Warn("Unknown app event scope: %s", artifact.Peek("scope"))
		}
	}
}

// readLoop continuously reads input from the user.
func (app *App) readLoop() {
	errnie.Trace()
	for app.running {
		b := readByte()
		if b != 0 {
			app.context.Keyboard.HandleInput(b)
		}
	}
}

// flipMode toggles the terminal between raw and cooked mode.
func (app *App) flipMode() {
	errnie.Trace()
	var err error
	if app.oldState == nil {
		if app.oldState, err = term.GetState(int(os.Stdin.Fd())); err != nil {
			fmt.Println("Error getting terminal state:", err)
			os.Exit(1)
		}

		// Make a copy of the old state to modify
		rawState := *app.oldState

		// Apply the raw mode attributes
		if err = term.Restore(int(os.Stdin.Fd()), &rawState); err != nil {
			fmt.Println("Error setting raw mode:", err)
			os.Exit(1)
		}

		return
	}

	term.Restore(int(os.Stdin.Fd()), app.oldState)
	app.oldState = nil
	// Show cursor upon exit
	fmt.Print("\033[?25h")
	fmt.Println("\nTerminal restored")
}

// cleanupAndExit gracefully exits the application.
func (app *App) cleanupAndExit() {
	errnie.Trace()
	app.running = false
	app.flipMode()
	fmt.Println("Exiting...")
	app.wg.Done()
}

// clearScreen clears the terminal screen using ANSI escape codes.
func (app *App) clearScreen() {
	errnie.Trace()
	fmt.Print("\033[H\033[2J")
	os.Stdout.Sync() // Flush output
}

// readByte reads a single byte from standard input.
func readByte() byte {
	errnie.Trace()
	var buf [1]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		errnie.Error(err)
		return 0
	}
	if n == 0 {
		return 0
	}
	return buf[0]
}
