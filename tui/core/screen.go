package core

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

type Screen struct {
	screen tcell.Screen
}

// NewScreen initializes a new screen.
func NewScreen() *Screen {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Error creating screen: %v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("Error initializing screen: %v", err)
	}

	// Set default style
	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack))

	s.Clear()

	return &Screen{screen: s}
}

// Render draws the buffer and cursor on the screen.
func (s *Screen) Render(buffer *Buffer, cursor *Cursor, mode Mode, browser BrowserInterface) {
	s.screen.Clear()

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	// Get screen dimensions
	_, height := s.screen.Size()

	// If browser is visible, render it on the left side
	if browser != nil && browser.IsVisible() {
		browserWidth := 30
		entries := browser.GetCurrentView()
		selected := browser.GetSelected()

		// Draw browser border
		for y := 0; y < height; y++ {
			s.screen.SetContent(browserWidth, y, '│', nil, style)
		}

		// Draw browser entries
		for i, entry := range entries {
			entryStyle := style
			if i == selected {
				entryStyle = style.Background(tcell.ColorDarkBlue)
			}
			for j, r := range entry {
				if j < browserWidth {
					s.screen.SetContent(j, i, r, nil, entryStyle)
				}
			}
		}

		// Adjust buffer rendering to account for browser width
		bufferOffset := browserWidth + 1
		for i := 0; i < buffer.LineCount(); i++ {
			line := buffer.GetLine(i)
			for j, r := range line {
				s.screen.SetContent(bufferOffset+j, i, r, nil, style)
			}
		}
	} else {
		// Original buffer rendering
		for i := 0; i < buffer.LineCount(); i++ {
			line := buffer.GetLine(i)
			for j, r := range line {
				s.screen.SetContent(j, i, r, nil, style)
			}
		}
	}

	// Position the cursor
	line, col := cursor.GetPosition()
	s.screen.ShowCursor(col, line)

	// Show mode in the last line
	modeText := "-- NORMAL --"
	switch mode {
	case ModeInsert:
		modeText = "-- INSERT --"
	case ModeVisual:
		modeText = "-- VISUAL --"
	}

	// Use existing height variable
	for i, r := range modeText {
		s.screen.SetContent(i, height-1, r, nil, style)
	}

	s.screen.Show()
}

// Fini finalizes the screen and restores the terminal state.
func (s *Screen) Fini() {
	if s.screen != nil {
		s.screen.Fini()
	}
}
