package main

import (
	"os"

	"github.com/theapemachine/amsh/cmd"
	"github.com/theapemachine/amsh/errnie"
)

func main() {
	errnie.Debug("Starting AMSH")
	if err := cmd.Execute(); err != nil {
		errnie.Error(err)
		os.Exit(1)
	}
	errnie.Debug("AMSH finished")
}
