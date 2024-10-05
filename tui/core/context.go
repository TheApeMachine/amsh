package core

import (
	"fmt"

	"github.com/theapemachine/amsh/data"
)

type Context struct {
	Mode     Mode
	Queue    *Queue
	Buffers  []*Buffer
	Keyboard *Keyboard
	Cursor   *Cursor
	Width    int
	Height   int
}

func NewContext(queue *Queue) *Context {
	return &Context{
		Mode:     &Normal{},
		Queue:    queue,
		Buffers:  make([]*Buffer, 0),
		Keyboard: NewKeyboard(queue),
		Cursor:   NewCursor(queue),
		Width:    80,
		Height:   24,
	}
}

func (ctx *Context) SetMode(mode Mode) {
	ctx.Mode = mode
}

func (ctx *Context) Run() {
	ctx.Mode.Run()
}

// UpdateBuffer updates the content of the buffer at a given line and column.
func (c *Context) UpdateBuffer(bufferIndex int, line int, col int, content []rune) {
	if bufferIndex < 0 || bufferIndex >= len(c.Buffers) {
		fmt.Println("Buffer index out of range")
		return
	}

	buffer := c.Buffers[bufferIndex]
	if line < 0 || line >= len(buffer.Data) {
		fmt.Println("Line index out of range")
		return
	}

	if col < 0 || col >= len(buffer.Data[line]) {
		fmt.Println("Column index out of range")
		return
	}

	buffer.Data[line] = append(buffer.Data[line][:col], append(content, buffer.Data[line][col:]...)...)
	buffer.renderFromLine(line) // Re-render the buffer from the modified line

	// Publish an event to the queue indicating that the buffer has been updated
	artifact := data.New("buffer", "event", "buffer_update", []byte(fmt.Sprintf("Buffer %d updated at line %d, col %d", bufferIndex, line, col)))
	c.Queue.Publish("buffer_updated", artifact)
}

// PublishEvent provides a convenient way to publish events to the queue
func (c *Context) PublishEvent(topic string, artifact *data.Artifact) {
	c.Queue.Publish(topic, artifact)
}
