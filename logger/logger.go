package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
	"bufio"
)

var (
	logFile *os.File
	writer  *bufio.Writer
	mu      sync.Mutex
)

func Init(logPath string) error {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		return fmt.Errorf("logger already initialized")
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	logFile = file
	writer = bufio.NewWriter(logFile)
	return nil
}

func Log(format string, v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if writer == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("%s: %s\n", timestamp, fmt.Sprintf(format, v...))

	if _, err := writer.WriteString(logEntry); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to log file: %v\n", err)
		return
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Error flushing log file: %v\n", err)
	}
}

func Close() {
	mu.Lock()
	defer mu.Unlock()

	if writer != nil {
		writer.Flush()
	}
	if logFile != nil {
		logFile.Close()
		logFile = nil
		writer = nil
	}
}
