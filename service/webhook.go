package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v3"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/marvin"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/utils"
	"github.com/theapemachine/errnie"
)

type Inbound struct {
	MessageID    string `json:"message_id"`
	TicketID     string `json:"ticket_id"`
	ChannelID    string `json:"channel_id"`
	ContactID    string `json:"contact_id"`
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email"`
	Message      string `json:"message"`
	EventType    string `json:"event_type"`
}

type GitHubReviewPayload struct {
	Action      string `json:"action"`
	PullRequest struct {
		Number int `json:"number"`
		Head   struct {
			Ref string `json:"ref"`
		} `json:"head"`
	} `json:"pull_request"`
	Review struct {
		State    string `json:"state"`
		Body     string `json:"body"`
		Comments []struct {
			Path     string `json:"path"`
			Position int    `json:"position"`
			Body     string `json:"body"`
		} `json:"comments"`
	} `json:"review"`
}

func (https *HTTPS) NewWebhook(origin, scope string) fiber.Handler {
	return func(ctx fiber.Ctx) (err error) {
		// Route to appropriate process based on origin
		switch origin {
		case "trengo":
			// Check if the payload is URL-encoded
			contentType := ctx.Get("Content-Type")
			var payload Inbound
			//var ticket []string

			if contentType == "application/x-www-form-urlencoded" {
				// Parse the body as a URL-encoded query string
				values := errnie.SafeMust(func() (url.Values, error) {
					return url.ParseQuery(string(ctx.Body()))
				})

				ticket := []string{
					fmt.Sprintf("message_id: %s", values.Get("message_id")),
					fmt.Sprintf("ticket_id: %s", values.Get("ticket_id")),
					fmt.Sprintf("message: %s", values.Get("message")),
					fmt.Sprintf("channel_id: %s", values.Get("channel_id")),
					fmt.Sprintf("contact_id: %s", values.Get("contact_id")),
					fmt.Sprintf("contact_name: %s", values.Get("contact_name")),
					fmt.Sprintf("contact_email: %s", values.Get("contact_email")),
					fmt.Sprintf("event_type: %s", values.Get("event_type")),
				}

				agent := marvin.NewAgent(ctx.Context(), "webhook", "helpdesk", data.New(
					"webhook",
					"helpdesk",
					"inbound",
					[]byte(viper.GetViper().GetString("ai.setups.marvin.templates.system")),
				))

				agent.Generate(data.New("webhook", "helpdesk", "inbound", []byte(utils.JoinWith("\n", ticket...))))
			} else {
				// Handle JSON payload (fallback or unexpected case)
				if err := json.Unmarshal(ctx.Body(), &payload); err != nil {
					errnie.Error(err)
					return ctx.SendStatus(fiber.StatusBadRequest)
				}
			}

			// Start helpdesk labelling process
		case "github":
			var payload GitHubReviewPayload
			if err := json.Unmarshal(ctx.Body(), &payload); err != nil {
				errnie.Error(err)
				return ctx.SendStatus(fiber.StatusBadRequest)
			}
		}

		return ctx.SendStatus(fiber.StatusOK)
	}
}

func (https *HTTPS) handleGitHubWebhook(ctx fiber.Ctx) error {
	event := ctx.Get("X-GitHub-Event")
	if event != "pull_request_review" {
		return ctx.SendStatus(fiber.StatusOK)
	}

	var payload GitHubReviewPayload
	if err := json.Unmarshal(ctx.Body(), &payload); err != nil {
		return err
	}

	// Create a message for the AI system to process the review
	message := data.New(
		"github_review",
		"github",
		"code_review",
		ctx.Body(),
	)
	message.Poke("chain", "github")

	return ctx.SendStatus(fiber.StatusOK)
}
