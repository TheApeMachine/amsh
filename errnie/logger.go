package errnie

import (
	"fmt"
	"runtime"

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
	fmt.Println("▫️", " ", muted(f.Name()), " ", blue(formatted))
}

// Raw logs a raw message with the appropriate symbol
func Raw(obj any) {
	level := viper.GetViper().GetString("loglevel")
	if level != "trace" && level != "debug" {
		return
	}

	spew.Dump(obj)
}

// Debug logs a debug message with the appropriate symbol
func Debug(format string, v ...interface{}) {
	level := viper.GetViper().GetString("loglevel")
	if level != "trace" && level != "debug" {
		return
	}

	fmt.Println("🐛", fmt.Sprintf(format, v...))
}

// Info logs an info message with the appropriate symbol
func Info(format string, v ...interface{}) {
	fmt.Println("🔷", fmt.Sprintf(format, v...))
}

// Warn logs a warning message with the appropriate symbol
func Warn(format string, v ...interface{}) {
	fmt.Println("⚠️ ", fmt.Sprintf(format, v...))
}

// Error logs an error message with the appropriate symbol
func Error(err error, v ...interface{}) error {
	if err == nil {
		return nil
	}

	fmt.Println("❗", red("[ERROR]"), err.Error(), fmt.Sprintf(v[0].(string), v[1:]...))
	return err
}
