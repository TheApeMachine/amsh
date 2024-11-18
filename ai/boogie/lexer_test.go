package boogie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

/*
TestLexer checks that the lexer correctly tokenizes Boogie programs.
*/
func TestLexer(t *testing.T) {
	Convey("Given an empty Boogie program ", t, func() {
		program := `out <= () <= in`

		Convey("When the program is tokenized", func() {
			lexemes := make([]Lexeme, 0)
			lexer := NewLexer(program)

			for lexeme := range lexer.Generate() {
				lexemes = append(lexemes, lexeme)
			}

			So(lexemes, ShouldResemble, []Lexeme{
				{ID: VALUE, Text: "out"},
				{ID: FLOW, Text: "<="},
				{ID: DELIMITER, Text: "("},
				{ID: DELIMITER, Text: ")"},
				{ID: FLOW, Text: "<="},
				{ID: VALUE, Text: "in"},
			})
		})
	})

	Convey("Given a single operation Boogie program ", t, func() {
		program := `out <= (analyze => send) <= in`

		Convey("When the program is tokenized", func() {
			lexemes := make([]Lexeme, 0)
			lexer := NewLexer(program)

			for lexeme := range lexer.Generate() {
				lexemes = append(lexemes, lexeme)
			}

			So(lexemes, ShouldResemble, []Lexeme{
				{ID: VALUE, Text: "out"},
				{ID: FLOW, Text: "<="},
				{ID: DELIMITER, Text: "("},
				{ID: OPERATION, Text: "analyze"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "send"},
				{ID: DELIMITER, Text: ")"},
				{ID: FLOW, Text: "<="},
				{ID: VALUE, Text: "in"},
			})
		})
	})

	Convey("Given a multi operation Boogie program ", t, func() {
		program := `out <= (
		    analyze => next
			verify  => send
		) <= in`

		Convey("When the program is tokenized", func() {
			lexemes := make([]Lexeme, 0)
			lexer := NewLexer(program)

			for lexeme := range lexer.Generate() {
				lexemes = append(lexemes, lexeme)
			}

			So(lexemes, ShouldResemble, []Lexeme{
				{ID: VALUE, Text: "out"},
				{ID: FLOW, Text: "<="},
				{ID: DELIMITER, Text: "("},
				{ID: OPERATION, Text: "analyze"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "next"},
				{ID: OPERATION, Text: "verify"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "send"},
				{ID: DELIMITER, Text: ")"},
				{ID: FLOW, Text: "<="},
				{ID: VALUE, Text: "in"},
			})
		})
	})
}
