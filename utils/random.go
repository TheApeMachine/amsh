package utils

import (
	"math/rand"
)

// GenerateRandomID generates a random ID for the section
func GenerateRandomID() int {
	return rand.Intn(1000)
}
