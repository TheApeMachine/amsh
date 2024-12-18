package utils

import (
	"encoding/json"
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

func StripMarkdown(s, language string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "```"+language, ""), "```", "")
}

func ContainsAny(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}

	return false
}

// ExtractJSONBlocks finds and parses JSON objects from a string
func ExtractJSONBlocks(s string) []map[string]interface{} {
	// Extract blocks marked with json language identifier
	codeBlocks := ExtractCodeBlocks(s)

	var results []map[string]interface{}
	for _, blocks := range codeBlocks["json"] {
		if block := ParseJSON(blocks); block != nil {
			results = append(results, block)
		}
	}

	return results
}

// ParseJSON safely parses a JSON string into a map
func ParseJSON(s string) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(s), &result); err == nil {
		return result
	}
	return nil
}
