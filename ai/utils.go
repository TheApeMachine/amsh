package ai

import (
	"encoding/json"
	"regexp"

	"github.com/gofiber/fiber/v3"
)

/*
ExtractJSON finds Markdown JSON blocks and returns an array of JSON objects.
*/
func ExtractJSON(content string) []fiber.Map {
	jsonBlocks := []fiber.Map{}

	re := regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```")
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		var jsonBlock fiber.Map
		err := json.Unmarshal([]byte(match[1]), &jsonBlock)
		if err != nil {
			continue
		}
		jsonBlocks = append(jsonBlocks, jsonBlock)
	}

	return jsonBlocks
}

/*
Colors returns a list of all ANSI colors.
*/
var Colors = []string{
	"\033[37m",
	"\033[30m",
	"\033[31m",
	"\033[32m",
	"\033[33m",
	"\033[34m",
	"\033[35m",
	"\033[36m",
}
