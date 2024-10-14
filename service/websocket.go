package service

import (
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/sockpuppet"
)

func NewWebSocketHandler() func(c *sockpuppet.WebsocketConn) {
	return func(c *sockpuppet.WebsocketConn) {
		defer c.Close()

		errnie.Debug("WebSocket connection established")

		inChan := make(chan string)
		writeChan := make(chan []byte)
		defer close(writeChan)
		defer close(inChan)

		// Reader goroutine
		go func() {
			errnie.Info("Reader goroutine started")

			var (
				msg []byte
				err error
			)

			for {
				if _, msg, err = c.ReadMessage(); err != nil {
					errnie.Error(err)
					break
				}

				errnie.Info("Received message: " + string(msg))
				inChan <- string(msg)
			}
		}()

		// Writer goroutine
		go func() {
			errnie.Info("Writer goroutine started")

			for msg := range writeChan {
				if err := c.WriteMessage(sockpuppet.TextMessage, msg); err != nil {
					errnie.Error(err)
					break
				}
			}
		}()
	}
}
