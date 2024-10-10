package tui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tui/core"
	"golang.org/x/term"
)

/*
App is the main application structure.
*/
type App struct {
	width    int
	height   int
	oldState *term.State
	context  *core.Context
	running  bool
}

// New creates a new application.
func New() *App {
	return &App{}
}

/*
Initialize sets up the application.
*/
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

	// Get terminal size
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

/*
Run starts the main event loop.
*/
func (app *App) Run() {
	defer app.flipMode() // Ensure terminal is restored when exiting

	// Start reading input
	go app.readLoop()

	// Keep the main goroutine alive until app is not running
	for app.running {
		time.Sleep(100 * time.Millisecond)
	}
}

/*
handleAppEvents listens for events on the 'app' topic.
*/
func (app *App) handleAppEvents() {
	appSub := app.context.Queue.Subscribe("app_event")
	for artifact := range appSub {
		role, err := artifact.Role()
		if err != nil {
			errnie.Error(err)
			continue
		}
		switch role {
		case "quit":
			app.cleanupAndExit()
		default:
			errnie.Warn("Unknown app event role: %s", role)
		}
	}
}

func (app *App) readLoop() {
	for app.running {
		b := readByte()
		if b != 0 {
			app.context.Keyboard.HandleInput(b)
		}
	}
}

/*
flipMode toggles the terminal between raw and cooked mode.
*/
func (app *App) flipMode() {
	var err error
	if app.oldState == nil {
		if app.oldState, err = term.MakeRaw(int(os.Stdin.Fd())); err != nil {
			fmt.Println("Error entering raw mode:", err)
			os.Exit(1)
		}
		// Ensure echo is disabled
		newState := *app.oldState
		term.Restore(int(os.Stdin.Fd()), &newState)

		// Hide cursor in raw mode if desired
		fmt.Print("\033[?25l") // Hide cursor
		return
	}

	term.Restore(int(os.Stdin.Fd()), app.oldState)
	app.oldState = nil
	// Show cursor upon exit
	fmt.Print("\033[?25h")
	fmt.Println("\nTerminal restored")
}

/*
cleanupAndExit gracefully exits the application.
*/
func (app *App) cleanupAndExit() {
	app.running = false
	app.flipMode()
	fmt.Println("Exiting...")
	os.Exit(0)
}

/*
clearScreen clears the terminal screen using ANSI escape codes.
*/
func (app *App) clearScreen() {
	fmt.Print("\033[H\033[2J")
	os.Stdout.Sync() // Flush output
}

// readByte reads a single byte from standard input.
func readByte() byte {
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
