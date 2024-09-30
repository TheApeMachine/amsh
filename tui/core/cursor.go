package core

import "fmt"

type Cursor struct {
	X     int
	Y     int
	queue *Queue
}

func NewCursor(queue *Queue) *Cursor {
	return &Cursor{
		X:     1,
		Y:     1,
		queue: queue,
	}
}

func (cursor *Cursor) Move(x, y int) {
	if x < 1 {
		x = 1
	}
	if y < 1 {
		y = 1
	}
	cursor.X = x
	cursor.Y = y
	fmt.Printf("\033[%d;%dH", y, x)
}

func (cursor *Cursor) MoveForward(n int, maxX int) {
	cursor.X += n
	if cursor.X > maxX {
		cursor.X = maxX
	}
	fmt.Printf("\033[%d;%dH", cursor.Y, cursor.X)
}

func (cursor *Cursor) MoveBackward(n int) {
	cursor.X -= n
	if cursor.X < 1 {
		cursor.X = 1
	}
	fmt.Printf("\033[%d;%dH", cursor.Y, cursor.X)
}

func (cursor *Cursor) MoveUp(n int) {
	cursor.Y -= n
	if cursor.Y < 1 {
		cursor.Y = 1
	}
	fmt.Printf("\033[%d;%dH", cursor.Y, cursor.X)
}

func (cursor *Cursor) MoveDown(n int, maxY int) {
	cursor.Y += n
	if cursor.Y > maxY {
		cursor.Y = maxY
	}
	fmt.Printf("\033[%d;%dH", cursor.Y, cursor.X)
}
