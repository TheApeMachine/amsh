package core

import (
	"fmt"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

type Mode interface {
	Enter(context *Context)
	Exit()
	Run()
}

// Normal mode implementation
type Normal struct {
	context      *Context
	bufferEvents <-chan *data.Artifact
}

func (n *Normal) Enter(context *Context) {
	n.context = context
	n.bufferEvents = context.Queue.Subscribe("buffer_event") // Subscribe to buffer events

	go n.handleBufferEvents() // Start handling buffer events in a separate goroutine
}

func (n *Normal) Exit() {
	if n.bufferEvents != nil {
		n.context.Queue.Unsubscribe("buffer_event", n.bufferEvents)
	}
}

func (n *Normal) Run() {
	// Mode-specific behavior here, like handling user input
	// In the normal mode, the user could navigate the buffer, execute commands, etc.
}

func (n *Normal) handleBufferEvents() {
	var (
		t       string
		payload []byte
		err     error
	)

	for event := range n.bufferEvents {
		if t, err = event.Type(); err != nil {
			errnie.Error(err)
			return
		}

		payload, err = event.Payload()
		if err != nil {
			errnie.Error(err)
			return
		}

		if t != "" {
			switch t {
			case "insert_rune":
				fmt.Printf("[Normal Mode] Buffer Event: %s\\n", string(payload))
			case "render_line":
				fmt.Printf("[Normal Mode] Rendered a line: %s\\n", string(payload))
			case "show_status":
				fmt.Printf("[Normal Mode] Status updated: %s\\n", string(payload))
			case "render_from_line":
				fmt.Printf("[Normal Mode] Render from line: %s\\n", string(payload))
			default:
				fmt.Printf("[Normal Mode] Unknown event: %s\\n", event.Type)
			}
		}
	}
}

// Insert mode implementation
type Insert struct {
	context *Context
}

func (i *Insert) Enter(context *Context) {
	i.context = context
}

func (i *Insert) Exit() {
}

func (i *Insert) Run() {
}

func (i *Insert) handleBufferEvents() {
}

func (i *Insert) handleInsertEvents() {
}

// Command mode implementation
type Command struct {
	context *Context
}

func (c *Command) Enter(context *Context) {
	c.context = context
}

func (c *Command) Exit() {
}

func (c *Command) Run() {
}

func (c *Command) handleCommandEvents() {
}
