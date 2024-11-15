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
			So(program.Root.Type, ShouldEqual, "group")
			So(program.Root.Children, ShouldHaveLength, 2)

			// Verify first operation: analyze
			analyzeOp := program.Root.Children[0]
			So(analyzeOp.Type, ShouldEqual, "analyze")
			So(analyzeOp.Behavior, ShouldEqual, "<temporal>")
			So(analyzeOp.Parameters, ShouldResemble, []string{"query", "\"data\""})
			So(analyzeOp.Outcomes, ShouldResemble, []string{"next", "cancel"})

			// Verify second operation: call
			callOp := program.Root.Children[1]
			So(callOp.Type, ShouldEqual, "call")
			So(callOp.Behavior, ShouldEqual, "<browser>")
			So(callOp.Parameters, ShouldResemble, []string{"query", "\"data\""})
			So(callOp.Outcomes, ShouldResemble, []string{"send", "back"})
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
			So(program.Input, ShouldEqual, "in")
			So(program.Output, ShouldEqual, "out")
			So(program.Root, ShouldNotBeNil)
			So(program.Root.Type, ShouldEqual, "group")
			So(program.Root.Children, ShouldHaveLength, 2)

			// Verify switch node
			switchNode := program.Root.Children[0]
			So(switchNode.Type, ShouldEqual, "switch")
			So(switchNode.Label, ShouldEqual, "[flow]")
			So(switchNode.Children, ShouldHaveLength, 2)

			// Verify 'success => next'
			successOp := switchNode.Children[0]
			So(successOp.Type, ShouldEqual, "success")
			So(successOp.Outcomes, ShouldResemble, []string{"next"})

			// Verify 'error => [error.jump]'
			errorOp := switchNode.Children[1]
			So(errorOp.Type, ShouldEqual, "error")
			So(errorOp.Outcomes, ShouldResemble, []string{"[error.jump]"})

			// Verify select node
			selectNode := program.Root.Children[1]
			So(selectNode.Type, ShouldEqual, "select")
			So(selectNode.Children, ShouldHaveLength, 1)

			// Verify match node within select
			matchNode := selectNode.Children[0]
			So(matchNode.Type, ShouldEqual, "match")
			So(matchNode.Children, ShouldHaveLength, 2)

			// Verify '<5> => next'
			retryOp := matchNode.Children[0]
			So(retryOp.Type, ShouldEqual, "<5>")
			So(retryOp.Outcomes, ShouldResemble, []string{"next"})

			// Verify 'default => back'
			defaultOp := matchNode.Children[1]
			So(defaultOp.Type, ShouldEqual, "default")
			So(defaultOp.Outcomes, ShouldResemble, []string{"back"})
		})

		Convey("When parsing a complex program with multiple features", func() {
			input := `
				; This demonstrates a complex workflow
				out <= (
					switch[mainflow] <= (
						; Initial analysis phase
						[data, "complex query"]analyze<temporal> => next | back | cancel

						; Processing phase with retries
						success => (
							select <= (
								[browser.settings]call<browser> => send | back
								match <= (
									<5> => next    ; Retry up to 5 times
									default => [mainflow.jump]
								)
								join <= (
									[results]process<parallel> => send | back
									validate => next | cancel
								)
							)
						)

						; Error handling
						error => join <= (
							[error.log]validate => next | cancel
							switch[errorflow] <= (
								retry => [mainflow.jump]
								fail => [error.handler.jump]
							)
						)

						; Default fallback
						default => [error.handler.jump]
					)
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

			// Verify main switch structure
			mainSwitch := program.Root.Children[0]
			So(mainSwitch.Type, ShouldEqual, "switch")
			So(mainSwitch.Label, ShouldEqual, "[mainflow]")
			So(mainSwitch.Children, ShouldHaveLength, 4) // analyze, success, error, default

			// Verify analyze operation
			analyzeOp := mainSwitch.Children[0]
			So(analyzeOp.Type, ShouldEqual, "analyze")
			So(analyzeOp.Behavior, ShouldEqual, "<temporal>")
			So(analyzeOp.Parameters, ShouldResemble, []string{"data", "\"complex query\""})
			So(analyzeOp.Outcomes, ShouldResemble, []string{"next", "back", "cancel"})

			// Verify success branch with select
			successBranch := mainSwitch.Children[1]
			So(successBranch.Type, ShouldEqual, "success")
			So(successBranch.Children, ShouldHaveLength, 1)

			selectNode := successBranch.Children[0]
			So(selectNode.Type, ShouldEqual, "select")
			So(selectNode.Children, ShouldHaveLength, 3) // call, match, join

			// Verify browser call
			browserCall := selectNode.Children[0]
			So(browserCall.Type, ShouldEqual, "call")
			So(browserCall.Behavior, ShouldEqual, "<browser>")
			So(browserCall.Parameters, ShouldResemble, []string{"browser.settings"})
			So(browserCall.Outcomes, ShouldResemble, []string{"send", "back"})

			// Verify match with retries
			matchNode := selectNode.Children[1]
			So(matchNode.Type, ShouldEqual, "match")
			So(matchNode.Children, ShouldHaveLength, 2)

			retryOp := matchNode.Children[0]
			So(retryOp.Type, ShouldEqual, "<5>")
			So(retryOp.Outcomes, ShouldResemble, []string{"next"})

			defaultMatchOp := matchNode.Children[1]
			So(defaultMatchOp.Type, ShouldEqual, "default")
			So(defaultMatchOp.Outcomes, ShouldResemble, []string{"[mainflow.jump]"})

			// Verify join in select
			joinNode := selectNode.Children[2]
			So(joinNode.Type, ShouldEqual, "join")
			So(joinNode.Children, ShouldHaveLength, 2)

			// Verify process operation
			processOp := joinNode.Children[0]
			So(processOp.Type, ShouldEqual, "process")
			So(processOp.Behavior, ShouldEqual, "<parallel>")
			So(processOp.Parameters, ShouldResemble, []string{"results"})
			So(processOp.Outcomes, ShouldResemble, []string{"send", "back"})

			// Verify validate operation
			validateOp := joinNode.Children[1]
			So(validateOp.Type, ShouldEqual, "validate")
			So(validateOp.Outcomes, ShouldResemble, []string{"next", "cancel"})

			// Verify error handling
			errorBranch := mainSwitch.Children[2]
			So(errorBranch.Type, ShouldEqual, "error")
			So(errorBranch.Children, ShouldHaveLength, 1)

			errorJoin := errorBranch.Children[0]
			So(errorJoin.Type, ShouldEqual, "join")
			So(errorJoin.Children, ShouldHaveLength, 2)

			// Verify validate operation in error join
			errorValidateOp := errorJoin.Children[0]
			So(errorValidateOp.Type, ShouldEqual, "validate")
			So(errorValidateOp.Parameters, ShouldResemble, []string{"error.log"})
			So(errorValidateOp.Outcomes, ShouldResemble, []string{"next", "cancel"})

			// Verify switch within error join
			errorSwitch := errorJoin.Children[1]
			So(errorSwitch.Type, ShouldEqual, "switch")
			So(errorSwitch.Label, ShouldEqual, "[errorflow]")
			So(errorSwitch.Children, ShouldHaveLength, 2)

			// Verify retry jump
			retryJump := errorSwitch.Children[0]
			So(retryJump.Type, ShouldEqual, "retry")
			So(retryJump.Outcomes, ShouldResemble, []string{"[mainflow.jump]"})

			// Verify fail jump
			failJump := errorSwitch.Children[1]
			So(failJump.Type, ShouldEqual, "fail")
			So(failJump.Outcomes, ShouldResemble, []string{"[error.handler.jump]"})

			// Verify default fallback
			defaultFallback := mainSwitch.Children[3]
			So(defaultFallback.Type, ShouldEqual, "default")
			So(defaultFallback.Outcomes, ShouldResemble, []string{"[error.handler.jump]"})
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
				{
					input: "out <= switch <= () <= in",
					err:   "Expected at least one operation in switch block",
				},
				{
					input: "out <= join <= () <= in",
					err:   "Expected at least one operation in join block",
				},
				{
					input: "out <= match[invalid] <= () <= in",
					err:   "Expected ARROW '<=' at position 2",
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
