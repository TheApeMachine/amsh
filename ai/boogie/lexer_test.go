package boogie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

/*
TestLexer checks that the lexer correctly tokenizes Boogie programs.
*/
func TestLexer(t *testing.T) {
	Convey("Given a Boogie program with a simple closure", t, func() {
		program := `out <= () <= in`
		
		Convey("When the program is tokenized", func() {
			tokens := Tokenize(program)
			
			Convey("Then the tokens should match the expected sequence", func() {
				expectedTokens := []string{"out", "<=", "(", ")", "<=", "in"}
				So(tokens, ShouldResemble, expectedTokens)
			})
		})
	})

	Convey("Given a Boogie program with a simple operation", t, func() {
		program := `out <= (analyze => next) <= in`
		
		Convey("When the program is tokenized", func() {
			tokens := Tokenize(program)
			
			Convey("Then the tokens should match the expected sequence", func() {
				expectedTokens := []string{"out", "<=", "(", "analyze", "=>", "next", ")", "<=", "in"}
				So(tokens, ShouldResemble, expectedTokens)
			})
		})
	})

	Convey("Given Boogie programs with fallback chaining", t, func() {
		Convey("When tokenizing a single fallback", func() {
			program := `out <= (analyze => next | back) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{"out", "<=", "(", "analyze", "=>", "next", "|", "back", ")", "<=", "in"}
			So(tokens, ShouldResemble, expectedTokens)
		})

		Convey("When tokenizing multiple fallbacks", func() {
			program := `out <= (analyze => next | back | cancel) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{"out", "<=", "(", "analyze", "=>", "next", "|", "back", "|", "cancel", ")", "<=", "in"}
			So(tokens, ShouldResemble, expectedTokens)
		})

		Convey("When tokenizing multiple operations with fallbacks", func() {
			program := `out <= (
				analyze => next | back
				verify  => send | cancel
			) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{
				"out", "<=", "(",
				"analyze", "=>", "next", "|", "back",
				"verify", "=>", "send", "|", "cancel",
				")", "<=", "in",
			}
			So(tokens, ShouldResemble, expectedTokens)
		})
	})

	Convey("Given Boogie programs with match constructs", t, func() {
		Convey("When tokenizing operations followed by a match", func() {
			program := `out <= (
				analyze => next
				verify  => next
				match (
					ok    => send
					error => cancel
					_     => back
				)
			) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{
				"out", "<=", "(",
				"analyze", "=>", "next",
				"verify", "=>", "next",
				"match", "(",
				"ok", "=>", "send",
				"error", "=>", "cancel",
				"_", "=>", "back",
				")", ")", "<=", "in",
			}
			So(tokens, ShouldResemble, expectedTokens)
		})

		Convey("When tokenizing operations with match and fallbacks", func() {
			program := `out <= (
				analyze => next | back
				verify  => next | cancel
				match (
					ok    => send
					error => back
					_     => cancel
				)
			) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{
				"out", "<=", "(",
				"analyze", "=>", "next", "|", "back",
				"verify", "=>", "next", "|", "cancel",
				"match", "(",
				"ok", "=>", "send",
				"error", "=>", "back",
				"_", "=>", "cancel",
				")", ")", "<=", "in",
			}
			So(tokens, ShouldResemble, expectedTokens)
		})
	})

	Convey("Given Boogie programs with labels and jumps", t, func() {
		Convey("When tokenizing a labeled block", func() {
			program := `out <= (
				[myLabel] => (
					analyze => next
					verify  => next
				)
			) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{
				"out", "<=", "(",
				"[myLabel]", "=>", "(",
				"analyze", "=>", "next",
				"verify", "=>", "next",
				")", ")", "<=", "in",
			}
			So(tokens, ShouldResemble, expectedTokens)
		})

		Convey("When tokenizing a program with label jumps", func() {
			program := `out <= (
				[myLabel] => (
					analyze => next
					verify  => next
					match (
						ok    => send
						error => cancel
						_     => [myLabel].jump
					)
				)
			) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{
				"out", "<=", "(",
				"[myLabel]", "=>", "(",
				"analyze", "=>", "next",
				"verify", "=>", "next",
				"match", "(",
				"ok", "=>", "send",
				"error", "=>", "cancel",
				"_", "=>", "[myLabel].jump",
				")", ")", ")", "<=", "in",
			}
			So(tokens, ShouldResemble, expectedTokens)
		})
	})

	Convey("Given Boogie programs with behaviors", t, func() {
		Convey("When tokenizing analysis behaviors", func() {
			program := `out <= (
				analyze<surface> => next
				verify<deep>    => send
			) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{
				"out", "<=", "(",
				"analyze", "<surface>", "=>", "next",
				"verify", "<deep>", "=>", "send",
				")", "<=", "in",
			}
			So(tokens, ShouldResemble, expectedTokens)
		})

		Convey("When tokenizing reasoning behaviors", func() {
			program := `out <= (
				analyze<chainofthought> => next
				verify<selfcritique>    => send
				match (
					ok    => send
					error => cancel
					_     => back
				)
			) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{
				"out", "<=", "(",
				"analyze", "<chainofthought>", "=>", "next",
				"verify", "<selfcritique>", "=>", "send",
				"match", "(",
				"ok", "=>", "send",
				"error", "=>", "cancel",
				"_", "=>", "back",
				")", ")", "<=", "in",
			}
			So(tokens, ShouldResemble, expectedTokens)
		})

		Convey("When tokenizing fallback behaviors", func() {
			program := `out <= (
				analyze => next | back<3> | cancel
				verify  => send | back<3> | cancel
			) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{
				"out", "<=", "(",
				"analyze", "=>", "next", "|", "back", "<3>", "|", "cancel",
				"verify", "=>", "send", "|", "back", "<3>", "|", "cancel",
				")", "<=", "in",
			}
			So(tokens, ShouldResemble, expectedTokens)
		})

		Convey("When tokenizing iteration behaviors", func() {
			program := `out <= (
				analyze <= <3> => next
				verify         => next
			) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{
				"out", "<=", "(",
				"analyze", "<=", "<3>", "=>", "next",
				"verify", "=>", "next",
				")", "<=", "in",
			}
			So(tokens, ShouldResemble, expectedTokens)
		})
	})

	Convey("Given Boogie programs with parameterized behaviors", t, func() {
		Convey("When tokenizing a behavior with parameters", func() {
			program := `out <= (
				call<{
					search,
					"some query"
				} => browser> => send
			) <= in`
			tokens := Tokenize(program)
			expectedTokens := []string{
				"out", "<=", "(",
				"call", "<{", "search", ",", "\"some query\"", "}", "=>", "browser>", "=>", "send",
				")", "<=", "in",
			}
			So(tokens, ShouldResemble, expectedTokens)
		})
	})
}
