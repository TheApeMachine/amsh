//go:build integration
// +build integration

package ai

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAIIntegration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	Convey("Given a real OpenAI API connection", t, func() {
		conn := &Conn{
			client: openai.NewClient(apiKey),
		}

		Convey("And a Team with real Agents", func() {
			agents := map[string]*Agent{
				"coder":    NewAgent(conn, CODER, nil, "coder"),
				"reviewer": NewAgent(conn, REVIEWER, nil, "reviewer"),
			}

			team := NewTeam(agents, "reviewer")

			Convey("When executing a team task", func() {
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()

				result, err := team.ExecuteTeamTask(ctx, "Write a simple 'Hello, World!' program in Python")

				Convey("Then the task should be executed successfully", func() {
					So(err, ShouldBeNil)
					So(result, ShouldNotBeNil)

					Convey("And the result should be valid JSON", func() {
						_, err := json.Marshal(result)
						So(err, ShouldBeNil)
					})
				})
			})
		})
	})
}
