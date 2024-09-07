package utils

import (
	"errors"
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
