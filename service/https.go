package service

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/theapemachine/amsh/mastercomputer"
	"github.com/theapemachine/amsh/sockpuppet"
	"github.com/theapemachine/amsh/twoface"
)

/*
HTTPS wraps the Fiber app and sets up the middleware. It also contains the mapping
to internal service endpoints.
*/
type HTTPS struct {
	app   *fiber.App
	queue *twoface.Queue
}

/*
NewHTTPS creates a new HTTPS service, configures the mapping to internal service endpoints
from the config file, and sets up fiber (v3) to serve TLS requests.
*/
func NewHTTPS() *HTTPS {
	manager := twoface.NewWorkerManager()
	builder := mastercomputer.NewBuilder(context.Background(), manager)

	reasoner := builder.NewWorker("reasoner")
	reasoner.Start()

	verifier := builder.NewWorker("verifier")
	verifier.Start()

	return &HTTPS{
		app: fiber.New(fiber.Config{
			CaseSensitive: true,
			StrictRouting: true,
			ServerHeader:  "Fiber",
			AppName:       "AMSH Service",
			JSONEncoder:   json.Marshal,
			JSONDecoder:   json.Unmarshal,
		}),
		queue: twoface.NewQueue(),
	}
}

/*
Up adds the middleware and starts the HTTPS service.
*/
func (https *HTTPS) Up() error {
	https.app.Use(
		cors.New(cors.Config{
			AllowOrigins: []string{"*"},
			AllowHeaders: []string{"*"},
			AllowMethods: []string{"*"},
		}),
		favicon.New(),
	)

	https.app.Use("/ws", func(c fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if sockpuppet.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	https.app.Get("/ws", sockpuppet.NewWebsocket(NewWebSocketHandler()))
	https.app.Post("/webhook/trengo", https.NewWebhook("trengo", "message"))
	https.app.Use("/", static.New("./frontend"))

	return https.app.Listen(":8567", fiber.ListenConfig{EnablePrefork: false})
}

/*
Shutdown gracefully shuts down the HTTPS service.
*/
func (https *HTTPS) Shutdown() error {
	return https.app.Shutdown()
}
