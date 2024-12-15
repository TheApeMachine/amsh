package commands

import (
	"fmt"
	"os"

	"github.com/theapemachine/errnie"
)

// EditorCommands holds the command implementations
type EditorCommands struct {
	quit     func() error
	write    func(filename string) error
	openFile func(filename string) error
}

// NewEditorCommands creates a new set of editor commands
func NewEditorCommands(quit func() error, write func(filename string) error, openFile func(filename string) error) *EditorCommands {
	return &EditorCommands{
		quit:     quit,
		write:    write,
		openFile: openFile,
	}
}

// RegisterBasicCommands registers the basic editor commands
func (ec *EditorCommands) RegisterBasicCommands(registry *Registry) {
	// Quit command
	registry.Register(Command{
		Name:        "q",
		Description: "Quit the editor",
		Execute: func(args []string) error {
			return ec.quit()
		},
	})

	// Write command
	registry.Register(Command{
		Name:        "w",
		Description: "Write buffer to file",
		Execute: func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("filename required")
			}
			return ec.write(args[0])
		},
	})

	// Write and quit command
	registry.Register(Command{
		Name:        "wq",
		Description: "Write buffer to file and quit",
		Execute: func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("filename required")
			}
			if err := ec.write(args[0]); err != nil {
				return err
			}
			return ec.quit()
		},
	})

	// File explorer command
	registry.Register(Command{
		Name:        "ex",
		Description: "Open file explorer",
		Execute: func(args []string) error {
			errnie.Log("Executing ex command")
			return ec.openFile("")
		},
	})

	// Edit file command
	registry.Register(Command{
		Name:        "e",
		Description: "Edit file",
		Execute: func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("filename required")
			}
			// Check if file exists
			if _, err := os.Stat(args[0]); os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", args[0])
			}
			return ec.openFile(args[0])
		},
	})

	// Help command
	registry.Register(Command{
		Name:        "help",
		Description: "Show available commands",
		Execute: func(args []string) error {
			// TODO: Implement help display
			return nil
		},
	})
}
