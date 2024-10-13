package utils

import (
	"errors"
	"fmt"
	"os"
)

/*
CheckFileExists is used to verify if the embedded config file has already been written to the user's home directory.
This is to make sure we don't overwrite the user's existing config file, which they may have customized.
*/
func CheckFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}

func EnsureLogsDir() error {
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}
	return nil
}

func IsOnlyNewlines(s string) bool {
	for _, c := range s {
		if c != '\n' {
			return false
		}
	}
	return true
}
