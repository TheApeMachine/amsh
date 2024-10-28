package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/theapemachine/amsh/ai/system"
	"github.com/theapemachine/amsh/berrt"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/integration/comms"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
)

/*
HTTPS wraps the Fiber app and sets up the middleware. It also contains the mapping
to internal service endpoints.
*/
type HTTPS struct {
	app         *fiber.App
	queue       *twoface.Queue
	slackEvents *comms.Events
	arch        *system.Architecture
}

/*
NewHTTPS creates a new HTTPS service, configures the mapping to internal service endpoints
from the config file, and sets up fiber (v3) to serve TLS requests.
*/
func NewHTTPS() *HTTPS {
	// Initialize the architecture
	arch := system.NewArchitecture("amsh")
	if arch == nil {
		errnie.Error(fmt.Errorf("architecture not initialized"))
		return nil
	}

	return &HTTPS{
		app: fiber.New(fiber.Config{
			CaseSensitive: true,
			StrictRouting: true,
			ServerHeader:  "Fiber",
			AppName:       "AMSH Service",
			JSONEncoder:   json.Marshal,
			JSONDecoder:   json.Unmarshal,
		}),
		slackEvents: comms.NewEvents(arch),
		arch:        arch, // Add the architecture
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

	// WebSocket route using adaptor
	https.app.Get("/ws", adaptor.HTTPHandler(handler(https.websocketHandler)))

	https.app.Post("/webhook/trengo", https.NewWebhook("trengo", "managing"))
	https.app.Post("/webhook/github", https.NewWebhook("github", "managing"))
	https.app.Post("/events/slack", https.slackEvents.Run)
	https.app.Use("/", static.New("./frontend"))

	// Start the main HTTP server
	return https.app.Listen(":8567", fiber.ListenConfig{EnablePrefork: false})
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

func handler(f http.HandlerFunc) http.Handler {
	return http.HandlerFunc(f)
}

func (https *HTTPS) websocketHandler(w http.ResponseWriter, r *http.Request) {
	// Add architecture check
	if https.arch == nil {
		errnie.Error(fmt.Errorf("architecture not initialized"))
		return
	}

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		errnie.Error(err)
		return
	}

	errnie.Debug("WebSocket connection established")

	go func() {
		defer conn.Close()

		for {
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				errnie.Error(err)
				break
			}

			if strings.Contains(string(msg), "ping") {
				continue
			}

			// Create a new message
			message := data.New(utils.NewName(), "task", "managing", msg)
			message.Poke("chain", "websocket")

			// Start discussion process
			resultChan := https.arch.ProcessManager.Execute(string(msg))
			if resultChan == nil {
				errnie.Error(fmt.Errorf("process result channel not found"))
				continue
			}

			// Send response back through websocket
			for result := range resultChan {
				berrt.Error("Stream", wsutil.WriteServerMessage(conn, op, []byte(result.Content)))
				break
			}

			https.queue.PubCh <- *message
		}
	}()
}
