package reasoning

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/amsh/ai/types"
)

func TestEngine(t *testing.T) {
	Convey("Given a reasoning engine", t, func() {
		validator := &Validator{
			chain:         &types.ReasoningChain{},
			knowledgeBase: NewKnowledgeBase(),
		}
		metaReasoner := NewMetaReasoner()

		// Change AddStrategy to RegisterStrategy
		metaReasoner.RegisterStrategy(MetaStrategy{
			Name:     "pattern_analysis",
			Priority: 1,
		})

		engine := NewEngine(validator, metaReasoner)

		Convey("When generating a reasoning step", func() {
			problem := "test problem"
			chain := &types.ReasoningChain{}

			step, err := engine.GenerateStep(context.Background(), problem, chain)

			Convey("Then it should create a valid step", func() {
				So(err, ShouldBeNil)
				So(step.Strategy, ShouldNotBeNil)
				So(step.Premise, ShouldNotBeNil)
				So(step.Conclusion, ShouldNotBeNil)
			})
		})
	})
}
