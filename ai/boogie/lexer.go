package boogie

import (
	"strings"
	"unicode"
)

// TokenType represents the type of token.
type TokenType int

const (
	NONE TokenType = iota
	OUT
	IN
	OUTFLOW
	INFLOW
	DELIMITER
	IDENTIFIER
	KEYWORD
	LITERAL
	COMMENT
)

var keywords = []string{
	"in",
	"out",
	"analyze",
	"reason",
	"call",
	"next",
	"cancel",
	"send",
	"back",
	"join",
	"verify",
	"match",
	"ok",
	"default",
}

// Lexeme represents a lexical token.
type Lexeme struct {
	Type  TokenType
	Value string
}

// Lexer represents a state machine for lexing boogie code.
type Lexer struct {
	input  string
	buffer strings.Builder
	state  TokenType
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		buffer: strings.Builder{},
		state:  NONE,
	}
}

func (lexer *Lexer) Generate() chan Lexeme {
	out := make(chan Lexeme)

	go func() {
		defer close(out)
		var lexeme *Lexeme

		for _, char := range lexer.input {
			lexeme = nil

			if lexeme = lexer.checkBuffer(); lexeme != nil {
				out <- *lexeme
			}

			switch {
			case unicode.IsSpace(char):
				if lexeme = lexer.checkBuffer(); lexeme != nil {
					out <- *lexeme
				}
			case char == '(' || char == ')' || char == '|':
				if lexeme = lexer.checkBuffer(); lexeme != nil {
					out <- *lexeme
				}

				out <- Lexeme{Type: DELIMITER, Value: string(char)}
			case char == '<':
				if lexeme = lexer.checkBuffer(); lexeme != nil {
					out <- *lexeme
				}

				lexer.state = OUTFLOW
			case char == '=':
				if lexeme = lexer.checkFlow(OUTFLOW); lexeme != nil {
					out <- *lexeme
				}

				lexer.state = INFLOW
			case char == '>':
				if lexeme = lexer.checkFlow(INFLOW); lexeme != nil {
					out <- *lexeme
				}
			case char == '"':
				if lexer.state == LITERAL {
					out <- *lexer.makeLexeme()
					break
				}

				lexer.state = LITERAL
			}

			if !unicode.IsSpace(char) {
				lexer.buffer.WriteRune(char)
			}
		}
	}()

	return out
}

func (lexer *Lexer) makeLexeme() *Lexeme {
	out := Lexeme{Type: lexer.state, Value: lexer.buffer.String()}
	lexer.buffer.Reset()
	lexer.state = NONE
	return &out
}

func (lexer *Lexer) checkBuffer() *Lexeme {
	if lexer.buffer.Len() == 0 || lexer.state == NONE {
		return nil
	}

	// Check for keywords
	for _, keyword := range keywords {
		if lexer.buffer.String() == keyword {
			return lexer.makeLexeme()
		}
	}

	return lexer.makeLexeme()
}

func (lexer *Lexer) checkFlow(flowType TokenType) *Lexeme {
	if lexer.state != flowType {
		return nil
	}

	return lexer.makeLexeme()
}

func (lexer *Lexer) checkDelimiter() *Lexeme {
	if lexer.state != DELIMITER {
		return nil
	}

	return lexer.makeLexeme()
}
