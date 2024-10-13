package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/theapemachine/amsh/errnie"
)

type LSPResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	Result  map[string]interface{} `json:"result"`
	ID      int                    `json:"id"`
}

type LSPError struct {
	Text string `json:"text"`
}

type Server struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	err    error
}

func NewServer() *Server {
	return &Server{}
}

// Start the gopls server without interfering with the TUI
func (server *Server) Start() (io.WriteCloser, chan LSPResponse, chan LSPError) {
	errnie.Debug("Starting gopls server")
	server.cmd = exec.Command("gopls", "-rpc.trace", "serve")

	// Create pipes for gopls stdin, stdout, and stderr
	if server.stdin, server.err = server.cmd.StdinPipe(); server.err != nil {
		errnie.Error(server.err)
		return nil, nil, nil
	}

	if server.stdout, server.err = server.cmd.StdoutPipe(); server.err != nil {
		errnie.Error(server.err)
		return nil, nil, nil
	}

	if server.stderr, server.err = server.cmd.StderrPipe(); server.err != nil {
		errnie.Error(server.err)
		return nil, nil, nil
	}

	// Start the gopls process
	if server.err = server.cmd.Start(); server.err != nil {
		errnie.Error(server.err)
		return nil, nil, nil
	}

	// Handle stdout and stderr in separate goroutines
	return server.stdin, server.handleStdout(), server.handleStderr()
}

func (server *Server) handleStdout() chan LSPResponse {
	out := make(chan LSPResponse)

	go func() {
		defer close(out)
		reader := bufio.NewReader(server.stdout)

		for {
			// Read the Content-Length header
			line, err := reader.ReadString('\n')
			if err != nil {
				errnie.Error(err)
				break
			}

			// Check for the Content-Length header
			if strings.HasPrefix(line, "Content-Length:") {
				var contentLength int
				_, err := fmt.Sscanf(line, "Content-Length: %d", &contentLength)
				if err != nil {
					errnie.Error(err)
					break
				}

				// Read the next line (should be empty)
				line, err = reader.ReadString('\n')
				if err != nil || line != "\r\n" {
					errnie.Error(err)
					break
				}

				// Read the JSON response body
				jsonBytes := make([]byte, contentLength)
				_, err = io.ReadFull(reader, jsonBytes)
				if err != nil {
					errnie.Error(err)
					break
				}

				response := LSPResponse{}
				json.Unmarshal(jsonBytes, &response)

				// Send the JSON response through the channel
				out <- response
			}
		}
	}()

	return out
}

// handleStderr reads and processes gopls stderr
func (server *Server) handleStderr() chan LSPError {
	out := make(chan LSPError)

	go func() {
		defer close(out)
		scanner := bufio.NewScanner(server.stderr)

		for scanner.Scan() {
			// Handle errors or warnings from gopls
			errnie.Info("gopls stderr: %s", scanner.Text())
			out <- LSPError{Text: scanner.Text()}
		}

		if server.err = scanner.Err(); server.err != nil {
			errnie.Error(server.err)
		}
	}()

	return out
}
