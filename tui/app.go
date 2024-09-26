package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/tui/core"
	"github.com/theapemachine/amsh/twoface"
	"golang.org/x/term"
)

/*
App wraps the entire application to ensure the terminal is restored to cooked mode upon exit.
*/
type App struct {
	pool     *twoface.Pool
	oldState *term.State
	keyboard *core.Keyboard
	err      chan error
	width    int
	height   int
}

/*
New creates a new App.
*/
func New() *App {
	return &App{
		pool:     twoface.NewPool(context.Background(), 10),
		keyboard: core.NewKeyboard(),
		err:      make(chan error),
	}
}

/*
Initialize sets up the application and recovery mechanisms.
*/
func (app *App) Initialize() *App {
	fmt.Println("Viper Configurations:")
	fmt.Println(viper.AllSettings())

	// Handle OS signals to ensure terminal restoration.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigChan
		app.flipMode()
		os.Exit(0)
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

	var err error

	if _, err = io.Copy(app.pool, app.keyboard); err != nil {
		app.err <- err
		os.Exit(1)
	}

	return app
}

func (app *App) Error() string {
	err := <-app.err
	return err.Error()
}

/*
flipMode toggles the terminal between raw and cooked mode.
*/
func (app *App) flipMode() (err error) {
	if app.oldState == nil {
		// Switch to raw mode
		if app.oldState, err = term.MakeRaw(int(os.Stdin.Fd())); err != nil {
			app.err <- err
			os.Exit(1)
		}
		return
	}

	// Restore terminal to cooked mode
	term.Restore(int(os.Stdin.Fd()), app.oldState)
	app.oldState = nil
	fmt.Fprintln(os.Stdout, "terminal restored")
	return
}
