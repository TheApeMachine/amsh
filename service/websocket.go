package service

import (
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/theapemachine/amsh/errnie"
)

func NewWebSocketHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade HTTP connection to WebSocket
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
}
