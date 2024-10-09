package utils

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/goombaio/namegenerator"
	"github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/format"
)

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

func BeautifyReasoning(ID string, reasoning format.Response) {
	fmt.Println("[", ID, "]")
	switch r := reasoning.(type) {
	case *format.ChainOfThought:
		fmt.Println(color.CyanString(r.ToString()))
	case *format.TreeOfThought:
		fmt.Println(color.GreenString(r.ToString()))
	case *format.FirstPrinciplesReasoning:
		fmt.Println(color.YellowString(r.ToString()))
	default:
		fmt.Println("Unknown reasoning type")
	}
	fmt.Println("[/", ID, "]")
	fmt.Println()
}

/*
Turns ChainOfThought into a more human readable format,
using colors to make sections more readable.
*/
func BeautifyChainOfThought(ID string, data format.ChainOfThought) {
	fmt.Println("[", ID, "]")
	for _, step := range data.Template.Steps {
		fmt.Println("[step]")
		fmt.Println("  thought:", color.RedString(step.Thought))
		fmt.Println("reasoning:", color.YellowString(step.Reasoning))
		fmt.Println("next step:", color.GreenString(step.NextStep))
		fmt.Println("[/step]")
		fmt.Println()
	}

	fmt.Println("   action:", color.BlueString(data.Template.Action))
	fmt.Println("   result:", color.MagentaString(data.Template.Result))
	fmt.Println("[/", ID, "]")
	fmt.Println()
}

func BeautifyToolCall(toolCall openai.ToolCall, args map[string]interface{}) {
	fmt.Println("[tool call]", color.BlueString(toolCall.Function.Name))

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

	fmt.Println("[/tool call]")
	fmt.Println()
}

func BeautifyMemory(memory *ai.Memory) {
	fmt.Println("[ Memory ]")
	fmt.Println("Short-Term Memory:")
	for _, entry := range memory.ShortTerm {
		fmt.Println("  -", color.YellowString(entry))
	}
	fmt.Println("[/ Memory ]")
}
