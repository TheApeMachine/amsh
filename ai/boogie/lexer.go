package boogie

import (
	"fmt"
	"strings"
	"unicode"
)

type LexerContext uint

const (
	NONE LexerContext = iota
	OUTERCTX
	COMMENTCTX
	CLOSURECTX
	LABELCTX
	PARAMETERCTX
	BEHAVIORCTX
)

type LexerState uint

const (
	UNKNOWN LexerState = iota
	COMMENT
	FLOWIN
	FLOWOUT
	CLOSURE
	STATEMENT
	LABEL
	PARAMETER
	BEHAVIOR
	EOF
)

var delimiters = map[string]LexerContext{
	";":  COMMENTCTX,
	"(":  CLOSURECTX,
	")":  CLOSURECTX,
	"[":  LABELCTX,
	"]":  LABELCTX,
	"<":  BEHAVIORCTX,
	">":  BEHAVIORCTX,
	"{":  PARAMETERCTX,
	"}":  PARAMETERCTX,
	"\n": NONE,
}

var keywords = map[string]LexerState{
	"out": STATEMENT,
	"in":  STATEMENT,
}

type Lexeme struct {
	ID   LexerState
	Text string
}

type Lexer struct {
	source   string
	ctx      LexerContext
	state    LexerState
	buffer   strings.Builder
	handlers map[LexerContext]func(char rune) (LexerContext, LexerState)
	lexeme   bool
}

func NewLexer(source string) *Lexer {
	return &Lexer{source: source}
}

func (lexer *Lexer) Initialize() *Lexer {
	lexer.handlers = map[LexerContext]func(char rune) (LexerContext, LexerState){
		NONE:         lexer.noneHandler,
		COMMENTCTX:   lexer.commentHandler,
		CLOSURECTX:   lexer.closureHandler,
		LABELCTX:     lexer.labelHandler,
		PARAMETERCTX: lexer.parameterHandler,
		BEHAVIORCTX:  lexer.behaviorHandler,
		OUTERCTX:     lexer.outerHandler,
	}

	return lexer
}

func (lexer *Lexer) ctxToString() string {
	switch lexer.ctx {
	case OUTERCTX:
		return "OUTER"
	case COMMENTCTX:
		return "COMMENT"
	case CLOSURECTX:
		return "CLOSURE"
	case LABELCTX:
		return "LABEL"
	case PARAMETERCTX:
		return "PARAMETER"
	case BEHAVIORCTX:
		return "BEHAVIOR"
	default:
		return "NONE"
	}
}

func (lexer *Lexer) stateToString() string {
	switch lexer.state {
	case STATEMENT:
		return "STATEMENT"
	case FLOWIN:
		return "FLOWIN"
	case FLOWOUT:
		return "FLOWOUT"
	case CLOSURE:
		return "CLOSURE"
	case LABEL:
		return "LABEL"
	case PARAMETER:
		return "PARAMETER"
	case BEHAVIOR:
		return "BEHAVIOR"
	}

	return "UNKNOWN"
}

func (lexer *Lexer) Generate() chan Lexeme {
	out := make(chan Lexeme, 1024)

	go func() {
		defer close(out)

		for _, char := range lexer.source + "  " {
			fmt.Println(lexer.ctxToString(), lexer.stateToString(), lexer.buffer.String())

			if lexer.lexeme {
				out <- Lexeme{ID: lexer.state, Text: lexer.buffer.String()}
				lexer.buffer.Reset()
				lexer.lexeme = false
				lexer.state = UNKNOWN
			}

			lexer.ctx, lexer.state = lexer.handlers[lexer.ctx](char)

			if unicode.IsSpace(char) {
				continue
			}

			lexer.buffer.WriteRune(char)
		}
	}()

	return out
}

func (lexer *Lexer) noneHandler(char rune) (LexerContext, LexerState) {
	if unicode.IsSpace(char) {
		// Check keywords.
		if _, ok := keywords[lexer.buffer.String()]; ok {
			lexer.lexeme = true
			return lexer.ctx, lexer.state
		}

		if lexer.state == PARAMETER {
			lexer.lexeme = true
			return lexer.ctx, lexer.state
		}
	}

	// Check if the character is a delimiter.
	if ctx, ok := delimiters[string(char)]; ok {
		// Set the context that corresponds to the delimiter.
		lexer.ctx = ctx

		if char == '(' || char == ')' {
			lexer.lexeme = true
			return lexer.ctx, CLOSURE
		}
	}

	// If the character is a letter, we need to switch to the STATEMENT state.
	if unicode.IsLetter(char) {
		if lexer.state != STATEMENT {
			lexer.buffer.Reset()
		}

		// Check for keywords.
		if _, ok := keywords[lexer.buffer.String()]; ok {
			lexer.lexeme = true
			return lexer.ctx, lexer.state
		}

		lexer.ctx = OUTERCTX
		lexer.state = STATEMENT
	}

	return lexer.ctx, lexer.state
}

func (lexer *Lexer) outerHandler(char rune) (LexerContext, LexerState) {
	// Check for delimiters.
	if ctx, ok := delimiters[string(char)]; ok {
		if char == '(' {
			lexer.lexeme = true
		}

		lexer.ctx = ctx
	}

	// Check for keywords.
	if _, ok := keywords[lexer.buffer.String()]; ok {
		lexer.lexeme = true
		return lexer.ctx, STATEMENT
	}

	return lexer.ctx, lexer.state
}

func (lexer *Lexer) commentHandler(char rune) (LexerContext, LexerState) {
	// Once in a comment context, we need to wait for a newline character.
	if char == '\n' {
		return NONE, UNKNOWN
	}

	return lexer.ctx, lexer.state
}

func (lexer *Lexer) closureHandler(char rune) (LexerContext, LexerState) {
	if char == ')' {
		lexer.lexeme = true
		return NONE, CLOSURE
	}

	if unicode.IsLetter(char) {
		if _, ok := keywords[lexer.buffer.String()]; ok {
			lexer.lexeme = true
		}

		if lexer.state != STATEMENT {
			lexer.buffer.Reset()
		}

		lexer.state = STATEMENT
	}

	return lexer.ctx, lexer.state
}

func (lexer *Lexer) labelHandler(char rune) (LexerContext, LexerState) {
	return lexer.ctx, lexer.state
}

func (lexer *Lexer) parameterHandler(char rune) (LexerContext, LexerState) {
	if char == ',' || char == '}' {
		return NONE, PARAMETER
	}

	return lexer.ctx, lexer.state
}

func (lexer *Lexer) behaviorHandler(char rune) (LexerContext, LexerState) {
	switch char {
	case '>':
		return NONE, BEHAVIOR
	case '=':
		lexer.lexeme = true
		lexer.state = FLOWIN
		return NONE, FLOWIN
	}

	return lexer.ctx, lexer.state
}
