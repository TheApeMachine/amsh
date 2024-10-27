package system

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkloadManager(t *testing.T) {
	Convey("Given a new Architecture", t, func() {
		arch, err := NewArchitecture(context.Background(), "test")
		So(err, ShouldBeNil)

		Convey("When creating a new WorkloadManager", func() {
			manager, err := NewWorkloadManager(context.Background(), arch)

			Convey("Then it should initialize successfully", func() {
				So(err, ShouldBeNil)
				So(manager, ShouldNotBeNil)
				So(manager.architecture, ShouldEqual, arch)
			})
		})
	})
}
