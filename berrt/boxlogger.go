package berrt

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/theckman/yacspin"
)

/*
BoxStyle represents different styles for the box logger.
*/
type BoxStyle string

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
	stackTrace := bl.captureStackTrace()
	codeSnippet := bl.captureCodeSnippet(stackTrace)

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
	emojiPrefix := getEmojiForStyle(style)
	titleLine := bl.alignTitle(emojiPrefix, title)
	contentStr := bl.renderValue(content, 1)
	contentLines := strings.Split(contentStr, "\n")
	maxContentWidth := len(titleLine)

	for _, line := range contentLines {
		if len(line) > maxContentWidth {
			maxContentWidth = len(line)
		}
	}

	bl.Width = max(maxContentWidth+2, bl.Width)
	top := fmt.Sprintf("╭%s╮", strings.Repeat("─", bl.Width))
	wrappedContent := ""

	for _, line := range contentLines {
		wrappedContent += bl.wrapContentLine(line, bl.Width-2)
	}

	bottom := fmt.Sprintf("╰%s╯", strings.Repeat("─", bl.Width))
	return fmt.Sprintf("%s\n%s%s", top, wrappedContent, bottom)
}

/*
wrapContentLine wraps a line of content to fit within the given width
*/
func (bl *BoxLogger) wrapContentLine(line string, maxWidth int) string {
	var wrappedContent strings.Builder
	indent := "    " // The indentation to be used for wrapping
	words := strings.Fields(line)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 > maxWidth-len(indent) {
			wrappedContent.WriteString(fmt.Sprintf("│ %s%-*s │\n", indent, maxWidth-len(indent), currentLine))
			currentLine = word
		} else {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		}
	}

	if currentLine != "" {
		wrappedContent.WriteString(fmt.Sprintf("│ %s%-*s │\n", indent, maxWidth-len(indent), currentLine))
	}

	return wrappedContent.String()
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
func (bl *BoxLogger) captureStackTrace() string {
	stackBuf := make([]byte, 1024)
	n := runtime.Stack(stackBuf, false)
	stack := strings.Split(string(stackBuf[:n]), "\n")
	var formattedStack strings.Builder

	for _, line := range stack {
		formattedStack.WriteString(fmt.Sprintf("  📌 %s\n", line))
	}

	return formattedStack.String()
}

/*
captureCodeSnippet captures a snippet of the code from the error location
*/
func (bl *BoxLogger) captureCodeSnippet(stackTrace string) string {
	lines := strings.Split(stackTrace, "\n")
	if len(lines) < 2 {
		return "Could not determine the error location from the stack trace."
	}

	// Parse the filename and line number from the stack trace
	lineParts := strings.Split(lines[1], " ")
	if len(lineParts) < 2 {
		return "Error parsing stack trace for code snippet."
	}

	fileInfo := lineParts[len(lineParts)-1]
	fileAndLine := strings.Split(fileInfo, ":")
	if len(fileAndLine) != 2 {
		return "Error parsing file and line number."
	}

	filePath := fileAndLine[0]
	lineNumber := 0
	fmt.Sscanf(fileAndLine[1], "%d", &lineNumber)

	// Read the source file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Sprintf("Could not open source file: %s", filePath)
	}
	defer file.Close()

	// Read and extract lines around the error
	startLine := max(1, lineNumber-3)
	endLine := lineNumber + 3
	var snippet strings.Builder

	scanner := bufio.NewScanner(file)
	currentLine := 1
	for scanner.Scan() {
		if currentLine >= startLine && currentLine <= endLine {
			lineContent := scanner.Text()
			if currentLine == lineNumber {
				snippet.WriteString(fmt.Sprintf("👉 %3d: %s\n", currentLine, lineContent))
			} else {
				snippet.WriteString(fmt.Sprintf("    %3d: %s\n", currentLine, lineContent))
			}
		}
		currentLine++
		if currentLine > endLine {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "Error reading source file for code snippet."
	}

	return snippet.String()
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

/*
max returns the maximum of two integers
*/
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
