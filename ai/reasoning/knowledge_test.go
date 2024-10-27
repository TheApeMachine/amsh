package reasoning

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/amsh/ai/types"
)

func TestKnowledgeBase(t *testing.T) {
	Convey("Given a knowledge base", t, func() {
		kb := NewKnowledgeBase()

		Convey("When adding a fact", func() {
			fact := types.LogicalExpression{
				Operation:  types.AND,
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
			rule := types.LogicalExpression{
				Operation: types.IF,
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
			expr := types.LogicalExpression{
				Operation: types.AND,
				Operands: []interface{}{
					"Socrates is mortal",
				},
				Confidence: 1.0,
			}

			err := kb.ValidateExpression(expr)

			Convey("Then it should not return an error", func() {
				So(err, ShouldBeNil)
				So(err, ShouldBeNil)
			})
		})
	})
}
