package mastercomputer

import (
	"fmt"
	"strings"

	"github.com/theapemachine/amsh/utils"
)

type MsgType uint

const (
	MsgTypeString MsgType = iota
	MsgTypeSystem
	MsgTypeUser
	MsgTypeAssistant
	MsgTypeToolCall
	MsgTypeToolMessage
)

type Output struct {
	Content string
}

func NewOutput() *Output {
	return &Output{}
}

func (output *Output) Console(worker *Worker, msgType MsgType, str ...string) (out string) {
	switch msgType {
	case MsgTypeString:
		out = strings.Join(str, " ")
	case MsgTypeSystem:
		output.renderBanner(strings.Join(str, " "))
	case MsgTypeUser:
		output.renderBanner(strings.Join(str, " "))
	case MsgTypeAssistant:
		output.renderCard(worker.name, worker.role, strings.Join(str, " "))
	case MsgTypeToolCall:
		out = output.renderToolCall(worker.name, worker.role, str...)
	}

	return
}

func (output *Output) renderBanner(msg string) (out string) {
	out = fmt.Sprintf(
		"**********************************************\n%s\n**********************************************\n", msg,
	)
	fmt.Println(utils.Muted(out))
	return
}

func (output *Output) renderCard(name string, role string, content string) (out string) {
	out = fmt.Sprintf("┌───[ %s (%s) ]───────────────────────────────", name, role)
	for _, line := range strings.Split(content, "\n") {
		out += fmt.Sprintf("%s%s\n", utils.Muted("| "), utils.Highlight(line))
	}
	out += fmt.Sprintf("%s\n", utils.Muted("└──────────────────────────────────────────────────────────────"))
	fmt.Println(out)
	return
}

func (output *Output) renderToolCall(name string, role string, str ...string) (out string) {
	var (
		tool   string
		args   string
		result string
	)

	if len(str) == 1 {
		tool = str[0]
	} else if len(str) == 2 {
		tool = str[0]
		args = str[1]
	} else {
		tool = str[0]
		args = str[1]
		result = str[2]
	}

	out = fmt.Sprintf("┌───< %s (%s) >───────────────────────────────", name, role)

	if tool != "" {
		out += fmt.Sprintf("%s%s\n", utils.Muted("| "), utils.Red(tool))
	}

	if args != "" {
		out += fmt.Sprintf("%s%s\n", utils.Muted("| "), utils.Yellow(args))
	}

	if result != "" {
		out += fmt.Sprintf("%s%s\n", utils.Muted("| "), utils.Green(result))
	}

	out += fmt.Sprintf("%s\n", utils.Muted("└──────────────────────────────────────────────────────────────"))
	fmt.Println(out)
	return
}
