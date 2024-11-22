package boogie

import (
	"strings"
	"unicode"

	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
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
	"match",
	"join",
}

var values = []string{
	"out",
	"in",
	"ok",
	"error",
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
}

type TokenType uint

const (
	UNKNOWN TokenType = iota
	COMMENT
	DELIMITER
	OPERATION
	VALUE
	FLOW
	PARAMETER
)

type Lexeme struct {
	ID   TokenType
	Text string
}

type Lexer struct {
	buffer     strings.Builder
	inBehavior bool
	lexeme     bool
	state      TokenType
}

func NewLexer() *Lexer {
	return &Lexer{}
}

func (lexer *Lexer) Generate(source string) chan Lexeme {
	errnie.Info("lexing")

	out := make(chan Lexeme, 1024)

	go func() {
		defer close(out)

		for _, char := range source + " " {
			lexer.processChar(char)

			if lexer.lexeme && lexer.buffer.Len() > 0 {
				out <- Lexeme{ID: lexer.state, Text: lexer.buffer.String()}
				lexer.buffer.Reset()
				lexer.lexeme = false
				lexer.state = UNKNOWN
			}

			if !unicode.IsSpace(char) && lexer.state != COMMENT {
				lexer.buffer.WriteRune(char)
			}
		}
	}()

	return out
}

func (lexer *Lexer) GenerateStream(loadStream chan provider.Event) chan Lexeme {
	errnie.Log("lexer.GenerateStream()")

	out := make(chan Lexeme, 1024)

	go func() {
		defer close(out)

		for chunk := range loadStream {
			// Ignore the markdown tags
			if strings.HasPrefix(chunk.Content, "```boogie") || strings.HasPrefix(chunk.Content, "```") {
				continue
			}

			for _, char := range chunk.Content {
				lexer.processChar(char)

				if lexer.lexeme && lexer.buffer.Len() > 0 {
					out <- Lexeme{ID: lexer.state, Text: lexer.buffer.String()}
					lexer.buffer.Reset()
					lexer.lexeme = false
					lexer.state = UNKNOWN
				}

				if !unicode.IsSpace(char) && lexer.state != COMMENT {
					lexer.buffer.WriteRune(char)
				}
			}
		}
	}()

	return out
}

func (lexer *Lexer) processChar(char rune) {
	shouldReturn1 := lexer.handleComment(char)
	if shouldReturn1 {
		return
	}

	if char == '<' {
		lexer.lexeme, lexer.state = lexer.check(operations, OPERATION)
		return
	}

	if lexer.buffer.String() == "<" && char != '=' {
		lexer.inBehavior = true
		lexer.lexeme, lexer.state = true, DELIMITER
		return
	}

	if lexer.inBehavior {
		if char == '>' {
			lexer.inBehavior = false
			lexer.lexeme, lexer.state = true, VALUE
		}
		return
	}

	if char == ';' {
		lexer.lexeme, lexer.state = false, COMMENT
		return
	}

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

	lexer.handleSpace(char)
}

func (lexer *Lexer) handleComment(char rune) bool {
	errnie.Log("lexer.handleComment(%s) [state: %d]", string(char), lexer.state)

	if lexer.state == COMMENT && char != '\n' {
		return true
	}

	if lexer.state == COMMENT && char == '\n' {
		lexer.state = UNKNOWN
		return true
	}

	return false
}

func (lexer *Lexer) handleSpace(char rune) bool {
	errnie.Log("lexer.handleSpace(%s) [state: %d]", string(char), lexer.state)

	if unicode.IsSpace(char) {
		if lexer.lexeme, lexer.state = lexer.check(operations, OPERATION); lexer.lexeme {
			return true
		}

		if lexer.lexeme, lexer.state = lexer.check(values, VALUE); lexer.lexeme {
			return true
		}

		if lexer.lexeme, lexer.state = lexer.check(flows, FLOW); lexer.lexeme {
			return true
		}

		if lexer.lexeme, lexer.state = lexer.check(delimiters, DELIMITER); lexer.lexeme {
			return true
		}
	}

	return false
}

func (lexer *Lexer) check(set []string, token TokenType) (bool, TokenType) {
	if utils.ContainsAny(set, lexer.buffer.String()) {
		return true, token
	}

	return false, lexer.state
}
