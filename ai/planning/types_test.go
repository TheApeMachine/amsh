package planning

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTypes(t *testing.T) {
	Convey("Given plan status constants", t, func() {
		Convey("Then they should have correct string values", func() {
			So(string(PlanStatusCreated), ShouldEqual, "created")
			So(string(PlanStatusActive), ShouldEqual, "active")
			So(string(PlanStatusBlocked), ShouldEqual, "blocked")
			So(string(PlanStatusComplete), ShouldEqual, "complete")
			So(string(PlanStatusCancelled), ShouldEqual, "cancelled")
		})
	})

	Convey("Given a CreatePlanRequest", t, func() {
		now := time.Now()
		req := CreatePlanRequest{
			Name:        "Test Plan",
			Description: "Test Description",
			EndTime:     now,
			Goals: []CreateGoalRequest{
				{
					Name:        "Goal 1",
					Description: "Goal Description",
					Priority:    1,
					Deadline:    now,
					Objectives: []CreateObjectiveRequest{
						{
							Name:        "Objective 1",
							Description: "Objective Description",
							Deadline:    now,
						},
					},
				},
			},
		}

		Convey("Then it should have valid fields", func() {
			So(req.Name, ShouldEqual, "Test Plan")
			So(req.Description, ShouldEqual, "Test Description")
			So(req.EndTime, ShouldEqual, now)
			So(req.Goals, ShouldHaveLength, 1)
			So(req.Goals[0].Name, ShouldEqual, "Goal 1")
			So(req.Goals[0].Objectives, ShouldHaveLength, 1)
			So(req.Goals[0].Objectives[0].Name, ShouldEqual, "Objective 1")
		})
	})

	Convey("Given generated IDs", t, func() {
		id1 := generateID()
		id2 := generateID()

		Convey("Then they should be unique", func() {
			So(id1, ShouldNotEqual, id2)
		})

		Convey("Then they should not be empty", func() {
			So(id1, ShouldNotBeEmpty)
			So(id2, ShouldNotBeEmpty)
		})
	})
}
