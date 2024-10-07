package ai

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

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

func ReplaceHolders(value string, values [][]string) string {
	for _, replacement := range values {
		value = strings.ReplaceAll(value, replacement[0], replacement[1])
	}

	return value
}

func ChunksToResponse(chunks []Chunk) string {
	if len(chunks) == 0 {
		return ""
	}

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf(
		"[%s (%s)]\n\n", chunks[0].Agent.ID, chunks[0].Agent.Type,
	))

	for _, chunk := range chunks {
		builder.WriteString(chunk.Response)
	}

	builder.WriteString("\n\n")
	return builder.String()
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
