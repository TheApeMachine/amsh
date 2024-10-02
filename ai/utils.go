package ai

import "strings"

/*
ExtractJSON removes code block markers from the content string.
*/
func ExtractJSON(content string) []byte {
	content = strings.ReplaceAll(content, "```json", "")
	content = strings.ReplaceAll(content, "```", "")
	content = strings.TrimSpace(content)
	return []byte(content)
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
