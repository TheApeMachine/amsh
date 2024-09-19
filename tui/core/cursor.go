package core

import (
	"fmt" // Added fmt for sending ANSI escape codes
)

/*
Cursor is a management structure that handles the cursor in terminal raw mode.
*/
type Cursor struct {
	x int
	y int
}

func NewCursor() *Cursor {
	return &Cursor{0, 0}
}

/*
Move moves the cursor to the given position.
*/
func (cursor *Cursor) Move(x, y int) {
	// Move cursor to (x, y) using ANSI escape codes
	fmt.Printf("\033[%d;%dH", y, x)
	cursor.x = x
	cursor.y = y
}

/*
Position returns the current position of the cursor.
*/
func (cursor *Cursor) Position() (x, y int) {
	return cursor.x, cursor.y
}
