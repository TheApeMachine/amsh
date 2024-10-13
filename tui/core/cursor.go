// File: core/cursor.go

package core

import (
	"fmt"

	"github.com/theapemachine/amsh/errnie"
)

type Cursor struct {
	X     int
	Y     int
	queue *Queue
	maxY  int
}

func NewCursor(queue *Queue) *Cursor {
	errnie.Trace()
	cursor := &Cursor{
		X:     1,
		Y:     1,
		queue: queue,
	}

	// Start event handling
	go cursor.run()

	return cursor
}

func (cursor *Cursor) SetMaxY(maxY int) {
	errnie.Trace()
	cursor.maxY = maxY
}

func (cursor *Cursor) run() {
	errnie.Trace()
	cursorSub := cursor.queue.Subscribe("cursor_event")
	for artifact := range cursorSub {
		switch artifact.Peek("scope") {
		case "left":
			cursor.MoveLeft(1)
		case "right":
			cursor.MoveRight(1)
		case "up":
			cursor.MoveUp(1)
		case "down":
			cursor.MoveDown(1)
		default:
			errnie.Warn("Unknown cursor event scope: %s", artifact.Peek("scope"))
		}
	}
}

func (cursor *Cursor) Move(x, y int) {
	errnie.Trace()
	if x < 1 {
		x = 1
	}
	if y < 1 {
		y = 1
	}
	if cursor.maxY > 0 && y > cursor.maxY {
		y = cursor.maxY
	}
	cursor.X = x
	cursor.Y = y
	fmt.Printf("\033[%d;%dH", cursor.Y, cursor.X)
}

func (cursor *Cursor) MoveLeft(n int) {
	errnie.Trace()
	cursor.Move(cursor.X-n, cursor.Y)
}

func (cursor *Cursor) MoveRight(n int) {
	errnie.Trace()
	cursor.Move(cursor.X+n, cursor.Y)
}

func (cursor *Cursor) MoveUp(n int) {
	errnie.Trace()
	cursor.Move(cursor.X, cursor.Y-n)
}

func (cursor *Cursor) MoveDown(n int) {
	errnie.Trace()
	cursor.Move(cursor.X, cursor.Y+n)
}
