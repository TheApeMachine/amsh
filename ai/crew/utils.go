package crew

import "strings"

/*
ExtractJSON removes code block markers from the content string.
*/
func ExtractJSON(content string) string {
	content = strings.ReplaceAll(content, "```json", "")
	content = strings.ReplaceAll(content, "```", "")
	content = strings.TrimSpace(content)
	return content
}
