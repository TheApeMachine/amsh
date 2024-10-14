package service

import (
	"context"
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/mastercomputer"
)

var v = viper.GetViper()
var routerCtx context.Context
var routerCancel context.CancelFunc
var routers []*mastercomputer.Worker

func NewWebhook() fiber.Handler {
	return func(ctx fiber.Ctx) (err error) {
		for _, router := range routers {
			if !router.OK {
				errnie.Warn("skipping router")
				continue
			}

			if router.OK && router.State == mastercomputer.WorkerStateReady {
				if _, err = io.Copy(router, data.New(
					string(ctx.Request().Header.Peek("Origin")),
					"request", "webhook", ctx.Body(),
				)); err != nil {
					errnie.Error(err)
					continue
				}
			}
		}

		if err != nil {
			ctx.SendString(err.Error())
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}

		return ctx.SendStatus(fiber.StatusOK)
	}
}
