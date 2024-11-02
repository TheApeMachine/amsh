// File: tui/app.go

package tui

import (
	"time"

	"github.com/containerd/console"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tui/core"
)

type App struct {
	buffer *core.Buffer
	cursor *core.Cursor
	screen *core.Screen
	err    chan error
}

// New creates a new application.
func New() *App {
	return &App{
		buffer: core.NewBuffer(),
		cursor: core.NewCursor(),
		screen: core.NewScreen(),
		err:    make(chan error),
	}
}

// Run starts the main event loop.
func (app *App) Run() chan error {
	current := console.Current()
	defer current.Reset()
	defer app.screen.Fini() // Ensure the screen is finalized

	errnie.MustVoid(current.SetRaw())
	current.Resize(
		errnie.SafeMust(func() (console.WinSize, error) {
			return current.Size()
		}),
	)

	keyboard := core.NewKeyboard()

	go func() {
		defer close(app.err)

		for {
			select {
			case msg := <-keyboard.Pipe():
				app.handleKey(msg)
				app.screen.Render(app.buffer, app.cursor)
			case <-app.err:
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	return app.err
}

// handleKey processes key messages and updates the buffer and cursor.
func (app *App) handleKey(msg core.KeyMsg) {
	line, col := app.cursor.GetPosition()

	switch msg.Cmd {
	case core.CmdTypeModeInsert:
		app.buffer.Insert(line, col, msg.Key)
		app.cursor.MoveRight(app.buffer)
	case core.CmdTypeModeNormal:
		// Handle normal mode commands
	case core.CmdTypeModeVisual:
		// Handle visual mode commands
	case core.CmdTypeQuit:
		app.err <- nil
	}
}
