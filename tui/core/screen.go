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

	return &Screen{screen: s}
}

// Render draws the buffer and cursor on the screen.
func (s *Screen) Render(buffer *Buffer, cursor *Cursor) {
	s.screen.Clear()

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	// Render each line of the buffer
	for i := 0; i < buffer.LineCount(); i++ {
		line := buffer.GetLine(i)
		for j, r := range line {
			s.screen.SetContent(j, i, r, nil, style)
		}
	}

	// Position the cursor
	line, col := cursor.GetPosition()
	s.screen.ShowCursor(col, line)

	// Refresh the screen
	s.screen.Show()
}

// Fini finalizes the screen and restores the terminal state.
func (s *Screen) Fini() {
	s.screen.Fini()
}
