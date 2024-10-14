package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/goombaio/namegenerator"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

func JSONtoMap(jsonString string) (map[string]any, error) {
	errnie.Debug(jsonString)
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonString), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func JoinWith(delim string, args ...string) string {
	return strings.Join(args, delim)
}

func ReplaceWith(template string, args [][]string) string {
	for _, arg := range args {
		template = strings.ReplaceAll(template, arg[0], arg[1])
	}

	return template
}

func StrategyInstructions(name string) string {
	prompt := viper.GetViper().GetString("ai.prompt.strategy." + name + ".instructions")

	if prompt == "" {
		return fmt.Errorf("no instructions for %s", name).Error()
	}

	return prompt
}

var existingIDs = make([]string, 0)

func NewID() string {
	newID := namegenerator.NewNameGenerator(time.Now().UnixNano()).Generate()

	for _, id := range existingIDs {
		if id == newID {
			return NewID()
		}
	}

	existingIDs = append(existingIDs, newID)
	return newID
}

func BeautifyToolCall(toolCall openai.ToolCall, args map[string]interface{}) {
	fmt.Println("[TOOL CALL]", color.BlueString(toolCall.Function.Name))

	// Find the longest key to determine the padding
	maxKeyLength := 0
	for key := range args {
		if len(key) > maxKeyLength {
			maxKeyLength = len(key)
		}
	}

	// Print each key-value pair with aligned colons
	for key, value := range args {
		fmt.Printf("%-*s : %v\n", maxKeyLength, key, value)
	}

	fmt.Println("[/TOOL CALL]")
	fmt.Println()
}
