package berrt

import (
	"fmt"
	"strings"

	"github.com/theapemachine/amsh/utils"
)

/*
Consumer is a specialized logging type designed to handle streaming, chunked JSON strings.
It strips away the structural elements of the JSON while maintaining indentation levels,
resulting in human-readable output.
*/
type Consumer struct {
	indent       int
	inKey        bool
	hasKey       bool
	inValue      bool
	hasValue     bool
	contextStack []rune // Stack to keep track of nested structures
}

func NewConsumer() *Consumer {
	return &Consumer{
		indent:       0, // Initialize indentation to 0
		contextStack: []rune{},
	}
}

func (consumer *Consumer) Print(stream <-chan string) {
	for chunk := range stream {
		for _, char := range chunk {
			switch char {
			case '{', '[':
				consumer.handleOpenBracket(char)
			case '}', ']':
				consumer.handleCloseBracket(char)
			case '"':
				consumer.handleQuote()
			case ',':
				consumer.handleComma()
			case ':':
				consumer.handleColon()
			case '\n', '\r', '\t':
				// Ignore whitespace characters within JSON
			default:
				if consumer.inValue {
					fmt.Print(utils.Green(string(char)))
				} else if consumer.inKey {
					fmt.Print(utils.Blue(string(char)))
				} else if consumer.hasValue {
					consumer.resetFlags()
					fmt.Print(strings.Repeat("\t", consumer.indent))
				} else {
					fmt.Print(string(char))
				}
			}
		}
	}
}

func (consumer *Consumer) handleOpenBracket(char rune) {
	fmt.Println(string(char))
	consumer.indent++
	consumer.contextStack = append(consumer.contextStack, char)
	fmt.Print(strings.Repeat("\t", consumer.indent))
}

func (consumer *Consumer) handleCloseBracket(char rune) {
	consumer.indent--
	if len(consumer.contextStack) > 0 {
		consumer.contextStack = consumer.contextStack[:len(consumer.contextStack)-1]
	}
	fmt.Println()
	fmt.Print(strings.Repeat("\t", consumer.indent))
	consumer.resetFlags()
}

func (consumer *Consumer) handleQuote() {
	if !consumer.inKey && !consumer.hasKey && !consumer.inValue && !consumer.hasValue {
		consumer.inKey = true
	} else if consumer.inKey {
		consumer.hasKey = true
	} else if consumer.hasKey && !consumer.inValue {
		consumer.inValue = true
	}
}

func (consumer *Consumer) handleComma() {
	if consumer.hasValue {
		consumer.resetFlags()
		fmt.Println(",")
		fmt.Print(strings.Repeat("\t", consumer.indent))
	}
}

func (consumer *Consumer) handleColon() {
	if consumer.hasKey {
		fmt.Print(": ")
		consumer.inValue = true
	}
}

func (consumer *Consumer) resetFlags() {
	consumer.inKey = false
	consumer.hasKey = false
	consumer.inValue = false
	consumer.hasValue = false
}
