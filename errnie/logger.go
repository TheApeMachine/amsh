package errnie

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
)

var (
	styles    = createDefaultStyles()
	logFile   *os.File
	logFileMu sync.Mutex

	logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		CallerOffset:    3,
		ReportTimestamp: true,
		TimeFormat:      time.TimeOnly,
		Level:           log.DebugLevel,
	})
)

/*
Initialize logging system by configuring log styles, setting log levels,
and initializing log files if applicable.
*/
func init() {
	// Initialize the log file
	initLogFile()
	if logFile == nil {
		fmt.Println("WARNING: Log file initialization failed!")
	}

	// Configure logger styles
	logger.SetStyles(styles)

	// Set log level based on configuration
	setLogLevel()

	// // Periodic routine to print the number of active goroutines
	// go func() {
	// 	for range time.Tick(time.Second * 5) {
	// 		logger.Debug("active goroutines", "count", runtime.NumGoroutine())
	// 	}
	// }()
}

/*
Set the appropriate logging level from Viper configuration.
*/
func setLogLevel() {
	switch viper.GetString("loglevel") {
	case "trace", "debug":
		logger.SetLevel(log.DebugLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.DebugLevel)
	}
}

/*
makeStyle creates a Lip Gloss style for the log levels.

This helper function is used to reduce redundancy in style creation by specifying background and foreground colors.

Example usage:

	style := makeStyle("#F7746D", "#EEEEEE")
*/
func makeStyle(background, foreground string) lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color(background)).Foreground(lipgloss.Color(foreground))
}

/*
Creates and returns the default set of styles for logging levels.
*/
func createDefaultStyles() *log.Styles {
	styles := log.DefaultStyles()
	styles.Levels[log.ErrorLevel] = makeStyle("#F7746D", "#EEEEEE")
	styles.Levels[log.WarnLevel] = makeStyle("#F7B96D", "#EEEEEE")
	styles.Levels[log.InfoLevel] = makeStyle("#6E95F7", "#EEEEEE")
	styles.Levels[log.DebugLevel] = makeStyle("#999999", "#EEEEEE")
	return styles
}

/*
Initialize the log file by creating or overwriting the log file.
Handles any errors during initialization gracefully.
*/
func initLogFile() {
	logDir := "./"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		return
	}

	logFilePath := filepath.Join(logDir, "amsh.log")
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}

	fmt.Printf("Log file successfully initialized: %s\n", logFilePath)
}

/*
Log a formatted message to the standard logger as well as to the log file.
*/
func Log(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if message == "" {
		return
	}
	writeToLog(message)
}

/*
Raw is a full decomposition of the object passed in.
*/
func Raw(v ...interface{}) {
	spew.Dump(v...)
	writeToLog(spew.Sprint(v...))
}

/*
Trace logs a trace message to the logger.
*/
func Trace(v ...interface{}) {
	logger.Debug(v[0], v[1:]...)
	writeToLog(fmt.Sprintf("%v", v))
}

/*
Debug logs a debug message to the logger.
*/
func Debug(format string, v ...interface{}) {
	logger.Debug(fmt.Sprintf(format, v...))
	writeToLog(fmt.Sprintf(format, v...))
}

/*
Note is a custom log message with a different style.
*/
func Note(format string, v ...interface{}) {
	logger.Info(fmt.Sprintf(format, v...))
	writeToLog(fmt.Sprintf(format, v...))
}

/*
Success is a custom log message with a different style.
*/
func Success(format string, v ...interface{}) {
	logger.Info(fmt.Sprintf(format, v...))
	writeToLog(fmt.Sprintf(format, v...))
}

/*
Info logs an info message to the logger.
*/
func Info(format string, v ...interface{}) {
	logger.Info(fmt.Sprintf(format, v...))
	writeToLog(fmt.Sprintf(format, v...))
}

/*
Warn logs a warn message to the logger.
*/
func Warn(format string, v ...interface{}) {
	logger.Warn(fmt.Sprintf(format, v...))
	writeToLog(fmt.Sprintf(format, v...))
}

/*
Error logs the error and returns it, useful for inline error logging and returning.

Example usage:

	err := someFunction()
	if err != nil {
		return Error(err, "additional context")
	}
*/
func Error(err error, v ...interface{}) error {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf("%s\n%s", err.Error(), getStackTrace())
	message += "\n" + getCodeSnippet(err.Error(), 0, 10)

	logger.Error(message, v...)
	writeToLog(message)
	return err
}

/*
Write a log message to the log file, ensuring thread safety.
*/
func writeToLog(message string) {
	if message == "" || logFile == nil {
		return
	}

	logFileMu.Lock()
	defer logFileMu.Unlock()

	// Strip ANSI escape codes and add a timestamp
	formattedMessage := fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), stripansi.Strip(strings.TrimSpace(message)))

	_, err := logFile.WriteString(formattedMessage)
	if err != nil {
		fmt.Printf("Failed to write to log file: %v\n", err)
	}

	// Ensure the write is flushed to disk
	if err := logFile.Sync(); err != nil {
		fmt.Printf("Failed to sync log file: %v\n", err)
	}
}

/*
Retrieve and format a stack trace from the current execution point.
*/
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var trace strings.Builder
	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		funcName := frame.Function
		if lastSlash := strings.LastIndexByte(funcName, '/'); lastSlash >= 0 {
			funcName = funcName[lastSlash+1:]
		}
		funcName = strings.Replace(funcName, ".", ":", 1)

		line := fmt.Sprintf("%s at %s(line %d)\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6E95F7")).Render(funcName),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#06C26F")).Render(filepath.Base(frame.File)),
			frame.Line,
		)
		trace.WriteString(line)
	}

	return "\n===[STACK TRACE]===\n" + trace.String() + "===[/STACK TRACE]===\n"
}

/*
Retrieve and return a code snippet surrounding the given line in the provided file.
*/
func getCodeSnippet(file string, line, radius int) string {
	fileHandle, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer fileHandle.Close()

	scanner := bufio.NewScanner(fileHandle)
	currentLine := 1
	var snippet strings.Builder

	for scanner.Scan() {
		if currentLine >= line-radius && currentLine <= line+radius {
			prefix := "  "
			if currentLine == line {
				prefix = "> "
			}
			snippet.WriteString(fmt.Sprintf("%s%d: %s\n", prefix, currentLine, scanner.Text()))
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return ""
	}

	return snippet.String()
}
