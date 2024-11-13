package boogie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParser(t *testing.T) {
	Convey("Given a boogie parser", t, func() {
		Convey("When parsing a simple program", func() {
			input := `
				out <= (
					analyze<temporal> => next | cancel
					[query, "data"]call<browser> => send | back
				) <= in
			`
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)
			parser := NewParser(tokens)
			program, err := parser.Parse()

			So(err, ShouldBeNil)
			So(program, ShouldNotBeNil)
			So(program.Input, ShouldEqual, "in")
			So(program.Output, ShouldEqual, "out")
			So(program.Root, ShouldNotBeNil)
			So(program.Root.Children, ShouldHaveLength, 2)
		})

		Convey("When parsing control flow structures", func() {
			input := `
				out <= (
					switch[flow] <= (
						success => next
						error => [error.jump]
					)
					select <= (
						match <= (
							<5> => next
							default => back
						)
					)
				) <= in
			`
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)
			parser := NewParser(tokens)
			program, err := parser.Parse()

			So(err, ShouldBeNil)
			So(program, ShouldNotBeNil)
			So(program.Root.Children, ShouldHaveLength, 2)

			switchNode := program.Root.Children[0]
			So(switchNode.Type, ShouldEqual, "switch")
			So(switchNode.Label, ShouldEqual, "[flow]")
			So(switchNode.Children, ShouldHaveLength, 2)

			selectNode := program.Root.Children[1]
			So(selectNode.Type, ShouldEqual, "select")
			So(selectNode.Children, ShouldHaveLength, 1)

			matchNode := selectNode.Children[0]
			So(matchNode.Type, ShouldEqual, "match")
			So(matchNode.Children, ShouldHaveLength, 2)

			So(matchNode.Children[0].Type, ShouldEqual, "match")
			So(matchNode.Children[0].Children, ShouldHaveLength, 1)
			So(matchNode.Children[0].Children[0].Type, ShouldEqual, "<5>")
			So(matchNode.Children[0].Children[0].Children, ShouldHaveLength, 1)
			So(matchNode.Children[0].Children[0].Children[0].Type, ShouldEqual, "next")

			So(matchNode.Children[1].Type, ShouldEqual, "match")
			So(matchNode.Children[1].Children, ShouldHaveLength, 1)
			So(matchNode.Children[1].Children[0].Type, ShouldEqual, "default")
			So(matchNode.Children[1].Children[0].Children, ShouldHaveLength, 1)
			So(matchNode.Children[1].Children[0].Children[0].Type, ShouldEqual, "back")
		})

		Convey("When parsing invalid programs", func() {
			cases := []struct {
				input string
				err   string
			}{
				{
					input: "out analyze => next",
					err:   "Expected ARROW '<=' at position 1",
				},
				{
					input: "out <= analyze => next in",
					err:   "Expected ARROW '<=' at position 5",
				},
				{
					input: "out <= analyze => <= in",
					err:   "Expected outcome at position 4",
				},
			}

			for _, c := range cases {
				lexer := NewLexer(c.input)
				tokens := collectTokens(lexer)
				parser := NewParser(tokens)
				_, err := parser.Parse()

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, c.err)
			}
		})
	})
}
