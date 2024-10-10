package errnie

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
)

var dark = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#666666")).Render
var muted = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#999999")).Render
var highlight = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#EEEEEE")).Render
var blue = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#6E95F7")).Render
var red = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#F7746D")).Render
var yellow = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#F7B96D")).Render
var green = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#06C26F")).Render

var (
	logFile     *os.File
	logFileMu   sync.Mutex
	logFilePath string
)

func init() {
	initLogFile()
}

func initLogFile() {
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		return
	}

	timestamp := time.Now().UnixNano()
	logFilePath = filepath.Join(logDir, fmt.Sprintf("amsh-%d.log", timestamp))

	var err error
	logFile, err = os.Create(logFilePath)
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
	}
}

func writeToLog(message string) {
	logFileMu.Lock()
	defer logFileMu.Unlock()

	if logFile != nil {
		fmt.Fprintln(logFile, stripansi.Strip(message))
	}
}

/*
Trace logs a trace message with the appropriate symbol
*/
func Trace() {
	if viper.GetViper().GetString("loglevel") != "trace" {
		return
	}

	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	_, line := f.FileLine(pc[0])
	formatted := fmt.Sprintf("%d", line)
	message := fmt.Sprintf("▫️  %s %s", muted(f.Name()), blue(formatted))
	fmt.Println(message)
	writeToLog(message)
}

// Raw logs a raw message with the appropriate symbol
func Raw(obj any) {
	level := viper.GetViper().GetString("loglevel")
	if level != "trace" && level != "debug" {
		return
	}

	spew.Dump(obj)
	writeToLog(spew.Sdump(obj))
}

// Debug logs a debug message with the appropriate symbol
func Debug(format string, v ...interface{}) {
	level := viper.GetViper().GetString("loglevel")
	if level != "trace" && level != "debug" {
		return
	}

	message := fmt.Sprintf("🐛 %s", fmt.Sprintf(format, v...))
	fmt.Println(message)
	writeToLog(message)
}

// Info logs an info message with the appropriate symbol
func Info(format string, v ...interface{}) {
	message := fmt.Sprintf("🔷 %s", fmt.Sprintf(format, v...))
	fmt.Println(message)
	writeToLog(message)
}

// Warn logs a warning message with the appropriate symbol
func Warn(format string, v ...interface{}) {
	message := fmt.Sprintf("⚠️ %s", fmt.Sprintf(format, v...))
	fmt.Println(message)
	writeToLog(message)
}

// Error logs an error message with the appropriate symbol
func Error(err error) error {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf("❗ %v", err)
	fmt.Println(message)
	writeToLog(message)
	return fmt.Errorf(message)
}
