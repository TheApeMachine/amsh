package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/theapemachine/amsh/cmd"
	"github.com/theapemachine/errnie"
)

func main() {
	errnie.Debug("Starting AMSH")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)

	go func() {
		<-c
		fmt.Println("\nProgram interrupted! Printing stack trace:")
		printStackTrace()
		os.Exit(1)
	}()

	if err := cmd.Execute(); err != nil {
		errnie.Error(err)
		os.Exit(1)
	}

	errnie.Debug("AMSH finished")
}

func printStackTrace() {
	buf := make([]byte, 1<<16) // 64KB buffer size
	stackSize := runtime.Stack(buf, true)
	stackTrace := string(buf[:stackSize])

	// Parse and filter the stack trace for relevant lines
	for _, line := range strings.Split(stackTrace, "\n") {
		// Print only lines containing file and line number information
		if strings.Contains(line, ".go:") {
			fmt.Println(line)
		}
	}
}
