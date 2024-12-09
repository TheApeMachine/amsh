package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tui/types"
)

// handleEvent processes keyboard input
func (app *App) handleEvent() {
	ev := app.screen.PollEvent()
	switch ev := ev.(type) {
	case *tcell.EventKey:
		app.handleKeyEvent(ev)
	case *tcell.EventResize:
		app.screen.Sync()
	}
}

// handleKeyEvent processes keyboard events based on current mode
func (app *App) handleKeyEvent(ev *tcell.EventKey) {
	switch app.mode {
	case types.Normal:
		app.handleNormalMode(ev)
	case types.Insert:
		app.handleInsertMode(ev)
	case types.Command:
		app.handleCommandMode(ev)
	case types.Visual:
		app.handleVisualMode(ev)
	case types.BrowserMode:
		app.handleBrowserMode(ev)
	case types.ChatMode:
		app.handleChatMode(ev)
	}
}

// handleNormalMode handles keys in normal mode
func (app *App) handleNormalMode(ev *tcell.EventKey) {
	// If in jump mode, handle jump input
	if app.activePane.IsJumpMode() {
		switch ev.Key() {
		case tcell.KeyEscape:
			app.activePane.ExitJumpMode()
		case tcell.KeyRune:
			errnie.Log("Jump mode input: %c", ev.Rune())
			if app.activePane.HandleJumpModeInput(ev.Rune()) {
				errnie.Log("Jumped using label: %s", app.activePane.GetJumpInput())
			}
		}
		return // Important: don't process any other keys in jump mode
	}

	switch ev.Key() {
	case tcell.KeyEscape:
		app.running = false
	case tcell.KeyUp:
		app.activePane.MoveCursorUp()
	case tcell.KeyDown:
		app.activePane.MoveCursorDown()
	case tcell.KeyLeft:
		app.activePane.MoveCursorLeft()
	case tcell.KeyRight:
		app.activePane.MoveCursorRight()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q':
			app.running = false
		case 'i':
			app.mode = types.Insert
		case ':':
			app.mode = types.Command
			app.commandBuffer = ""
		case 'v':
			app.mode = types.Visual
			app.activePane.StartSelection()
		case 'f':
			// Start jump mode
			errnie.Log("Starting jump mode")
			app.activePane.StartJumpMode()
		case 'h':
			app.activePane.MoveCursorLeft()
		case 'j':
			app.activePane.MoveCursorDown()
		case 'k':
			app.activePane.MoveCursorUp()
		case 'l':
			app.activePane.MoveCursorRight()
		case 'p':
			// Paste after cursor
			app.activePane.MoveCursorRight()
			app.activePane.Paste()
		case 'P':
			// Paste before cursor
			app.activePane.Paste()
		case '\t':
			if app.activePane == app.leftPane {
				app.activePane = app.rightPane
			} else {
				app.activePane = app.leftPane
			}
		case 'c':
			app.mode = types.ChatMode
			app.chatInput = ""
			app.chatCursor = len(app.chatHistory)
		}
	}
}

// handleInsertMode handles keys in insert mode
func (app *App) handleInsertMode(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		app.mode = types.Normal
	case tcell.KeyEnter:
		app.activePane.InsertNewline()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		app.activePane.Backspace()
	case tcell.KeyRune:
		app.activePane.InsertRune(ev.Rune())
	}
}

// handleCommandMode handles keys in command mode
func (app *App) handleCommandMode(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		app.mode = types.Normal
		app.commandBuffer = ""
	case tcell.KeyEnter:
		errnie.Log("Executing command and keeping current mode")
		app.executeCommand()
		// Only switch back to normal mode if we're not entering browser mode
		if app.mode != types.BrowserMode {
			app.mode = types.Normal
		}
		app.commandBuffer = ""
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(app.commandBuffer) > 0 {
			app.commandBuffer = app.commandBuffer[:len(app.commandBuffer)-1]
		}
	case tcell.KeyRune:
		app.commandBuffer += string(ev.Rune())
	}
}

// handleVisualMode handles keys in visual mode
func (app *App) handleVisualMode(ev *tcell.EventKey) {
	errnie.Log("Visual mode key: %v", ev.Key())
	switch ev.Key() {
	case tcell.KeyEscape:
		errnie.Log("Visual mode: ESC pressed, clearing selection and returning to normal mode")
		app.activePane.ClearSelection()
		app.mode = types.Normal
	case tcell.KeyUp:
		app.activePane.MoveCursorUp()
		app.activePane.UpdateSelection()
	case tcell.KeyDown:
		app.activePane.MoveCursorDown()
		app.activePane.UpdateSelection()
	case tcell.KeyLeft:
		app.activePane.MoveCursorLeft()
		app.activePane.UpdateSelection()
	case tcell.KeyRight:
		app.activePane.MoveCursorRight()
		app.activePane.UpdateSelection()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'h':
			app.activePane.MoveCursorLeft()
			app.activePane.UpdateSelection()
		case 'j':
			app.activePane.MoveCursorDown()
			app.activePane.UpdateSelection()
		case 'k':
			app.activePane.MoveCursorUp()
			app.activePane.UpdateSelection()
		case 'l':
			app.activePane.MoveCursorRight()
			app.activePane.UpdateSelection()
		case 'v':
			// Exit visual mode
			app.activePane.ClearSelection()
			app.mode = types.Normal
		case 'y':
			// Yank selection to clipboard
			text := app.activePane.Yank()
			errnie.Log("Yanked text: %s", text)
			app.activePane.ClearSelection()
			app.mode = types.Normal
		case 'd':
			// Delete selection (yank and clear)
			text := app.activePane.Yank()
			errnie.Log("Cut text: %s", text)
			// TODO: Implement delete selection
			app.activePane.ClearSelection()
			app.mode = types.Normal
		}
	}
}

// handleBrowserMode handles keys in browser mode
func (app *App) handleBrowserMode(ev *tcell.EventKey) {
	errnie.Log("Browser mode handling key: %v", ev.Key())
	switch ev.Key() {
	case tcell.KeyEscape:
		errnie.Log("Browser mode: ESC pressed, returning to normal mode")
		app.mode = types.Normal
	case tcell.KeyEnter:
		errnie.Log("Browser mode: ENTER pressed")
		if path, isDir, err := app.browser.Enter(); err != nil {
			errnie.Log("Browser mode: Error on Enter: %v", err)
			app.lastError = err
		} else if !isDir && path != "" {
			errnie.Log("Browser mode: Selected file: %s", path)
			if err := app.openFile(path); err != nil {
				errnie.Log("Browser mode: Error opening file: %v", err)
				app.lastError = err
			} else {
				errnie.Log("Browser mode: Successfully opened file, returning to normal mode")
				app.mode = types.Normal
			}
		} else {
			errnie.Log("Browser mode: Entered directory: %s", path)
		}
	case tcell.KeyUp:
		errnie.Log("Browser mode: Up arrow pressed")
		app.browser.MoveUp()
	case tcell.KeyDown:
		errnie.Log("Browser mode: Down arrow pressed")
		app.browser.MoveDown()
	case tcell.KeyLeft:
		errnie.Log("Browser mode: Left arrow pressed - going to parent directory")
		if path, isDir, err := app.browser.Enter(); err != nil {
			app.lastError = err
		} else if isDir && path == ".." {
			// Already handled by Enter()
		}
	case tcell.KeyRight:
		errnie.Log("Browser mode: Right arrow pressed - entering directory/file")
		if path, isDir, err := app.browser.Enter(); err != nil {
			app.lastError = err
		} else if !isDir && path != "" {
			if err := app.openFile(path); err != nil {
				app.lastError = err
			} else {
				app.mode = types.Normal
			}
		}
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'j':
			errnie.Log("Browser mode: Moving down")
			app.browser.MoveDown()
		case 'k':
			errnie.Log("Browser mode: Moving up")
			app.browser.MoveUp()
		case 'h':
			errnie.Log("Browser mode: Moving to parent directory")
			if path, isDir, err := app.browser.Enter(); err != nil {
				app.lastError = err
			} else if isDir && path == ".." {
				// Already handled by Enter()
			}
		case 'l':
			errnie.Log("Browser mode: Entering directory/file")
			if path, isDir, err := app.browser.Enter(); err != nil {
				app.lastError = err
			} else if !isDir && path != "" {
				if err := app.openFile(path); err != nil {
					app.lastError = err
				} else {
					app.mode = types.Normal
				}
			}
		case 'q':
			errnie.Log("Browser mode: 'q' pressed, returning to normal mode")
			app.mode = types.Normal
		default:
			errnie.Log("Browser mode: Unhandled rune: %c", ev.Rune())
		}
	default:
		errnie.Log("Browser mode: Unhandled key: %v", ev.Key())
	}
}

// handleChatMode handles keys in chat mode
func (app *App) handleChatMode(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		app.mode = types.Normal
		app.chatInput = ""
		app.chatCursor = -1
	case tcell.KeyEnter:
		app.SendChatMessage()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(app.chatInput) > 0 {
			app.chatInput = app.chatInput[:len(app.chatInput)-1]
		}
	case tcell.KeyUp:
		// Navigate chat history
		if app.chatCursor > 0 {
			app.chatCursor--
			app.chatInput = app.chatHistory[app.chatCursor]
		}
	case tcell.KeyDown:
		// Navigate chat history
		if app.chatCursor < len(app.chatHistory)-1 {
			app.chatCursor++
			app.chatInput = app.chatHistory[app.chatCursor]
		} else if app.chatCursor == len(app.chatHistory)-1 {
			app.chatCursor = len(app.chatHistory)
			app.chatInput = ""
		}
	case tcell.KeyRune:
		app.chatInput += string(ev.Rune())
	}
}
