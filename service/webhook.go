package service

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/spf13/viper"
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
		template := viper.GetViper().GetString("webhook." + origin + "." + scope)

		message := Inbound{}
		if err = ctx.Bind().Body(&message); err != nil {
			return err
		}

		template = strings.ReplaceAll(template, "{message_id}", message.MessageID)
		template = strings.ReplaceAll(template, "{ticket_id}", message.TicketID)
		template = strings.ReplaceAll(template, "{contact_id}", message.ContactID)
		template = strings.ReplaceAll(template, "{contact_name}", message.ContactName)
		template = strings.ReplaceAll(template, "{contact_email}", message.ContactEmail)
		template = strings.ReplaceAll(template, "{message}", message.Message)

		https.queue.Publish(data.New("trengo", "webhook", "broadcast", []byte(template)))

		return ctx.SendStatus(fiber.StatusOK)
	}
}
