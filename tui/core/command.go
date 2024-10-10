// File: core/command.go

package core

import (
	"fmt"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

// Command mode implementation
type Command struct {
	context     *Context
	commandSub  <-chan *data.Artifact
	cancel      chan struct{}
	commandLine string
}

func (c *Command) Enter(ctx *Context) {
	errnie.Trace()
	c.context = ctx
	c.commandLine = ""
	c.cancel = make(chan struct{})

	// Subscribe to command input events.
	c.commandSub = ctx.Queue.Subscribe("command_input")

	// Move cursor to bottom left and display ':' prompt
	c.context.Cursor.Move(1, c.context.Height)
	fmt.Print("\033[K:")                       // Clear the line and print ':'
	c.context.Cursor.Move(2, c.context.Height) // Position cursor after ':'

	// Hide the cursor in Command Mode
	fmt.Print("\033[?25l")

	// Start handling command input
	go c.handleCommandInput()
}

func (c *Command) Exit() {
	errnie.Trace()

	// Cancel the command input handling goroutine
	close(c.cancel)

	// Unsubscribe from command input events
	c.context.Queue.Unsubscribe("command_input", c.commandSub)

	// Clear the command prompt
	fmt.Print("\033[K")
	c.context.Cursor.Move(c.context.Cursor.X, c.context.Cursor.Y-1) // Move cursor to previous line

	// Show the cursor upon exiting Command Mode
	fmt.Print("\033[?25h")
}

func (c *Command) handleCommandInput() {
	errnie.Trace()
	for {
		select {
		case <-c.cancel:
			return
		case artifact := <-c.commandSub:
			switch artifact.Peek("scope") {
			case "UpdateCommandInput":
				c.commandLine += string(artifact.Peek("payload"))
				c.displayCommandInput(c.commandLine)
			case "BackspaceCommandInput":
				if len(c.commandLine) > 0 {
					c.commandLine = c.commandLine[:len(c.commandLine)-1]
					c.displayCommandInput(c.commandLine)
				}
			case "SubmitCommandInput":
				c.executeCommand(c.commandLine)
				// Switch back to Normal mode
				artifact := data.New("CommandMode", "mode_change", "NormalMode", nil)
				c.context.Queue.Publish("mode_change", artifact)
				return
			}
		}
	}
}

func (c *Command) displayCommandInput(cmd string) {
	errnie.Trace()
	// Save current cursor position
	fmt.Print("\033[s")
	// Move cursor to command line and display command
	fmt.Printf("\033[%d;1H\033[K:%s", c.context.Height, cmd)
	// Restore cursor position
	fmt.Print("\033[u")
	c.context.Cursor.Move(len(cmd)+2, c.context.Height)
}

func (c *Command) executeCommand(cmd string) {
	errnie.Trace()
	// Process the entered command here
	switch cmd {
	case "quit", "q":
		artifact := data.New("CommandMode", "app_event", "Quit", nil)
		c.context.Queue.Publish("app_event", artifact)
	default:
		// Handle other commands or show an error message
		activeBuffer := c.context.Buffers[0]
		activeBuffer.ShowStatus(fmt.Sprintf("Unknown command: %s", cmd))
	}
}
