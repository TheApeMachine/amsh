package ai

import (
	"testing"
	"strings"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func init() {
	// Set up test configuration
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("../cmd/cfg")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func TestPrompt(t *testing.T) {
	Convey("Given a new Prompt", t, func() {
		prompt := NewPrompt()

		Convey("It should be empty", func() {
			So(prompt.Build(), ShouldBeEmpty)
		})

		Convey("When adding a role template", func() {
			prompt.AddRoleTemplate(CODER)

			Convey("It should contain the coder template", func() {
				expected := viper.GetString("prompt.template.role.coder") + "\n\n"
				So(prompt.Build(), ShouldEqual, expected)
			})
		})

		Convey("When adding a scratchpad", func() {
			context := "Test context"
			prompt.AddScratchpad(context)

			Convey("It should contain the scratchpad template with the context", func() {
				template := viper.GetString("prompt.template.scratchpad")
				expected := strings.Replace(template, "{context}", context, 1) + "\n\n"
				So(prompt.Build(), ShouldEqual, expected)
			})
		})

		Convey("When adding content", func() {
			contentType := "readme"
			content := "Test content"
			prompt.AddContent(contentType, content)

			Convey("It should contain the content template with the content", func() {
				template := viper.GetString("prompt.template.content.readme")
				expected := strings.Replace(template, "{readme}", content, 1) + "\n\n"
				So(prompt.Build(), ShouldEqual, expected)
			})
		})

		Convey("When adding instructions", func() {
			prompt.AddInstructions()

			Convey("It should contain the instructions template", func() {
				expected := viper.GetString("prompt.template.instructions") + "\n\n"
				So(prompt.Build(), ShouldEqual, expected)
			})
		})
	})
}

func TestGetRoleString(t *testing.T) {
	Convey("Given different RoleTypes", t, func() {
		Convey("CODER should return 'coder'", func() {
			So(getRoleString(CODER), ShouldEqual, "coder")
		})

		Convey("REVIEWER should return 'reviewer'", func() {
			So(getRoleString(REVIEWER), ShouldEqual, "reviewer")
		})

		Convey("Unknown role should return 'unknown'", func() {
			So(getRoleString(RoleType(999)), ShouldEqual, "unknown")
		})
	})
}