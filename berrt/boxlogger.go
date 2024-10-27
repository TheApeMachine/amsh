package berrt

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/theckman/yacspin"
)

/*
BoxStyle represents different styles for the box logger.
*/
type BoxStyle string

var Dark = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#666666")).Render
var Muted = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#999999")).Render
var Highlight = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#EEEEEE")).Render
var Blue = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#6E95F7")).Render
var Red = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#F7746D")).Render
var Yellow = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#F7B96D")).Render
var Green = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#06C26F")).Render
var Purple = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("#6C50FF")).Render

const (
	AIStyle      BoxStyle = "ai"
	SuccessStyle BoxStyle = "success"
	InfoStyle    BoxStyle = "info"
	WarningStyle BoxStyle = "warning"
	ErrorStyle   BoxStyle = "error"
)

/*
HeaderAlignment defines the place of the header label in the
top part of the box.
*/
type HeaderAlignment string

const (
	HeaderLeft   HeaderAlignment = "left"
	HeaderRight  HeaderAlignment = "right"
	HeaderCenter HeaderAlignment = "center"
)

/*
BoxLogger is a highly segmented way to log messages. It is both
optimized for human readability and machine readability.
*/
type BoxLogger struct {
	HeaderAlignment HeaderAlignment
	Width           int
}

var bl *BoxLogger
var once sync.Once

func init() {
	once.Do(func() {
		bl = NewBoxLogger()
	})
}

/*
NewBoxLogger initializes a new BoxLogger with default settings.
*/
func NewBoxLogger() *BoxLogger {
	return &BoxLogger{
		HeaderAlignment: HeaderLeft,
		Width:           80,
	}
}

func WithSpinner(task func(), msg string) {
	// Create a spinner configuration
	spinnerConfig := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[59],
		Suffix:          " ",
		SuffixAutoColon: true,
		Message:         msg,
		StopCharacter:   "✓",
		StopColors:      []string{"fgGreen"},
	}

	// Initialize the spinner
	spinner, err := yacspin.New(spinnerConfig)
	if err != nil {
		fmt.Println("Error creating spinner:", err)
		return
	}

	// Start the spinner before executing the task
	spinner.Start()

	// Execute the given task
	task()

	// Stop the spinner after the task is completed
	spinner.Stop()
}

/*
Success logs a message with the "Success" style
*/
func AI(title string, content any) {
	fmt.Println(bl.drawBox(title, AIStyle, content))
}

/*
Success logs a message with the "Success" style
*/
func Success(title string, content any) {
	fmt.Println(bl.drawBox(title, SuccessStyle, content))
}

/*
Info logs a message with the "Info" style
*/
func Info(title string, content any) {
	fmt.Println(bl.drawBox(title, InfoStyle, content))
}

/*
Warning logs a message with the "Warning" style
*/
func Warning(title string, content any) {
	fmt.Println(bl.drawBox(title, WarningStyle, content))
}

/*
Error logs an error with a stack trace and code preview
*/
func Error(title string, err error) {
	if err == nil {
		return
	}

	stackTrace, file, line := bl.captureStackTrace()
	codeSnippet := bl.captureCodeSnippet(file, line, 5)

	content := map[string]any{
		"Error":        err.Error(),
		"Stack Trace":  stackTrace,
		"Code Snippet": codeSnippet,
	}
	fmt.Println(bl.drawBox(title, ErrorStyle, content))
}

/*
drawBox is the internal function that creates the box. It takes
care of dynamically adjusting the width of the box based on
the content and the alignment of the title.
*/
func (bl *BoxLogger) drawBox(title string, style BoxStyle, content any) string {
	color := getColorForStyle(style)
	emojiPrefix := getEmojiForStyle(style)
	titleLine := bl.alignTitle(emojiPrefix, title)
	contentStr := bl.renderValue(content, 1)
	contentLines := strings.Split(contentStr, "\n")

	maxContentWidth := 80

	bl.Width = max(maxContentWidth+2, bl.Width)
	out := []string{
		color("╭[") + Muted(titleLine) + color("]"+strings.Repeat("─", bl.Width-(len(titleLine)-3))+"╮"),
	}

	for _, line := range contentLines {
		out = append(out, bl.wrapContentLine(line, bl.Width-2, color))
	}

	out = append(out, color("╰"+strings.Repeat("─", bl.Width)+"╯"))
	return strings.Join(out, "\n")
}

/*
wrapContentLine wraps a line of content to fit within the given width
*/
func (bl *BoxLogger) wrapContentLine(line string, maxWidth int, color func(strs ...string) string) string {
	out := []string{}
	indent := "    "
	remainingWidth := maxWidth - len(indent)
	words := strings.Fields(line)

	if len(words) == 0 {
		// Handle empty line
		padding := strings.Repeat(" ", maxWidth)
		return color("│ " + padding + " │")
	}

	currentLine := indent
	for _, word := range words {
		// Check if adding this word would exceed the width
		if len(currentLine)+len(word)+1 > remainingWidth {
			// Add current line to output with proper padding
			padding := strings.Repeat(" ", maxWidth-len(currentLine))
			out = append(out, color("│ "+Highlight(currentLine+padding)+color(" │")))

			// Start new line with indent
			currentLine = indent + word
		} else {
			// Add word to current line
			if currentLine == indent {
				currentLine += word
			} else {
				currentLine += " " + word
			}
		}
	}

	// Add final line if there's anything left
	if currentLine != indent {
		padding := strings.Repeat(" ", maxWidth-len(currentLine))
		out = append(out, color("│ "+Highlight(currentLine+padding)+color(" │")))
	}

	return strings.Join(out, "\n")
}

/*
alignTitle aligns the title based on the HeaderAlignment configuration
*/
func (bl *BoxLogger) alignTitle(emojiPrefix string, title string) string {
	switch bl.HeaderAlignment {
	case HeaderLeft:
		return fmt.Sprintf("%s %s", emojiPrefix, title)
	case HeaderCenter:
		totalTitleLength := len(emojiPrefix) + len(title) + 1
		padding := (bl.Width - totalTitleLength) / 2
		return fmt.Sprintf("%s%s %s", strings.Repeat(" ", padding), emojiPrefix, title)
	case HeaderRight:
		totalTitleLength := len(emojiPrefix) + len(title) + 1
		padding := bl.Width - totalTitleLength - 2
		return fmt.Sprintf("%s%s %s", strings.Repeat(" ", padding), emojiPrefix, title)
	default:
		return fmt.Sprintf("%s %s", emojiPrefix, title)
	}
}

/*
renderValue converts various input types into a readable string
representation
*/
func (bl *BoxLogger) renderValue(value any, indentLevel int) string {
	indent := strings.Repeat("    ", indentLevel)
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Map:
		result := ""
		for _, key := range v.MapKeys() {
			result += fmt.Sprintf("\n%s%s: %v", indent, key, bl.renderValue(v.MapIndex(key).Interface(), indentLevel+1))
		}
		return result
	case reflect.Slice, reflect.Array:
		result := "["
		for i := 0; i < v.Len(); i++ {
			result += fmt.Sprintf("\n%s- %v", indent, bl.renderValue(v.Index(i).Interface(), indentLevel+1))
		}
		result += "\n" + indent[:len(indent)-4] + "]"
		return result
	case reflect.Struct:
		result := ""
		typeOfV := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := typeOfV.Field(i)
			fieldValue := v.Field(i)
			result += fmt.Sprintf("\n%s%s: %v", indent, field.Name, bl.renderValue(fieldValue.Interface(), indentLevel+1))
		}
		return result
	case reflect.Ptr:
		if v.IsNil() {
			return "nil"
		}
		return bl.renderValue(v.Elem().Interface(), indentLevel)
	default:
		return fmt.Sprintf("%v", value)
	}
}

/*
captureStackTrace captures a stack trace for the current goroutine
*/
func (bl *BoxLogger) captureStackTrace() (string, string, int) {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	file := ""
	line := 0

	out := []string{
		"===[STACK TRACE]===",
	}

	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		if file == "" {
			file = frame.File
			line = frame.Line
		}

		// Format the function name
		funcName := frame.Function
		if lastSlash := strings.LastIndexByte(funcName, '/'); lastSlash >= 0 {
			funcName = funcName[lastSlash+1:]
		}
		funcName = strings.Replace(funcName, ".", ":", 1)

		// Construct the colored line
		line := fmt.Sprintf("%s%s%s %s(%d)\n",
			Blue(funcName),
			Muted(" at "),
			Green(filepath.Base(frame.File)),
			Yellow("line"),
			frame.Line,
		)

		out = append(out, line)
	}

	out = append(out, "===[/STACK TRACE]===")

	return strings.Join(out, "\n"), file, line
}

/*
captureCodeSnippet captures a snippet of the code from the error location
*/
func (bl *BoxLogger) captureCodeSnippet(file string, line, radius int) string {
	fileHandle, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer fileHandle.Close()

	scanner := bufio.NewScanner(fileHandle)
	currentLine := 1
	var snippet string

	for scanner.Scan() {
		if currentLine >= line-radius && currentLine <= line+radius {
			prefix := "  "
			if currentLine == line {
				prefix = "> "
			}
			snippet += fmt.Sprintf("%s%d: %s\n", prefix, currentLine, scanner.Text())
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return ""
	}

	return snippet
}

/*
getEmojiForStyle returns the emoji for the given style.
*/
func getEmojiForStyle(style BoxStyle) string {
	switch style {
	case AIStyle:
		return "✨"
	case SuccessStyle:
		return "🤘"
	case InfoStyle:
		return "ℹ️ "
	case WarningStyle:
		return "⚠️ "
	case ErrorStyle:
		return "🤬"
	default:
		return "🟢"
	}
}

func getColorForStyle(style BoxStyle) func(strs ...string) string {
	switch style {
	case AIStyle:
		return Purple
	case SuccessStyle:
		return Green
	case InfoStyle:
		return Blue
	case WarningStyle:
		return Yellow
	case ErrorStyle:
		return Red
	}

	return Highlight
}

/*
max returns the maximum of two integers
*/
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
