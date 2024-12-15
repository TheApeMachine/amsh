package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/theapemachine/amsh/tui/types"
	"github.com/theapemachine/errnie"
)

// draw renders the current state to the screen
func (app *App) draw() {
	app.screen.Clear()
	width, height := app.screen.Size()

	if app.mode == types.BrowserMode {
		app.drawBrowser(width, height-1)
	} else if app.mode == types.ChatMode {
		app.drawChat(width, height-1)
	} else {
		splitAt := width / 2
		app.drawBuffer(app.leftPane, 0, splitAt, height)
		app.drawBuffer(app.rightPane, splitAt, width, height)
	}

	// Draw status line
	app.drawStatus(height - 1)
	app.screen.Show()
}

// drawBuffer renders a buffer to the screen
func (app *App) drawBuffer(buf *types.Buffer, startX, endX, height int) {
	if buf == nil {
		return
	}

	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	activeStyle := style.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	selectedStyle := style.Reverse(true)
	jumpStyle := style.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow).Bold(true)

	// Get jump points if in jump mode
	jumpPoints := make(map[string]types.Position)
	if buf.IsJumpMode() {
		jumpPoints = buf.GetJumpPoints()
	}

	for y := 0; y < height-1; y++ {
		lineNum := y + buf.GetScroll()
		line := buf.GetLine(lineNum)
		if line != nil {
			x := startX
			for xOffset, ch := range line {
				if x < endX {
					useStyle := style
					if buf == app.activePane {
						useStyle = activeStyle
					}
					// Check if this position is selected
					if buf.IsPositionSelected(xOffset, lineNum) {
						useStyle = selectedStyle
					}

					// Check if this position has a jump label
					if buf.IsJumpMode() {
						found := false
						for label, pos := range jumpPoints {
							if pos.X == xOffset && pos.Y == lineNum {
								// Draw jump label
								for i, labelCh := range []rune(label) {
									if x+i < endX {
										app.screen.SetContent(x+i, y, labelCh, nil, jumpStyle)
									}
								}
								x += len(label)
								found = true
								break
							}
						}
						if found {
							continue
						}
					}

					app.screen.SetContent(x, y, ch, nil, useStyle)
					x++
				}
			}
		}
		// Fill the rest of the line with background color
		for x := startX + len(line); x < endX; x++ {
			useStyle := style
			if buf == app.activePane {
				useStyle = activeStyle
			}
			// Check if this position is selected
			if buf.IsPositionSelected(x-startX, lineNum) {
				useStyle = selectedStyle
			}
			app.screen.SetContent(x, y, ' ', nil, useStyle)
		}
	}

	// Draw cursor if this is the active pane and not in jump mode
	if buf == app.activePane && !buf.IsJumpMode() {
		cursorX, cursorY := buf.GetCursor()
		app.screen.ShowCursor(startX+cursorX, cursorY)
	}
}

// drawStatus draws the status line
func (app *App) drawStatus(y int) {
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite).Reverse(true)
	errorStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorRed).Reverse(true)
	jumpStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow).Bold(true)

	var status string
	var useStyle tcell.Style

	if app.lastError != nil {
		// Show error
		status = "Error: " + app.lastError.Error()
		useStyle = errorStyle
	} else if app.activePane.IsJumpMode() {
		// Show jump mode status
		status = "Jump mode - input: " + app.activePane.GetJumpInput()
		useStyle = jumpStyle
	} else if app.mode == types.Command {
		// Show command input
		status = ":" + app.commandBuffer
		useStyle = style
	} else {
		// Show normal status
		status = app.mode.String()
		if app.activePane.GetFilename() != "" {
			status += " | " + app.activePane.GetFilename()
		}
		useStyle = style
	}

	// Clear the status line first
	width, _ := app.screen.Size()
	for i := 0; i < width; i++ {
		app.screen.SetContent(i, y, ' ', nil, useStyle)
	}

	// Draw the status text
	for i, ch := range []rune(status) {
		if i < width {
			app.screen.SetContent(i, y, ch, nil, useStyle)
		}
	}
}

// drawBrowser renders the file browser
func (app *App) drawBrowser(width, height int) {
	errnie.Log("Drawing browser, width: %d, height: %d", width, height)
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	selectedStyle := style.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite).Reverse(true)
	dirStyle := style.Foreground(tcell.ColorWhite)
	selectedDirStyle := selectedStyle
	headerStyle := style.Reverse(true)

	// Draw current path
	path := "Location: " + app.browser.GetCurrentPath()
	errnie.Log("Browser current path: %s", path)
	for i := 0; i < width; i++ {
		if i < len(path) {
			app.screen.SetContent(i, 0, []rune(path)[i], nil, headerStyle)
		} else {
			app.screen.SetContent(i, 0, ' ', nil, headerStyle)
		}
	}

	// Draw entries
	entries := app.browser.GetEntries()
	selected := app.browser.GetSelected()
	scroll := app.browser.GetScroll()
	errnie.Log("Browser entries: %d, selected: %d, scroll: %d", len(entries), selected, scroll)

	for i := 0; i < height-1; i++ {
		y := i + 1 // Start after the path line
		idx := i + scroll

		if idx >= len(entries) {
			// Fill empty lines with background
			for x := 0; x < width; x++ {
				app.screen.SetContent(x, y, ' ', nil, style)
			}
			continue
		}

		entry := entries[idx]
		name := entry.Name
		if entry.IsDir {
			name += "/"
		}

		// Choose style based on selection and type
		useStyle := style
		if idx == selected {
			if entry.IsDir {
				useStyle = selectedDirStyle
			} else {
				useStyle = selectedStyle
			}
		} else if entry.IsDir {
			useStyle = dirStyle
		}

		// Draw the entry
		for x := 0; x < width; x++ {
			if x < len(name) {
				app.screen.SetContent(x, y, []rune(name)[x], nil, useStyle)
			} else {
				app.screen.SetContent(x, y, ' ', nil, useStyle)
			}
		}
	}
	errnie.Log("Browser drawing complete")
}

// drawChat renders the chat interface
func (app *App) drawChat(width, height int) {
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	aiStyle := style.Foreground(tcell.ColorGreen)
	humanStyle := style.Foreground(tcell.ColorBlue)
	inputStyle := style.Reverse(true)

	// Draw messages
	messages := app.chat.GetMessages()
	y := height - 3 // Leave space for input
	for i := len(messages) - 1; i >= 0 && y >= 0; i-- {
		msg := messages[i]
		useStyle := humanStyle
		if msg.From == "ai" {
			useStyle = aiStyle
		}
		for x, ch := range []rune(msg.Content) {
			if x < width {
				app.screen.SetContent(x, y, ch, nil, useStyle)
			}
		}
		y--
	}

	// Draw input area
	for x, ch := range []rune(app.chatInput) {
		if x < width {
			app.screen.SetContent(x, height-1, ch, nil, inputStyle)
		}
	}
	app.screen.ShowCursor(len(app.chatInput), height-1)
}
