package service

import (
	"context"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/theapemachine/amsh/integration/comms"
	"github.com/theapemachine/amsh/mastercomputer"
	"github.com/theapemachine/amsh/sockpuppet"
	"github.com/theapemachine/amsh/twoface"
	"sync"
)

/*
HTTPS wraps the Fiber app and sets up the middleware. It also contains the mapping
to internal service endpoints.
*/
type HTTPS struct {
	app         *fiber.App
	queue       *twoface.Queue
	slackEvents *comms.Events
}

/*
NewHTTPS creates a new HTTPS service, configures the mapping to internal service endpoints
from the config file, and sets up fiber (v3) to serve TLS requests.
*/
func NewHTTPS() *HTTPS {
	builder := mastercomputer.NewBuilder()

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
		queue:       twoface.NewQueue(),
		slackEvents: comms.NewEvents(),
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

	var wg sync.WaitGroup
	wg.Add(2)

	// Start the main HTTP server
	go func() {
		defer wg.Done()
		if err := https.app.Listen(":8567", fiber.ListenConfig{EnablePrefork: false}); err != nil {
			panic(err)
		}
	}()

	// Start the Slack events server
	go func() {
		defer wg.Done()
		if err := https.slackEvents.Run(context.Background()); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
	return nil
}

/*
Shutdown gracefully shuts down the HTTPS service.
*/
func (https *HTTPS) Shutdown() error {
	// Shutdown the main HTTP server
	if err := https.app.Shutdown(); err != nil {
		return err
	}

	// Add any necessary cleanup for the Slack events server
	// For now, we don't have a specific shutdown method for it

	return nil
}
