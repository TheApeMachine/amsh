package service

import (
	"context"
	"net/http"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/integration/comms"
	"github.com/theapemachine/amsh/mastercomputer"
	"github.com/theapemachine/amsh/twoface"
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
	// Initialize the messaging queue
	queue := twoface.NewQueue()

	// Initialize the worker manager
	builder := mastercomputer.NewBuilder()

	for _, agent := range []mastercomputer.WorkerType{
		mastercomputer.WorkerTypeManager,
		mastercomputer.WorkerTypeReasoner,
		mastercomputer.WorkerTypeVerifier,
		mastercomputer.WorkerTypeCommunicator,
		mastercomputer.WorkerTypeResearcher,
		mastercomputer.WorkerTypeExecutor,
	} {
		worker := builder.NewWorker(agent)
		worker.Start()
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
		queue:       queue,
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

	// WebSocket route using adaptor
	https.app.Get("/ws", adaptor.HTTPHandler(handler(websocketHandler)))

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

func handler(f http.HandlerFunc) http.Handler {
	return http.HandlerFunc(f)
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
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

			errnie.Info("Received message: " + string(msg))

			// Process the message (you can add your custom logic here)
			response := []byte("Acknowledged: " + string(msg))

			err = wsutil.WriteServerMessage(conn, op, response)
			if err != nil {
				errnie.Error(err)
				break
			}
		}
	}()
}
