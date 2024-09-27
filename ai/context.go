package ai

import (
	"regexp"
	"strings"
)

type Context struct {
	Agent     *Agent
	History   []string
	Responses []string
}

func NewContext(agent *Agent) *Context {
	return &Context{
		Agent:     agent,
		History:   make([]string, 0),
		Responses: make([]string, 0),
	}
}

func (ctx *Context) ExtractChecklist(step int) []string {
	re := regexp.MustCompile(`(?m)^\s*-\s*(.*)$`)
	matches := re.FindAllStringSubmatch(ctx.Responses[step], -1)

	var checklistItems []string

	for _, match := range matches {
		if len(match) > 1 {
			checklistItems = append(checklistItems, strings.TrimSpace(match[1]))
		}
	}

	return checklistItems
}
