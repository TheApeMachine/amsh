package service

import (
	"github.com/gofiber/fiber/v3"
	"github.com/theapemachine/amsh/data"
)

type Inbound struct {
	MessageID    string `json:"message_id"`
	TicketID     string `json:"ticket_id"`
	ContactID    string `json:"contact_id"`
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email"`
	Message      string `json:"message"`
}

func (https *HTTPS) NewWebhook(origin, scope string) fiber.Handler {
	return func(ctx fiber.Ctx) (err error) {
		message := data.New("webhook", "trengo", "managing", ctx.Body())
		message.Poke("chain", "trengo")

		return ctx.SendStatus(fiber.StatusOK)
	}
}
