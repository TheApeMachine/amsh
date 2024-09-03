package ai

import (
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewConn(t *testing.T) {
	Convey("Given an OpenAI API key in the environment", t, func() {
		originalAPIKey := os.Getenv("OPENAI_API_KEY")
		os.Setenv("OPENAI_API_KEY", "test-api-key")
		defer os.Setenv("OPENAI_API_KEY", originalAPIKey)

		Convey("When creating a new Conn", func() {
			conn := NewConn()

			Convey("It should not be nil", func() {
				So(conn, ShouldNotBeNil)
			})

			Convey("It should have a client", func() {
				So(conn.client, ShouldNotBeNil)
			})

			Convey("The client should be an instance of openai.Client", func() {
				So(conn.client, ShouldHaveSameTypeAs, &openai.Client{})
			})
		})
	})
}
