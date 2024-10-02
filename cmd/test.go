package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/ui"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, restrict origins
	},
}

var addr string

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(ui.Logo)

		if err := ensureLogsDir(); err != nil {
			return err
		}

		pipeline := ai.NewPipeline(
			cmd.Context(),
			ai.NewConn(),
		).Initialize()

		fmt.Printf("WebSocket server starting on %s\n", addr)

		srv := &http.Server{Addr: addr, Handler: nil}

		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			handleWebSocket(w, r, pipeline)
		})

		// Graceful shutdown setup
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errnie.Error(err)
			}
		}()

		<-ctx.Done()
		stop()
		fmt.Println("Shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		fmt.Println("Server gracefully stopped")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "Address to listen on")
	errnie.Debug("Test command initialized")
}

func ensureLogsDir() error {
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}
	return nil
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, pipeline *ai.Pipeline) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		errnie.Error(err)
		return
	}
	defer conn.Close()

	inChan := make(chan string)
	writeChan := make(chan []byte)
	defer close(writeChan)

	// Reader goroutine
	go func() {
		defer close(inChan)
		for {
			_, message, err := conn.ReadMessage()
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
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				errnie.Error(err)
				break
			}
		}
	}()

	// Processor goroutine
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

			if len(accumulator) > 0 && accumulator[len(accumulator)-1] == '\n' && !isOnlyNewlines(accumulator) {
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

func isOnlyNewlines(s string) bool {
	for _, c := range s {
		if c != '\n' {
			return false
		}
	}
	return true
}
