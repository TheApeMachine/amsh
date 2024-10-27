package system

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessManager(t *testing.T) {
	Convey("Given a new ProcessManager", t, func() {
		pm := NewProcessManager(&Architecture{})

		Convey("When registering a new process", func() {
			handler := func(ctx context.Context, input interface{}) (interface{}, error) {
				return "processed", nil
			}
			err := pm.RegisterProcess("test", "test process", []string{"team1"}, handler)

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)
			})

			Convey("And when starting the process", func() {
				result, err := pm.StartProcess(context.Background(), "test", "input")

				Convey("Then it should execute successfully", func() {
					So(err, ShouldBeNil)
					So(result, ShouldEqual, "processed")
				})
			})

			Convey("And when registering the same process again", func() {
				err := pm.RegisterProcess("test", "test process", []string{"team1"}, handler)

				Convey("Then it should fail", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "already registered")
				})
			})
		})

		Convey("When starting a non-existent process", func() {
			result, err := pm.StartProcess(context.Background(), "nonexistent", "input")

			Convey("Then it should fail", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "not found")
				So(result, ShouldBeNil)
			})
		})
	})
}
