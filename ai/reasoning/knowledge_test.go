package reasoning

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestKnowledgeBase(t *testing.T) {
	Convey("Given a knowledge base", t, func() {
		kb := NewKnowledgeBase()

		Convey("When adding a fact", func() {
			fact := LogicalExpression{
				Operation:  AND,
				Operands:   []interface{}{"Socrates", "is human"},
				Confidence: 1.0,
			}
			kb.AddFact("socrates_human", fact)

			Convey("Then it should be retrievable", func() {
				stored, exists := kb.facts["socrates_human"]
				So(exists, ShouldBeTrue)
				So(stored, ShouldResemble, fact)
			})
		})

		Convey("When adding a rule", func() {
			rule := LogicalExpression{
				Operation: IF,
				Operands: []interface{}{
					"IF human THEN mortal",
				},
				Confidence: 1.0,
			}
			kb.AddRule("mortality_rule", rule)

			Convey("Then it should be retrievable", func() {
				stored, exists := kb.rules["mortality_rule"]
				So(exists, ShouldBeTrue)
				So(stored, ShouldResemble, rule)
			})
		})

		Convey("When validating an expression", func() {
			expr := LogicalExpression{
				Operation: AND,
				Operands: []interface{}{
					"Socrates is mortal",
				},
				Confidence: 1.0,
			}

			err := kb.ValidateExpression(expr)

			Convey("Then it should not return an error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
