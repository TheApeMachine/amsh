package tui

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tui/core"
	"golang.org/x/term"
)

type App struct {
	cursor    *core.Cursor
	oldState  *term.State
	width     int
	height    int
	mode      core.Mode // Changed from int to core.Mode
	running   bool      // Control loop flag
	buffers   []*core.Buffer
	bufPtr    int
	filename  string      // Current file name
	command   string      // Current command input
	statusMsg string      // Current status message
	queue     *core.Queue // Message queue
}

func New() *App {
	return &App{
		buffers: make([]*core.Buffer, 0),
	}
}

func (app *App) Initialize() *App {
	// Handle OS signals to ensure terminal restoration.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigChan
		app.cleanupAndExit()
	}()

	// Switch to raw mode.
	app.flipMode()

	// Clear the screen after switching to raw mode.
	app.clearScreen()

	// Get terminal size.
	app.width, app.height, _ = term.GetSize(int(os.Stdout.Fd()))

	// Initialize the message queue.
	app.queue = core.NewQueue(100)

	// Create a new cursor with the queue.
	app.cursor = core.NewCursor(app.queue)

	// Create a new buffer with the queue.
	app.buffers = append(app.buffers, core.NewBuffer(app.height, app.cursor, app.queue))
	app.bufPtr = 0

	// Initialize cursor position and render buffer.
	app.buffers[app.bufPtr].Cursor.Move(1, 1)
	app.buffers[app.bufPtr].Render()

	// Subscribe to various topics.
	cursorSub := app.queue.Subscribe("cursor")
	go app.handleEvents(cursorSub)

	bufferSub := app.queue.Subscribe("buffer")
	go app.handleEvents(bufferSub)

	modeSub := app.queue.Subscribe("mode_change")
	go app.handleEvents(modeSub)

	appQuitSub := app.queue.Subscribe("app")
	go app.handleEvents(appQuitSub)

	commandInputSub := app.queue.Subscribe("command_input")
	go app.handleCommandInput(commandInputSub)

	// Initialize the Keyboard and start its input loop.
	keyboard := core.NewKeyboard(app.queue)
	go func() {
		for app.running {
			keyboard.ReadInput()
		}
	}()

	app.running = true
	return app
}

func (app *App) Run() {
	defer app.flipMode() // Ensure the terminal is restored when the program exits.

	// Main loop for running the application.
	for app.running {
		// The application relies on event-driven architecture.
		select {}
	}
}

// handleEvents processes messages from the queue.
func (app *App) handleEvents(ch <-chan *data.Artifact) {
	var (
		role    string
		payload []byte
		err     error
	)

	for artifact := range ch {
		if role, err = artifact.Role(); err != nil {
			errnie.Error(err)
		}

		if payload, err = artifact.Payload(); err != nil {
			errnie.Error(err)
		}

		switch role {
		case "MoveUp":
			app.cursor.MoveUp(1)
		case "MoveDown":
			app.cursor.MoveDown(1, app.height)
		case "MoveForward":
			app.cursor.MoveForward(1, app.width)
		case "MoveBackward":
			app.cursor.MoveBackward(1)
		case "InsertChar":
			if len(payload) > 0 {
				ch := rune(payload[0])
				app.buffers[app.bufPtr].InsertChar(ch)
			}
		case "DeleteChar":
			app.buffers[app.bufPtr].DeleteCharUnderCursor()
		case "Backspace":
			app.buffers[app.bufPtr].HandleBackspace()
		case "Enter":
			app.buffers[app.bufPtr].HandleEnter()
		case "NormalMode":
			app.mode = core.NormalMode
			app.clearStatusLine()
			app.showCursorShape("block")
			// Publish mode change event.
			app.queue.Publish("mode_change", data.New("App", "ModeChange", "NormalMode", nil))
		case "InsertMode":
			app.mode = core.InsertMode
			app.showStatus("-- INSERT --")
			app.showCursorShape("bar")
			app.queue.Publish("mode_change", data.New("App", "ModeChange", "InsertMode", nil))
		case "CommandMode":
			app.mode = core.CommandMode
			app.command = ""
			app.showStatus(":")
			app.queue.Publish("mode_change", data.New("App", "ModeChange", "CommandMode", nil))
		case "Quit":
			app.cleanupAndExit()
		}
	}
}

// handleCommandInput processes command mode inputs.
func (app *App) handleCommandInput(ch <-chan *data.Artifact) {
	var (
		payload []byte
		err     error
	)
	for artifact := range ch {
		if payload, err = artifact.Payload(); err != nil {
			errnie.Error(err)
		}

		if len(payload) == 0 {
			continue
		}

		commandChar := payload[0]
		switch commandChar {
		case 13: // Enter key
			command := strings.TrimSpace(app.command)
			app.executeCommand(command)
			app.mode = core.NormalMode
			app.clearStatusLine()
			app.showCursorShape("block")
			// Publish mode change event.
			app.queue.Publish("mode_change", data.New("App", "ModeChange", "NormalMode", nil))
		case 27: // Escape key
			app.mode = core.NormalMode
			app.clearStatusLine()
			app.showCursorShape("block")
			app.queue.Publish("mode_change", data.New("App", "ModeChange", "NormalMode", nil))
		default:
			app.command += string(commandChar)
			app.showStatus(":" + app.command)
		}
	}
}

// executeCommand parses and executes the given command.
func (app *App) executeCommand(command string) {
	switch {
	case command == "q":
		app.cleanupAndExit()
	case command == "w":
		if app.filename == "" {
			app.showStatus("No file name. Use :w filename")
		} else {
			app.saveToFile(app.filename)
		}
	case strings.HasPrefix(command, "w "):
		filename := strings.TrimSpace(command[2:])
		app.saveToFile(filename)
	case strings.HasPrefix(command, "e "):
		filename := strings.TrimSpace(command[2:])
		app.openFile(filename)
	default:
		app.showStatus("Unknown command: " + command)
	}
}

// openFile opens a file and loads its contents into the buffer.
func (app *App) openFile(filename string) {
	dataBytes, err := os.ReadFile(filename)
	if err != nil {
		app.statusMsg = "Error opening file: " + err.Error()
		app.showStatus(":" + app.statusMsg)
		return
	}

	lines := strings.Split(string(dataBytes), "\n")
	buffer := app.buffers[app.bufPtr]
	buffer.Data = make([][]rune, len(lines))
	for i, line := range lines {
		buffer.Data[i] = []rune(line)
	}

	buffer.Filename = filename
	buffer.Cursor.Move(1, 1)
	buffer.Render()
	buffer.StatusMsg = "Opened file: " + filename
	buffer.ShowStatus(buffer.StatusMsg)
}

// saveToFile saves the buffer contents to the given file.
func (app *App) saveToFile(filename string) {
	var sb strings.Builder
	buffer := app.buffers[app.bufPtr]
	for _, line := range buffer.Data {
		sb.WriteString(string(line))
		sb.WriteString("\n")
	}

	err := os.WriteFile(filename, []byte(sb.String()), 0644)
	if err != nil {
		buffer.StatusMsg = "Error saving file: " + err.Error()
	} else {
		buffer.StatusMsg = "File saved: " + filename
	}

	buffer.Filename = filename
	buffer.ShowStatus(buffer.StatusMsg)
}

// showStatus prints a status message at the bottom of the screen.
func (app *App) showStatus(status string) {
	fmt.Printf("\033[%d;1H", app.height)
	fmt.Print("\033[K") // Clear the status line
	fmt.Print(status)
	app.cursor.Move(app.cursor.X, app.cursor.Y)
	flushStdout()
}

// clearStatusLine clears the status line without affecting the text buffer.
func (app *App) clearStatusLine() {
	// Move to the status line and clear it.
	fmt.Printf("\033[%d;1H\033[K", app.height)
	// Move cursor back to the current position.
	fmt.Printf("\033[%d;%dH", app.cursor.Y, app.cursor.X)
	flushStdout()
}

// showCursorShape changes the cursor shape based on the mode.
func (app *App) showCursorShape(shape string) {
	switch shape {
	case "block":
		fmt.Print("\033[1 q") // Block cursor
	case "bar":
		fmt.Print("\033[5 q") // Bar cursor
	}
	flushStdout()
}

// flipMode toggles the terminal between raw and cooked mode.
func (app *App) flipMode() {
	var err error
	if app.oldState == nil {
		// Switch to raw mode.
		if app.oldState, err = term.MakeRaw(int(os.Stdin.Fd())); err != nil {
			fmt.Println("Error entering raw mode:", err)
			os.Exit(1)
		}
		return
	}

	// Restore terminal to cooked mode.
	term.Restore(int(os.Stdin.Fd()), app.oldState)
	app.oldState = nil
	fmt.Println("\nTerminal restored")
}

// cleanupAndExit handles the graceful exit of the application.
func (app *App) cleanupAndExit() {
	app.running = false // Stop the main loop
	app.flipMode()      // Restore the terminal
	fmt.Println("Exiting...")
	os.Exit(0) // Exit the program
}

// clearScreen clears the terminal screen using ANSI escape codes.
func (app *App) clearScreen() {
	fmt.Print("\033[H\033[2J") // Clear screen and move cursor to top-left corner
	flushStdout()              // Ensure output is flushed
}

// flushStdout ensures that the stdout buffer is properly flushed.
func flushStdout() {
	os.Stdout.Sync() // Ensure any buffered output is flushed
}
