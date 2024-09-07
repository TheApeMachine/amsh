package ai

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewTeam(t *testing.T) {
	Convey("Given a map of Agents", t, func() {
		agents := map[string]*Agent{
			"coder":    NewAgent(NewConn(), CODER, nil, "test"),
			"reviewer": NewAgent(NewConn(), REVIEWER, nil, "test"),
		}

		Convey("When creating a new Team", func() {
			team := NewTeam(agents, "test")

			Convey("It should not be nil", func() {
				So(team, ShouldNotBeNil)
			})

			Convey("It should have the correct number of agents", func() {
				So(len(team.agents), ShouldEqual, 2)
			})

			Convey("It should contain the correct agents", func() {
				So(team.agents["coder"], ShouldEqual, agents["coder"])
				So(team.agents["reviewer"], ShouldEqual, agents["reviewer"])
			})
		})
	})
}
