package core

import (
	"context"
	"fmt"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

// Mode interface defines the behavior for different modes
type Mode interface {
	Enter(context *Context)
	Exit()
	Run()
}

// Normal mode implementation
type Normal struct {
	context      *Context
	bufferEvents <-chan *data.Artifact
	cancel       context.CancelFunc
}

func (n *Normal) Enter(ctx *Context) {
	errnie.Trace()
	n.context = ctx

	// Subscribe to buffer events.
	n.bufferEvents = ctx.Queue.Subscribe("buffer_event")

	// Create a cancellable context for handling buffer events.
	var cancel context.CancelFunc
	ctxBuf, cancel := context.WithCancel(context.Background())
	n.cancel = cancel

	go n.handleBufferEvents(ctxBuf)

	// Show the cursor when entering Normal mode
	fmt.Print("\033[?25h")
}

func (n *Normal) Exit() {
	errnie.Trace()

	if n.bufferEvents != nil {
		n.context.Queue.Unsubscribe("buffer_event", n.bufferEvents)
	}

	// Cancel the buffer events handling goroutine.
	if n.cancel != nil {
		n.cancel()
	}
}

func (n *Normal) Run() {
	// Implement Normal mode logic, such as handling user input.
	// This could involve reading from keyboard input and processing commands.
	// Currently, it's handled elsewhere; ensure this method is utilized if needed.
}

func (n *Normal) handleBufferEvents(ctx context.Context) {
	for {
		select {
		case event, ok := <-n.bufferEvents:
			if !ok {
				return
			}
			n.processEvent(event)
		case <-ctx.Done():
			return
		}
	}
}

func (n *Normal) processEvent(event *data.Artifact) {
	t, err := event.Type()
	if err != nil {
		errnie.Error(err)
		return
	}

	payload, err := event.Payload()
	if err != nil {
		errnie.Error(err)
		return
	}

	if t != "" {
		switch t {
		case "buffer_update":
			fmt.Printf("[Normal Mode] Buffer updated: %s\n", string(payload))
			// Optionally, trigger a re-render or handle buffer updates.
		case "render_line":
			fmt.Printf("[Normal Mode] Line rendered: %s\n", string(payload))
			// Handle rendering if needed.
		case "show_status":
			fmt.Printf("[Normal Mode] Status message: %s\n", string(payload))
		case "render_from_line":
			fmt.Printf("[Normal Mode] Render from line: %s\n", string(payload))
		default:
			errnie.Warn("Unknown event type in Normal mode: %s", t)
		}
	}
}

// Insert mode implementation
type Insert struct {
	context      *Context
	bufferEvents <-chan *data.Artifact
	cancel       context.CancelFunc
}

func (i *Insert) Enter(ctx *Context) {
	errnie.Trace()
	i.context = ctx

	// Subscribe to buffer events.
	i.bufferEvents = ctx.Queue.Subscribe("buffer_event")

	// Create a cancellable context for handling buffer events.
	var cancel context.CancelFunc
	ctxBuf, cancel := context.WithCancel(context.Background())
	i.cancel = cancel

	go i.handleBufferEvents(ctxBuf)

	// Optionally, you can perform actions specific to entering Insert mode,
	// such as updating status messages.
}

func (i *Insert) Exit() {
	errnie.Trace()

	if i.bufferEvents != nil {
		i.context.Queue.Unsubscribe("buffer_event", i.bufferEvents)
	}

	// Cancel the buffer events handling goroutine.
	if i.cancel != nil {
		i.cancel()
	}

	// Optionally, perform actions specific to exiting Insert mode.
}

func (i *Insert) Run() {
	// Implement Insert mode logic, such as capturing user input for text insertion.
	// Currently, it's handled elsewhere via event handling.
}

func (i *Insert) handleBufferEvents(ctx context.Context) {
	for {
		select {
		case event, ok := <-i.bufferEvents:
			if !ok {
				return
			}
			i.processEvent(event)
		case <-ctx.Done():
			return
		}
	}
}

func (i *Insert) processEvent(event *data.Artifact) {
	t, err := event.Type()
	if err != nil {
		errnie.Error(err)
		return
	}

	payload, err := event.Payload()
	if err != nil {
		errnie.Error(err)
		return
	}

	if t != "" {
		switch t {
		case "insert_rune":
			fmt.Printf("[Insert Mode] Inserted rune: %s\n", string(payload))
			// Handle rune insertion logic if needed.
		case "render_line":
			fmt.Printf("[Insert Mode] Line rendered: %s\n", string(payload))
			// Handle rendering if needed.
		case "show_status":
			fmt.Printf("[Insert Mode] Status message: %s\n", string(payload))
		case "render_from_line":
			fmt.Printf("[Insert Mode] Render from line: %s\n", string(payload))
		default:
			errnie.Warn("Unknown event type in Insert mode: %s", t)
		}
	}
}

// Command mode implementation
type Command struct {
	context     *Context
	commandSub  <-chan *data.Artifact
	cancel      context.CancelFunc
	commandLine string
}

func (c *Command) Enter(ctx *Context) {
	errnie.Trace()
	c.context = ctx
	c.commandLine = ""

	// Subscribe to command input events.
	c.commandSub = ctx.Queue.Subscribe("command_input")

	// Create a cancellable context for handling command input.
	cmdCtx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	// Move cursor to bottom left and display ':' prompt
	c.context.Cursor.Move(1, c.context.Height)
	fmt.Print("\033[K:")                       // Clear the line and print ':'
	c.context.Cursor.Move(2, c.context.Height) // Position cursor after ':'

	// Hide the cursor in Command Mode
	fmt.Print("\033[?25l")

	// Start handling command input
	go c.handleCommandInput(cmdCtx)
}

func (c *Command) Exit() {
	errnie.Trace()

	// Cancel the command input handling goroutine
	if c.cancel != nil {
		c.cancel()
	}

	// Optionally, clear the command prompt
	fmt.Print("\033[K")
	c.context.Cursor.Move(c.context.Cursor.X, c.context.Cursor.Y-1) // Move cursor to previous line

	// Show the cursor upon exiting Command Mode
	fmt.Print("\033[?25h")
}

func (c *Command) Run() {
	// Implement Command mode logic, such as executing user commands.
	// Currently, it's handled via event subscription.
}

func (c *Command) handleCommandInput(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case artifact := <-c.commandSub:
			role, err := artifact.Role()
			if err != nil {
				errnie.Error(err)
				continue
			}
			switch role {
			case "UpdateCommandInput":
				payload, err := artifact.Payload()
				if err != nil {
					errnie.Error(err)
					continue
				}
				c.commandLine += string(payload)
				c.displayCommandInput(c.commandLine)
			case "BackspaceCommandInput":
				if len(c.commandLine) > 0 {
					c.commandLine = c.commandLine[:len(c.commandLine)-1]
					c.displayCommandInput(c.commandLine)
				}
			case "SubmitCommandInput":
				c.executeCommand(c.commandLine)
				c.context.SetMode(&Normal{}) // Switch back to Normal mode
				return
			}
		}
	}
}

func (c *Command) displayCommandInput(cmd string) {
	// Format the command prompt
	status := fmt.Sprintf(":%s", cmd)

	// Use the Buffer's ShowStatus method to display the command
	// Assuming you have access to the active Buffer
	if len(c.context.Buffers) > 0 {
		activeBuffer := c.context.Buffers[0]
		activeBuffer.ShowStatus(status)
	}

	// Move cursor to after the prompt
	c.context.Cursor.Move(len(cmd)+2, c.context.Height)
}

func (c *Command) executeCommand(cmd string) {
	// Process the entered command here
	// For example, handle "quit" to exit the application
	switch cmd {
	case "quit", "q":
		c.context.Queue.Publish("app_event", data.New("app_event", "quit", "", nil))
	default:
		// Handle other commands or show an error message
		activeBuffer := c.context.Buffers[0]
		activeBuffer.ShowStatus(fmt.Sprintf("Unknown command: %s", cmd))
	}
}
