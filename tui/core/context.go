package core

type Context struct {
	Queue    *Queue
	Keyboard *Keyboard
	Cursor   *Cursor
	Buffers  []*Buffer
	Height   int
	Width    int
}

func NewContext(queue *Queue, width, height int) *Context {
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

// SetMode sets the current mode and ensures all components are updated
func (ctx *Context) SetMode(mode Mode) {
	if ctx.Keyboard.mode != nil {
		ctx.Keyboard.mode.Exit()
	}
	ctx.Keyboard.mode = mode
	mode.Enter(ctx)
}
