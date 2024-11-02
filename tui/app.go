// File: tui/app.go

package tui

import (
	"github.com/containerd/console"
	"github.com/theapemachine/amsh/tui/core"
	"github.com/theapemachine/amsh/tui/features"
	"log"
)

type App struct {
	buffer   *core.Buffer
	cursor   *core.Cursor
	screen   *core.Screen
	err      chan error
	mode     core.Mode
	teleport *features.Teleport
	browser  *features.Browser
}

// New creates a new application.
func New() *App {
	return &App{
		buffer:   core.NewBuffer(),
		cursor:   core.NewCursor(),
		screen:   core.NewScreen(),
		err:      make(chan error),
		mode:     core.ModeNormal,
		teleport: features.NewTeleport(),
		browser:  features.NewBrowser(),
	}
}

// Run starts the main event loop.
func (app *App) Run() chan error {
	current := console.Current()

	// Set raw mode and handle cleanup
	if err := current.SetRaw(); err != nil {
		app.err <- err
		return app.err
	}

	// Initialize screen
	app.screen.Render(app.buffer, app.cursor, app.mode, app.browser)

	keyboard := core.NewKeyboard()
	keyChan := keyboard.Pipe()

	go func() {
		defer func() {
			// Restore terminal state
			current.Reset()
			app.screen.Fini()
			close(app.err)
		}()

		// Use for range to iterate over keyChan
		for msg := range keyChan {
			if msg.Cmd == core.CmdTypeQuit {
				app.err <- nil
				return
			}
			app.handleKey(msg)
		}
	}()

	return app.err
}

// handleKey processes key messages and updates the buffer and cursor.
func (app *App) handleKey(msg core.KeyMsg) {
	log.Printf("App received command: %v", msg.Cmd)

	// If teleport is active, handle its input first
	if app.teleport.IsActive() {
		if msg.Cmd == core.CmdTypeModeNormal {
			app.teleport.Toggle() // Deactivate on ESC
		} else if msg.Cmd == core.CmdTypeNone {
			if matched, target := app.teleport.AddInput(msg.Key); matched {
				app.cursor.MoveTo(target.Line, target.Col)
				app.teleport.Toggle() // Deactivate after jump
			}
		}
		app.screen.Render(app.buffer, app.cursor, app.mode, app.browser)
		return
	}

	// Handle browser-specific navigation when browser is visible
	if app.browser.IsVisible() {
		switch msg.Cmd {
		case core.CmdMoveUp:
			app.browser.MoveUp()
			app.screen.Render(app.buffer, app.cursor, app.mode, app.browser)
			return
		case core.CmdMoveDown:
			app.browser.MoveDown()
			app.screen.Render(app.buffer, app.cursor, app.mode, app.browser)
			return
		case core.CmdMoveRight: // Enter directory or file
			if path, err := app.browser.Enter(); err == nil && path != "" {
				if lines, err := app.browser.LoadFile(path); err == nil {
					app.buffer = core.NewBuffer()
					app.buffer.Lines = lines
					app.cursor.MoveTo(0, 0)
				}
			}
			app.screen.Render(app.buffer, app.cursor, app.mode, app.browser)
			return
		case core.CmdMoveLeft: // Go back in directory
			app.browser.Back()
			app.screen.Render(app.buffer, app.cursor, app.mode, app.browser)
			return
		case core.CmdBrowserToggle:
			app.browser.Toggle()
			app.screen.Render(app.buffer, app.cursor, app.mode, app.browser)
			return
		}
	}

	// Get cursor position for editor commands
	line, col := app.cursor.GetPosition()

	// Handle regular editor commands when browser is not visible
	switch msg.Cmd {
	case core.CmdTypeModeInsert:
		app.mode = core.ModeInsert
	case core.CmdTypeModeNormal:
		app.mode = core.ModeNormal
	case core.CmdMoveLeft:
		app.cursor.MoveLeft()
	case core.CmdMoveRight:
		app.cursor.MoveRight(app.buffer)
	case core.CmdMoveUp:
		app.cursor.MoveUp()
	case core.CmdMoveDown:
		app.cursor.MoveDown(app.buffer)
	case core.CmdDeleteChar:
		app.buffer.Delete(line, col)
	case core.CmdTypeNone:
		if app.mode == core.ModeInsert {
			app.buffer.Insert(line, col, msg.Key)
			app.cursor.MoveRight(app.buffer)
		}
	case core.CmdBackspace:
		if app.mode == core.ModeInsert && col > 0 {
			app.cursor.MoveLeft()
			newLine, newCol := app.cursor.GetPosition()
			app.buffer.Delete(newLine, newCol)
		}
	case core.CmdNewLineBelow:
		app.buffer.InsertLine(line + 1)
		app.cursor.MoveTo(line+1, 0)
		app.mode = core.ModeInsert
	case core.CmdNewLineAbove:
		app.buffer.InsertLine(line)
		app.cursor.MoveTo(line, 0)
		app.mode = core.ModeInsert
	case core.CmdEnter:
		if app.mode == core.ModeInsert {
			app.buffer.SplitLine(line, col)
			app.cursor.MoveTo(line+1, 0)
		}
	case core.CmdTeleportMode:
		app.teleport.Toggle()
		if app.teleport.IsActive() {
			app.teleport.Analyze(app.buffer, line, col)
		}
	case core.CmdBrowserToggle:
		app.browser.Toggle()
	case core.CmdBrowserUp:
		if app.browser.IsVisible() {
			app.browser.MoveUp()
		} else if app.mode == core.ModeNormal {
			app.cursor.MoveUp()
		}
	case core.CmdBrowserDown:
		if app.browser.IsVisible() {
			app.browser.MoveDown()
		} else if app.mode == core.ModeNormal {
			app.cursor.MoveDown(app.buffer)
		}
	case core.CmdBrowserEnter:
		if app.browser.IsVisible() {
			if path, err := app.browser.Enter(); err == nil && path != "" {
				// Load the selected file
				if lines, err := app.browser.LoadFile(path); err == nil {
					app.buffer = core.NewBuffer()
					app.buffer.Lines = lines
					app.cursor.MoveTo(0, 0)
				}
			}
		}
	case core.CmdBrowserBack:
		if app.browser.IsVisible() {
			app.browser.Back()
		}
	}

	app.screen.Render(app.buffer, app.cursor, app.mode, app.browser)
}
