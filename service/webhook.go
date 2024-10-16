package service

import (
	"github.com/gofiber/fiber/v3"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
)

var queue = twoface.NewQueue()

func NewWebhook() fiber.Handler {
	return func(ctx fiber.Ctx) (err error) {
		origin := ctx.Request().Header.Peek("Origin")
		queue.Publish(data.New(string(origin), "webhook", "message", ctx.Body()).Poke(
			"user", `
			[TASK]
				The following request has been received on the webhook channel.
			[/TASK]

			[REQUEST]
		      {request}
			[/REQUEST]
			`,
		))

		return ctx.SendStatus(fiber.StatusOK)
	}
}
