package boogie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

/*
TestLexer checks that the lexer correctly tokenizes Boogie programs.
*/
func TestLexer(t *testing.T) {
	Convey("Given a Boogie program with an empty closure", t, func() {
		program := `out <= () <= in`

		Convey("When the program is tokenized", func() {
			lexemes := make([]Lexeme, 0)
			lexer := NewLexer(program).Initialize()

			for lexeme := range lexer.Generate() {
				lexemes = append(lexemes, lexeme)
			}

			So(lexemes, ShouldResemble, []Lexeme{
				{ID: STATEMENT, Text: "out"},
				{ID: FLOWIN, Text: "<="},
				{ID: CLOSURE, Text: "("},
				{ID: CLOSURE, Text: ")"},
				{ID: FLOWIN, Text: "<="},
				{ID: STATEMENT, Text: "in"},
			})
		})
	})
}
