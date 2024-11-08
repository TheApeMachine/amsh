package utils

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goombaio/namegenerator"
)

func JoinWith(delim string, args ...string) string {
	return strings.Join(args, delim)
}

func ReplaceWith(template string, args [][]string) string {
	for _, arg := range args {
		template = strings.ReplaceAll(template, "{"+arg[0]+"}", arg[1])
	}

	return template
}

func NewID() string {
	return uuid.New().String()
}

var existingNames = make([]string, 0)

func NewName() string {
	newName := namegenerator.NewNameGenerator(time.Now().UnixNano()).Generate()

	for _, name := range existingNames {
		if name == newName {
			return NewName()
		}
	}

	existingNames = append(existingNames, newName)
	return newName
}

func StringPtr(s string) *string {
	return &s
}

/*
ExtractCodeBlocks extracts Markdown code blocks from a string,
and returns a map of language to code block.
*/
func ExtractCodeBlocks(s string) map[string][]string {
	// Match code blocks with language identifiers
	re := regexp.MustCompile("```([a-zA-Z0-9]+)\n([\\s\\S]*?)```")
	matches := re.FindAllStringSubmatch(s, -1)

	codeBlocks := make(map[string][]string)
	for _, match := range matches {
		if len(match) >= 3 {
			language := match[1]
			code := strings.TrimSpace(match[2])
			codeBlocks[language] = append(codeBlocks[language], code)
		}
	}

	return codeBlocks
}

func StripMarkdown(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "```json", ""), "```", "")
}
