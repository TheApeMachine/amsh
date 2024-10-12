package errnie

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/invopop/jsonschema"
	"github.com/spf13/viper"
)

var fixing = false

func GenerateSchema[T any]() interface{} {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

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

type ErrorAnalysis struct {
	Steps []Step `json:"steps" jsonschema_description:"Steps taken to analyze the error"`
	Fixes []Fix  `json:"fixes" jsonschema_description:"Fixes needed to resolve the error, if any, otherwise leave empty for the next iteration" jsonschema_omitempty:"true"`
}

type Step struct {
	Thought string  `json:"thought" jsonschema_description:"Thoughts on what is happening"`
	Missing string  `json:"missing" jsonschema_description:"What is missing to fully understand the error"`
	Request Request `json:"request" jsonschema_description:"Request for additional context"`
}

type Request struct {
	Filenames []string `json:"filenames" jsonschema_description:"Filenames to include in the context"`
	Searches  []string `json:"searches" jsonschema_description:"Search the full codebase by keywords"`
}

type Fix struct {
	Old string `json:"old" jsonschema_description:"Old code fragment matched for replacement"`
	New string `json:"new" jsonschema_description:"New code fragment to replace the old code"`
	Why string `json:"why" jsonschema_description:"Detailed reason for why you are sure this is the correct fix."`
}

func JSONtoMap(jsonString string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonString), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func init() {
	initLogFile()
	// sync.OnceFunc(func() {
	// 	// Periodically print the number of active goroutines.
	// 	go func() {
	// 		for range time.Tick(time.Second * 5) {
	// 			fmt.Printf("Active goroutines: %d\n", runtime.NumGoroutine())
	// 		}
	// 	}()
	// })()
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
	if !fixing {
		fmt.Println(message)
	}
	writeToLog(message)
}

// Raw logs a raw message with the appropriate symbol
func Raw(obj any) {
	level := viper.GetViper().GetString("loglevel")

	if level != "trace" && level != "debug" {
		return
	}

	if !fixing {
		spew.Dump(obj)
	}
	writeToLog(spew.Sdump(obj))
}

// Debug logs a debug message with the appropriate symbol
func Debug(format string, v ...interface{}) {
	level := viper.GetViper().GetString("loglevel")
	if level != "trace" && level != "debug" {
		return
	}

	message := fmt.Sprintf("🐛 %s", fmt.Sprintf(format, v...))
	if !fixing {
		fmt.Println(message)
	}
	writeToLog(message)
}

// Info logs an info message with the appropriate symbol
func Info(format string, v ...interface{}) {
	message := fmt.Sprintf("🔷 %s", fmt.Sprintf(format, v...))
	if !fixing {
		fmt.Println(message)
	}
	writeToLog(message)
}

// Warn logs a warning message with the appropriate symbol
func Warn(format string, v ...interface{}) {
	message := fmt.Sprintf("⚠️ %s", fmt.Sprintf(format, v...))
	if !fixing {
		fmt.Println(message)
	}
	writeToLog(message)
}

var (
	errorHandler *ErrorHandler
	initOnce     sync.Once
)

// Error logs an error message with the appropriate symbol, a code snippet, and a stack trace
func Error(err error) error {
	if err == nil {
		return nil
	}

	// Capture the caller's file and line number
	var pc [10]uintptr
	n := runtime.Callers(2, pc[:])
	if n == 0 {
		message := fmt.Sprintf("❗ %v", err)
		if !fixing {
			fmt.Println(message)
		}
		writeToLog(message)
		return fmt.Errorf(message)
	}

	frames := runtime.CallersFrames(pc[:n])
	var relevantFrame runtime.Frame
	for i := 0; i < 3; i++ {
		frame, more := frames.Next()
		if !more {
			break
		}
		relevantFrame = frame
	}
	file := relevantFrame.File
	line := relevantFrame.Line

	// Format the error message with the function name, file, and line number
	message := fmt.Sprintf("❗ %s:%d %v", file, line, err)
	if !fixing {
		fmt.Println(message)
	}
	writeToLog(message)

	// Display a code snippet from the file (e.g., 2 lines before and after the error line)
	const snippetRadius = 2
	codeSnippet := getCodeSnippet(file, line, snippetRadius)
	if codeSnippet != "" {
		snippetMessage := fmt.Sprintf("📄 Code snippet (around %s:%d):\n%s", file, line, codeSnippet)
		if !fixing {
			fmt.Println(snippetMessage)
		}
		writeToLog(snippetMessage)
	}

	// Capture and print the stack trace
	stackTrace := getStackTrace()

	if !fixing {
		fmt.Println("📊 Stack trace:")
		fmt.Println(stackTrace)
	}
	writeToLog(stackTrace)

	if !fixing {
		initOnce.Do(func() {
			errorHandler = NewErrorHandler()
			fixing = true
			go func() {
				defer func() {
					fixing = false
				}()
				if err := errorHandler.Error(err); err != nil {
					fmt.Printf("Error during analysis and fix process: %v\n", err)
				}
				// Consider removing this line if you don't want to exit the program after analysis
				os.Exit(1)
			}()
		})
	}

	return fmt.Errorf(message)
}
