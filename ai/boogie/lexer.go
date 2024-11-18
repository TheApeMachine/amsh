package boogie

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/theapemachine/amsh/utils"
)

var operations = []string{
	"analyze",
	"verify",
	"reason",
	"generate",
	"call",
	"send",
	"next",
	"back",
	"cancel",
	"halt",
}

var values = []string{
	"out",
	"in",
}

var flows = []string{
	"<=",
	"=>",
}

var delimiters = []string{
	"(",
	")",
	"[",
	"]",
	"{",
	"}",
	",",
	"|",
	"\n",
}

type TokenType uint

const (
	UNKNOWN TokenType = iota
	DELIMITER
	OPERATION
	VALUE
	BEHAVIOR
	FLOW
	PARAMETER
)

type Lexeme struct {
	ID   TokenType
	Text string
}

type Lexer struct {
	source string
	buffer strings.Builder
	lexeme bool
	state  TokenType
	memory *Lexeme
}

func NewLexer(source string) *Lexer {
	return &Lexer{source: source}
}

func (lexer *Lexer) Generate() chan Lexeme {
	out := make(chan Lexeme, 1024)

	go func() {
		defer close(out)

		for _, char := range lexer.source + " " {
			fmt.Println(lexer.state, lexer.buffer.String())

			lexer.processChar(char)

			if lexer.lexeme && lexer.buffer.Len() > 0 {
				out <- Lexeme{ID: lexer.state, Text: lexer.buffer.String()}
				lexer.buffer.Reset()
				lexer.lexeme = false
				lexer.state = UNKNOWN
			}

			if !unicode.IsSpace(char) {
				lexer.buffer.WriteRune(char)
			}
		}
	}()

	return out
}

func (lexer *Lexer) processChar(char rune) {
	// Check if current character is a delimiter
	if utils.ContainsAny(delimiters, string(char)) {
		// If we have content in buffer, process it before handling the delimiter
		if lexer.buffer.Len() > 0 {
			if ok, token := lexer.check(operations, OPERATION); ok {
				lexer.lexeme = true
				lexer.state = token
				return
			}
		}

		lexer.lexeme, lexer.state = true, DELIMITER
		return
	}

	if unicode.IsSpace(char) {
		if lexer.lexeme, lexer.state = lexer.check(operations, OPERATION); lexer.lexeme {
			return
		}

		if lexer.lexeme, lexer.state = lexer.check(values, VALUE); lexer.lexeme {
			return
		}

		if lexer.lexeme, lexer.state = lexer.check(flows, FLOW); lexer.lexeme {
			return
		}

		if lexer.lexeme, lexer.state = lexer.check(delimiters, DELIMITER); lexer.lexeme {
			return
		}
	}
}

func (lexer *Lexer) check(set []string, token TokenType) (bool, TokenType) {
	if utils.ContainsAny(set, lexer.buffer.String()) {
		return true, token
	}

	return false, lexer.state
}
