package tui

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tui/commands"
	"github.com/theapemachine/amsh/tui/types"
)

// App represents the main application state
type App struct {
	screen        tcell.Screen
	leftPane      *types.Buffer
	rightPane     *types.Buffer
	activePane    *types.Buffer
	mode          types.Mode
	running       bool
	commandBuffer string
	cmdRegistry   *commands.Registry
	lastError     error // For displaying command errors
	browser       *types.Browser
	chat          *types.Chat
	chatInput     string
	chatHistory   []string
	chatCursor    int // For navigating chat history
}

// NewApp creates a new application instance
func NewApp() *App {
	// Initialize logging
	os.Setenv("LOGFILE", "true")
	os.Setenv("NOCONSOLE", "true")
	errnie.InitLogger()

	screen := errnie.SafeMust(func() (tcell.Screen, error) {
		return tcell.NewScreen()
	})

	// Initialize screen and handle any error
	errnie.SafeMust(func() (bool, error) {
		return true, screen.Init()
	})

	app := &App{
		screen:        screen,
		mode:          types.Normal,
		running:       true,
		commandBuffer: "",
		cmdRegistry:   commands.NewRegistry(),
		browser:       types.NewBrowser(),
		chat:          types.NewChat(),
		chatInput:     "",
		chatHistory:   make([]string, 0),
		chatCursor:    -1,
	}

	// Initialize with empty buffers
	app.leftPane = types.NewBuffer()
	app.rightPane = types.NewBuffer()
	app.activePane = app.leftPane

	// Initialize commands
	cmds := commands.NewEditorCommands(
		app.quit,
		app.writeFile,
		app.openFile,
	)
	cmds.RegisterBasicCommands(app.cmdRegistry)

	return app
}

// Command implementations
func (app *App) quit() error {
	app.running = false
	return nil
}

func (app *App) writeFile(filename string) error {
	// If no filename is provided and buffer has a filename, use that
	if filename == "" {
		filename = app.activePane.GetFilename()
		if filename == "" {
			return fmt.Errorf("no filename specified")
		}
	}

	// Create or truncate the file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write each line to the file
	for i := 0; i < app.activePane.LineCount(); i++ {
		line := app.activePane.GetLine(i)
		if i > 0 {
			// Add newline before each line except the first
			if _, err := file.WriteString("\n"); err != nil {
				return fmt.Errorf("failed to write newline: %w", err)
			}
		}
		if _, err := file.WriteString(string(line)); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	// Update the buffer's filename
	app.activePane.SetFilename(filename)
	return nil
}

func (app *App) openFile(filename string) error {
	// If no filename provided, enter browser mode
	if filename == "" {
		errnie.Log("Opening file browser")
		app.mode = types.BrowserMode
		err := app.browser.Refresh()
		if err != nil {
			errnie.Log("Error refreshing browser: %v", err)
			return err
		}
		errnie.Log("Browser mode activated")
		return nil
	}
	errnie.Log("Opening file: %s", filename)
	return app.activePane.LoadFile(filename)
}

// Run starts the main application loop
func (app *App) Run() error {
	defer app.screen.Fini()
	errnie.Log("Starting application main loop")

	for app.running {
		errnie.Log("Main loop iteration - mode: %s", app.mode.String())
		app.draw()
		app.handleEvent()
	}

	errnie.Log("Application loop ended, running = %v", app.running)
	return nil
}

// executeCommand executes the current command buffer
func (app *App) executeCommand() {
	errnie.Log("Executing command: %s (current mode: %s)", app.commandBuffer, app.mode.String())
	if err := app.cmdRegistry.Execute(app.commandBuffer); err != nil {
		errnie.Log("Command error: %v", err)
		app.lastError = err
	}
	errnie.Log("Command executed, new mode: %s", app.mode.String())
}

// SendChatMessage sends a message in chat mode
func (app *App) SendChatMessage() {
	if app.chatInput == "" {
		return
	}

	// Create context from current state
	context := &types.MessageContext{
		FilePath: app.activePane.GetFilename(),
	}

	// If there's selected text, include it
	if app.mode == types.Visual {
		context.Selection = app.activePane.GetSelectedText()
	}

	// Add message to chat
	app.chat.AddMessage(app.chatInput, "human", context)

	// Add to history
	app.chatHistory = append(app.chatHistory, app.chatInput)
	app.chatCursor = len(app.chatHistory)

	// Clear input
	app.chatInput = ""
}
