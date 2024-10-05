package core

import (
	"fmt"

	"github.com/theapemachine/amsh/data"
)

type Context struct {
	Queue   *Queue
	Buffers []*Buffer
	Cursor  *Cursor
	Width   int
	Height  int
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
