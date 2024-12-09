package commands

import (
	"fmt"
	"strings"
)

// Command represents a command that can be executed
type Command struct {
	Name        string
	Args        []string
	Description string
	Execute     func(args []string) error
}

// ParseCommand parses a command string into name and arguments
func ParseCommand(input string) (name string, args []string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

// CommandError represents an error that occurred during command execution
type CommandError struct {
	Command string
	Message string
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("Command '%s': %s", e.Command, e.Message)
}

// Registry holds all available commands
type Registry struct {
	commands map[string]Command
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a new command to the registry
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name] = cmd
}

// Get returns a command by name
func (r *Registry) Get(name string) (Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// Execute executes a command string
func (r *Registry) Execute(input string) error {
	name, args := ParseCommand(input)
	if name == "" {
		return &CommandError{Command: input, Message: "empty command"}
	}

	cmd, ok := r.Get(name)
	if !ok {
		return &CommandError{Command: name, Message: "unknown command"}
	}

	return cmd.Execute(args)
}
