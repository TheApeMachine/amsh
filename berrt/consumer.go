package berrt

import (
	"fmt"
	"strings"

	"github.com/theapemachine/amsh/utils"
)

/*
Consumer is a specialized logging type which is purposely designed
to deal with streaming, chunked JSON strings. It strips away the
structural elements of the JSON, while adhering to the indentation
levels, leaving only human-readable ouput.
*/
type Consumer struct {
	indent   int
	inKey    bool
	hasKey   bool
	inValue  bool
	hasValue bool
}

func NewConsumer() *Consumer {
	return &Consumer{indent: -1}
}

func (consumer *Consumer) Print(stream <-chan string) {
	for chunk := range stream {
		for _, char := range chunk {
			switch char {
			case '{', '[':
				consumer.handleOpenBracket()
			case '}', ']':
				consumer.handleCloseBracket()
			case '"':
				consumer.handleQuote()
			case ',':
				// noop
			default:
				if consumer.inValue {
					fmt.Print(utils.Green(string(char)))
				} else if consumer.inKey {
					fmt.Print(utils.Blue(string(char)))
				} else if consumer.hasValue {
					consumer.inKey = false
					consumer.hasKey = false
					consumer.inValue = false
					consumer.hasValue = false
					fmt.Print(strings.Repeat("\t", consumer.indent))
				}
			}
		}
	}
}

func (consumer *Consumer) handleOpenBracket() {
	consumer.indent++
}

func (consumer *Consumer) handleCloseBracket() {
	consumer.indent--
}

func (consumer *Consumer) handleQuote() {
	if !consumer.inKey && !consumer.hasKey {
		consumer.inKey = true
	} else if consumer.inKey {
		consumer.hasKey = true
	} else if consumer.hasKey && !consumer.inValue {
		consumer.inValue = true
	} else if consumer.inValue {
		consumer.hasValue = true
	}
}
