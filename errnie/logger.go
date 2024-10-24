package errnie

import (
	"bufio"
	"encoding/json"
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
	Dark      = "#666666"
	Muted     = "#999999"
	Highlight = "#EEEEEE"
	Blue      = "#6E95F7"
	Red       = "#F7746D"
	Yellow    = "#F7B96D"
	Green     = "#06C26F"
	Purple    = "#6C50FF"

	styles = log.DefaultStyles()

	logFile     *os.File
	logFileMu   sync.Mutex
	logFilePath string
	nonofile    bool

	logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		CallerOffset:    3,
		ReportTimestamp: true,
		TimeFormat:      time.TimeOnly,
		Level:           log.InfoLevel,
	})
)

func JSONtoMap(jsonString string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonString), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func init() {
	nonofile = os.Getenv("LOGFILE") == ""

	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color(Red)).
		Foreground(lipgloss.Color(Highlight))
	styles.Levels[log.WarnLevel] = lipgloss.NewStyle().
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color(Yellow)).
		Foreground(lipgloss.Color(Highlight))
	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color(Blue)).
		Foreground(lipgloss.Color(Highlight))
	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color(Muted)).
		Foreground(lipgloss.Color(Highlight))

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		CallerOffset:    2,
		ReportTimestamp: true,
		TimeFormat:      time.TimeOnly,
	})

	logger.SetStyles(styles)

	switch viper.GetViper().GetString("loglevel") {
	case "trace":
		logger.SetLevel(log.DebugLevel)
	case "debug":
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

	initLogFile()
	sync.OnceFunc(func() {
		// Periodically print the number of active goroutines.
		go func() {
			for range time.Tick(time.Second * 5) {
				log.Debug("active goroutines", "count", runtime.NumGoroutine())
			}
		}()
	})()
}

func initLogFile() {
	if nonofile {
		return
	}

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

func Note(format string, v ...interface{}) {
	fmt.Println(lipgloss.NewStyle().Background(lipgloss.Color(Purple)).Foreground(lipgloss.Color(Highlight)).Render("NOTE"), fmt.Sprintf(format, v...))
}

/*
Trace logs a trace message with the appropriate symbol
*/
func Trace(v ...interface{}) {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	_, line := f.FileLine(pc[0])
	formatted := fmt.Sprintf("%d", line)

	// Print the method name with arguments
	fn := f.Name()
	for _, arg := range v {
		fn += fmt.Sprintf(" %v", arg)
	}

	logger.Debug("TRACE", "name", fn, "line", line)
	writeToLog(fmt.Sprintf("▫️  %s %s", fn, formatted))
}

/*
Raw provides a deep dump of the object, which is useful for
debugging complex data structures.
*/
func Raw(obj any) {
	level := viper.GetViper().GetString("loglevel")

	if level != "trace" && level != "debug" {
		return
	}

	logger.Debug(spew.Sdump(obj))
	writeToLog(spew.Sdump(obj))
}

/*
Debug logs a debug message with the appropriate symbol
*/
func Debug(msg interface{}, v ...interface{}) {
	logger.Debug(msg, v...)
}

/*
Info logs an info message with the appropriate symbol
*/
func Info(msg interface{}, v ...interface{}) {
	logger.Info(msg, v...)
}

/*
Warn logs a warning message with the appropriate symbol
*/
func Warn(msg interface{}, v ...interface{}) {
	logger.Warn(msg, v...)
}

/*
Error logs the error and returns it, which makes it easy to insert
errnie error logging in many types of situations, acting as a
transparent wrapper around the error.
*/
func Error(err error, v ...interface{}) error {
	if err == nil {
		return nil
	}

	// Build the error message with stack trace and code snippet.
	message := fmt.Sprintf("%s\n%s", err.Error(), getStackTrace())
	message += "\n" + getCodeSnippet(err.Error(), 0, 10)

	logger.Error(message, v...)
	writeToLog(message)
	return err
}

func writeToLog(message string, v ...interface{}) {
	if message == "" || nonofile {
		return
	}

	logFileMu.Lock()
	defer logFileMu.Unlock()
	_, err := logFile.WriteString(stripansi.Strip(fmt.Sprintf(message, v...)) + "\n")
	if err != nil {
		fmt.Printf("Failed to write to log file: %v\n", err)
	}
}

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

		// Format the function name
		funcName := frame.Function
		if lastSlash := strings.LastIndexByte(funcName, '/'); lastSlash >= 0 {
			funcName = funcName[lastSlash+1:]
		}
		funcName = strings.Replace(funcName, ".", ":", 1)

		// Construct the colored line
		line := fmt.Sprintf("%s%s%s %s(%d)\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color(Blue)).Render(funcName),
			lipgloss.NewStyle().Foreground(lipgloss.Color(Muted)).Render(" at "),
			lipgloss.NewStyle().Foreground(lipgloss.Color(Green)).Render(filepath.Base(frame.File)),
			lipgloss.NewStyle().Foreground(lipgloss.Color(Yellow)).Render("line"),
			frame.Line,
		)

		trace.WriteString(line)
	}

	return "\n===[STACK TRACE]===\n" + trace.String() + "===[/STACK TRACE]===\n"
}

func getCodeSnippet(file string, line, radius int) string {
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
