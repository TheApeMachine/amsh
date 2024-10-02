package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/sockpuppet"
	"github.com/theapemachine/amsh/utils"
)

func NewWebSocketHandler(pipeline *ai.Pipeline) func(c *sockpuppet.WebsocketConn) {
	return func(c *sockpuppet.WebsocketConn) {
		defer c.Close()

		inChan := make(chan string)
		writeChan := make(chan []byte)
		defer close(writeChan)

		// Reader goroutine
		go func() {
			defer close(inChan)
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					errnie.Error(err)
					break
				}
				errnie.Debug("Received message: " + string(message))
				inChan <- string(message)
			}
		}()

		// Writer goroutine
		go func() {
			for msg := range writeChan {
				if err := c.WriteMessage(sockpuppet.TextMessage, msg); err != nil {
					errnie.Error(err)
					break
				}
			}
		}()

		// Processor
		for promptIn := range inChan {
			parts := strings.Split(promptIn, "<:>")
			if len(parts) != 2 {
				errnie.Error(fmt.Errorf("invalid message format"))
				continue
			}

			iterations, err := strconv.Atoi(parts[0])
			if err != nil {
				errnie.Error(fmt.Errorf("invalid iteration count: %v", err))
				continue
			}

			prompt := parts[1]

			logFileName := fmt.Sprintf("logs/run_%s.md", time.Now().Format("2006-01-02_15-04-05"))
			logFile, err := os.Create(logFileName)
			if err != nil {
				errnie.Error(fmt.Errorf("failed to create log file: %v", err))
				continue
			}

			accumulator := ""
			for chunk := range pipeline.Generate(prompt, iterations) {
				stripped := stripansi.Strip(chunk.Response)
				accumulator += stripped
				fmt.Print(chunk.Response)

				if len(accumulator) > 0 && accumulator[len(accumulator)-1] == '\n' && !utils.IsOnlyNewlines(accumulator) {
					if _, err := logFile.WriteString(accumulator); err != nil {
						errnie.Error(fmt.Errorf("failed to write to log file: %v", err))
					}

					chunk.Response = accumulator
					buf, err := json.Marshal(chunk)
					if err != nil {
						errnie.Error(fmt.Errorf("failed to marshal chunk: %v", err))
						continue
					}

					writeChan <- buf
					accumulator = ""
				}
			}

			if err := logFile.Close(); err != nil {
				errnie.Error(fmt.Errorf("failed to close log file: %v", err))
			}
			errnie.Debug("Run complete")
		}
	}
}
