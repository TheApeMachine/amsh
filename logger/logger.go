package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	logger      *log.Logger
	fileLogger  *log.Logger
	file        *os.File
	tickID      int
	indentLevel int
)

// Init initializes two loggers: one for the file and one for in-memory (TUI)
func Init(filename string) error {
	var err error
	file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}

	// Create file logger without colors
	fileLogger = log.NewWithOptions(file, log.Options{
		ReportCaller: true,
		CallerOffset: 2,
	})

	// Create in-memory logger (use io.Discard for TUI output as you manage separately)
	logger = log.NewWithOptions(io.Discard, log.Options{
		ReportCaller: true,
		CallerOffset: 2,
	})

	return nil
}

func SetIndentLevel(level int) {
	indent := strings.Repeat("    ", level)

	styles := log.DefaultStyles()
	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString(indent).
		Background(lipgloss.Color("204")).
		Foreground(lipgloss.Color("0"))

	fileLogger.SetStyles(styles)
	logger.SetStyles(styles)
}

// Close closes the log file
func Close() {
	if file != nil {
		file.Close()
	}
}

// StartTick marks the beginning of a "tick" cycle
func StartTick() {
	if fileLogger == nil {
		return
	}

	indentLevel = 0
	SetIndentLevel(indentLevel)
	tickID++

	symbol := "🟢"

	fileLogger.Info(fmt.Sprintf("\n\n%s [%d] %s", symbol, tickID, "START TICK"))
	logger.Info(fmt.Sprintf("\n\n%s [%d] %s", symbol, tickID, "START TICK"))
}

// EndTick marks the end of a "tick" cycle
func EndTick() {
	if fileLogger == nil {
		return
	}
	symbol := "🔴"

	fileLogger.Info(fmt.Sprintf("\n%s [%d] %s", symbol, tickID, "END TICK"))
	logger.Info(fmt.Sprintf("\n%s [%d] %s", symbol, tickID, "END TICK"))
}

// StartSection starts a new section in the log and returns a function to end the section
func StartSection(name, sectionType string) func() {
	if fileLogger == nil {
		return func() {}
	}

	indentLevel++
	SetIndentLevel(indentLevel)

	symbol := "🟩"

	indent := strings.Repeat("    ", indentLevel)
	fileLogger.Info(fmt.Sprintf("\n%s [%d] %s", indent+symbol, tickID, name))
	logger.Info(fmt.Sprintf("\n%s [%d] %s", indent+symbol, tickID, name))

	// Return a function to end the section
	return func() {
		EndSection(name)
	}
}

// EndSection ends the current section in the log
func EndSection(name string) {
	if fileLogger == nil {
		return
	}

	symbol := "🟥"

	indent := strings.Repeat("    ", indentLevel)
	fileLogger.Info(fmt.Sprintf("\n%s [%d] %s", indent+symbol, tickID, name))
	logger.Info(fmt.Sprintf("\n%s [%d] %s", indent+symbol, tickID, name))

	indentLevel--
	SetIndentLevel(indentLevel)
}

// LogWithGroup logs a message under a specific functional group (e.g., BUFFER, EDITOR)
func LogWithGroup(group, format string, v ...interface{}) {
	if fileLogger == nil || logger == nil {
		return
	}

	_, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s:%d", filepath.Base(file), line)

	message := fmt.Sprintf(format, v...)
	fileLogger.Info(fmt.Sprintf("[%s] %s", group, message), "caller", caller)
	logger.Info(fmt.Sprintf("[%s] %s", group, message), "caller", caller)
}

// IndentedLog logs messages with indentation to show nesting levels
func IndentedLog(indentLevel int, format string, v ...interface{}) {
	if fileLogger == nil || logger == nil {
		return
	}

	indent := strings.Repeat("    ", indentLevel)
	message := fmt.Sprintf(format, v...)
	fileLogger.Info(indent + message)
	logger.Info(indent + message)
}

func LogWithSymbol(symbol, format string, v ...interface{}) {
	if fileLogger == nil || logger == nil {
		return
	}

	// The log message will automatically get indented based on the global indentLevel
	message := fmt.Sprintf(format, v...)
	fileLogger.Info(fmt.Sprintf("%s %s", symbol, message))
	logger.Info(fmt.Sprintf("%s %s", symbol, message))
}

// Print logs a message with the appropriate symbol
func Print(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

// Log logs a message with the appropriate symbol
func Log(format string, v ...interface{}) {
	LogWithSymbol("🔷", format, v...)
}

// Debug logs a debug message with the appropriate symbol
func Debug(format string, v ...interface{}) {
	LogWithSymbol("🐛", format, v...)
}

// Info logs an info message with the appropriate symbol
func Info(format string, v ...interface{}) {
	LogWithSymbol("🔷", format, v...)
}

// Warn logs a warning message with the appropriate symbol
func Warn(format string, v ...interface{}) {
	LogWithSymbol("⚠️", format, v...)
}

// Error logs an error message with the appropriate symbol
func Error(format string, v ...interface{}) {
	if len(v) == 0 || v[0] == nil {
		return
	}

	LogWithSymbol("❗", format, v...)
}

// GetLogger returns the logger instance (for external usage if needed)
func GetLogger() *log.Logger {
	return logger
}
