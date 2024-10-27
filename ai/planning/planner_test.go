package planning

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPlanner(t *testing.T) {
	Convey("Given a new Planner", t, func() {
		planner := NewPlanner()

		Convey("When creating a new plan", func() {
			now := time.Now()
			endTime := now.Add(24 * time.Hour)
			req := CreatePlanRequest{
				Name:        "Test Plan",
				Description: "Test Description",
				EndTime:     endTime,
				Goals: []CreateGoalRequest{
					{
						Name:        "Goal 1",
						Description: "First Goal",
						Priority:    1,
						Deadline:    endTime,
						Objectives: []CreateObjectiveRequest{
							{
								Name:        "Objective 1",
								Description: "First Objective",
								Deadline:    endTime,
							},
						},
					},
				},
			}

			plan, err := planner.CreatePlan(context.Background(), req)

			Convey("Then the plan should be created successfully", func() {
				So(err, ShouldBeNil)
				So(plan, ShouldNotBeNil)
				So(plan.ID, ShouldNotBeEmpty)
				So(plan.Name, ShouldEqual, "Test Plan")
				So(plan.Description, ShouldEqual, "Test Description")
				So(plan.Status, ShouldEqual, PlanStatusCreated)
				So(plan.Goals, ShouldHaveLength, 1)

				goal := plan.Goals[0]
				So(goal.Name, ShouldEqual, "Goal 1")
				So(goal.Priority, ShouldEqual, 1)
				So(goal.Status, ShouldEqual, GoalStatusPending)
				So(goal.Objectives, ShouldHaveLength, 1)

				objective := goal.Objectives[0]
				So(objective.Name, ShouldEqual, "Objective 1")
				So(objective.Status, ShouldEqual, ObjectiveStatusPending)
			})
		})

		Convey("When updating a plan", func() {
			// First create a plan
			plan, _ := planner.CreatePlan(context.Background(), CreatePlanRequest{
				Name:    "Test Plan",
				EndTime: time.Now().Add(24 * time.Hour),
			})

			Convey("Then updating a non-existent plan should fail", func() {
				err := planner.UpdatePlan(context.Background(), "non-existent-id", PlanUpdates{})
				So(err, ShouldEqual, ErrPlanNotFound)
			})

			Convey("Then updating an existing plan should succeed", func() {
				updates := PlanUpdates{
					TaskUpdates: []TaskUpdate{
						{
							TaskID:   "task-1",
							Progress: 0.5,
							Status:   TaskStatusActive,
						},
					},
				}
				err := planner.UpdatePlan(context.Background(), plan.ID, updates)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestResourcePool(t *testing.T) {
	Convey("Given a new ResourcePool", t, func() {
		rp := NewResourcePool()

		Convey("When setting and getting available resources", func() {
			rp.SetAvailable("cpu", 100.0)
			available := rp.GetAvailable("cpu")

			Convey("Then the values should match", func() {
				So(available, ShouldEqual, 100.0)
			})

			Convey("Then non-existent resources should return zero", func() {
				memory := rp.GetAvailable("memory")
				So(memory, ShouldEqual, 0.0)
			})
		})

		Convey("When accessing resources concurrently", func() {
			done := make(chan bool)
			go func() {
				rp.SetAvailable("cpu", 100.0)
				done <- true
			}()
			go func() {
				_ = rp.GetAvailable("cpu")
				done <- true
			}()

			Convey("Then it should not panic", func() {
				<-done
				<-done
			})
		})
	})
}
