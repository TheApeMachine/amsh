package utils

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/goombaio/namegenerator"
	"github.com/theapemachine/amsh/ai"
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

/*
Turns ChainOfThought into a more human readable format,
using colors to make sections more readable.
*/
func BeautifyChainOfThought(data ai.ChainOfThought) {
	for _, step := range data.Steps {
		fmt.Println(color.RedString(step.Thought))
		fmt.Println(color.YellowString(step.Reasoning))
		fmt.Println(color.GreenString(step.NextStep))
	}

	fmt.Println(color.BlueString(data.Action))
	fmt.Println(color.MagentaString(data.Result))
}
