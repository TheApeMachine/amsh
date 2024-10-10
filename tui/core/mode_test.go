package core

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// ... existing code ...

func TestNormalMode(t *testing.T) {
	Convey("Normal mode tests", t, func() {
		context := &Context{}
		normal := &Normal{}

		Convey("Entering normal mode", func() {
			normal.Enter(context)
			So(normal.context, ShouldEqual, context)
			So(normal.bufferEvents, ShouldNotBeNil)
		})

		Convey("Exiting normal mode", func() {
			normal.Exit()
			So(normal.bufferEvents, ShouldBeNil)
		})

		Convey("Running normal mode", func() {
			normal.Run()
			// Add assertions for Run behavior if applicable
		})
	})
}

func TestInsertMode(t *testing.T) {
	Convey("Insert mode tests", t, func() {
		context := &Context{}
		insert := &Insert{}

		Convey("Entering insert mode", func() {
			insert.Enter(context)
			So(insert.context, ShouldEqual, context)
		})

		Convey("Exiting insert mode", func() {
			insert.Exit()
			// Add assertions for Exit behavior if applicable
		})

		Convey("Running insert mode", func() {
			insert.Run()
			// Add assertions for Run behavior if applicable
		})
	})
}

func TestCommandMode(t *testing.T) {
	Convey("Command mode tests", t, func() {
		context := &Context{}
		command := &Command{}

		Convey("Entering command mode", func() {
			command.Enter(context)
			So(command.context, ShouldEqual, context)
		})

		Convey("Exiting command mode", func() {
			command.Exit()
			// Add assertions for Exit behavior if applicable
		})

		Convey("Running command mode", func() {
			command.Run()
			// Add assertions for Run behavior if applicable
		})
	})
}
