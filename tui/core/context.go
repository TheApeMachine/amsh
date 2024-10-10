// File: core/context.go

package core

import "github.com/theapemachine/amsh/errnie"

type Context struct {
	Queue    *Queue
	Keyboard *Keyboard
	Cursor   *Cursor
	Buffers  []*Buffer
	Height   int
	Width    int
}

// NewContext initializes the application context.
func NewContext(queue *Queue, width, height int) *Context {
	errnie.Trace()
	context := &Context{
		Queue:  queue,
		Width:  width,
		Height: height,
	}

	// Initialize the cursor
	context.Cursor = NewCursor(queue)

	// Initialize the keyboard and set its context
	context.Keyboard = NewKeyboard(queue)
	context.Keyboard.SetContext(context)

	// Initialize a buffer and add it to the context
	buffer := NewBuffer(height-1, context.Cursor, queue)
	context.Buffers = []*Buffer{buffer}

	// Set the initial mode to Normal
	context.SetMode(&Normal{})

	return context
}

// SetMode sets the current mode and ensures all components are updated.
func (ctx *Context) SetMode(mode Mode) {
	errnie.Trace()
	if ctx.Keyboard.mode != nil {
		ctx.Keyboard.mode.Exit()
	}
	ctx.Keyboard.mode = mode
	mode.Enter(ctx)
}

// Run can be used to trigger rendering or other periodic tasks.
func (ctx *Context) Run() {
	errnie.Trace()
	// Currently unused since rendering is event-driven.
}
