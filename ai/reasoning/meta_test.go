package reasoning

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMetaReasoner(t *testing.T) {
	Convey("Given a meta reasoner", t, func() {
		mr := NewMetaReasoner()

		// Initialize with default resources
		mr.InitializeResources(map[string]float64{
			"cpu":    1.0,
			"memory": 1.0,
			"time":   1.0,
		})

		// Register default strategies
		mr.RegisterStrategy(MetaStrategy{
			Name:     "deduction",
			Priority: 8,
			Resources: map[string]float64{
				"cpu":    0.5,
				"memory": 0.3,
			},
			Constraints: []string{"time_critical", "high_accuracy"},
		})

		mr.RegisterStrategy(MetaStrategy{
			Name:     "induction",
			Priority: 6,
			Resources: map[string]float64{
				"cpu":    0.3,
				"memory": 0.2,
			},
			Constraints: []string{"pattern_matching", "data_driven"},
		})

		Convey("When evaluating strategies", func() {
			strategy := &MetaStrategy{
				Name:     "deduction",
				Priority: 8,
				Resources: map[string]float64{
					"cpu":    0.5,
					"memory": 0.3,
				},
				Constraints: []string{"time_critical"},
			}
			problem := "Is Socrates mortal?"
			constraints := []string{"time_critical"}

			score := mr.evaluateStrategy(problem, strategy, constraints)

			Convey("Then it should return a valid score", func() {
				So(score, ShouldBeBetweenOrEqual, 0, 1)
			})
		})

		Convey("When checking resource allocation", func() {
			mr.resources["cpu"] = 1.0
			mr.resources["memory"] = 1.0

			strategy := &MetaStrategy{
				Resources: map[string]float64{
					"cpu":    0.5,
					"memory": 0.3,
				},
			}

			canAllocate := mr.canAllocateResources(strategy)

			Convey("Then it should confirm availability", func() {
				So(canAllocate, ShouldBeTrue)
			})
		})

		Convey("When allocating resources", func() {
			mr.resources["cpu"] = 1.0
			strategy := &MetaStrategy{
				Resources: map[string]float64{
					"cpu": 0.5,
				},
			}

			mr.AllocateResources(strategy)

			Convey("Then it should decrease available resources", func() {
				So(mr.resources["cpu"], ShouldEqual, 0.5)
			})

			Convey("And when releasing resources", func() {
				mr.ReleaseResources(strategy)

				Convey("Then it should restore available resources", func() {
					So(mr.resources["cpu"], ShouldEqual, 1.0)
				})
			})
		})
	})
}
