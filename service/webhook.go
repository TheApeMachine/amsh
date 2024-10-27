package service

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

type Inbound struct {
	MessageID    string `json:"message_id"`
	TicketID     string `json:"ticket_id"`
	ContactID    string `json:"contact_id"`
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email"`
	Message      string `json:"message"`
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
		if https.arch == nil {
			errnie.Error(fmt.Errorf("architecture not initialized"))
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}

		// Create data message
		message := data.New("webhook", origin, scope, ctx.Body())
		message.Poke("chain", origin)

		// Route to appropriate process based on origin
		switch origin {
		case "trengo":
			var payload Inbound
			if err := json.Unmarshal(ctx.Body(), &payload); err != nil {
				errnie.Error(err)
				return ctx.SendStatus(fiber.StatusBadRequest)
			}

			// Start helpdesk labelling process
			go func() {
				resultChan := https.arch.ProcessManager.HandleProcess(
					ctx.Context(),
					"helpdesk_labelling",
					payload,
				)
				if resultChan == nil {
					errnie.Error(fmt.Errorf("process result channel not found"))
					return
				}
				for result := range resultChan {
					errnie.Debug(fmt.Sprintf("Process result: %v", result))
				}
			}()

		case "github":
			var payload GitHubReviewPayload
			if err := json.Unmarshal(ctx.Body(), &payload); err != nil {
				errnie.Error(err)
				return ctx.SendStatus(fiber.StatusBadRequest)
			}

			// Start code review process
			go func() {
				resultChan := https.arch.ProcessManager.HandleProcess(
					ctx.Context(),
					"code_review",
					payload,
				)
				if resultChan == nil {
					errnie.Error(fmt.Errorf("process result channel not found"))
					return
				}
				for result := range resultChan {
					errnie.Debug(fmt.Sprintf("Process result: %v", result))
				}
			}()
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
	https.queue.PubCh <- *message

	return ctx.SendStatus(fiber.StatusOK)
}
