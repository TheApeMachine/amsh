package lsp

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/theapemachine/amsh/logger"
)

type LSPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id,omitempty"`
}

type Client struct {
	stdin  io.WriteCloser
	nextID int
	mu     sync.Mutex
}

// NewClient initializes a new LSP client.
func NewClient(stdin io.WriteCloser) *Client {
	return &Client{
		stdin:  stdin,
		nextID: 1,
	}
}

func (client *Client) sendRequest(method string, params interface{}) error {
	client.mu.Lock()
	id := client.nextID
	client.nextID++
	client.mu.Unlock()

	request := LSPRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return err
	}

	_, err = client.stdin.Write(append(data, '\n'))
	return err
}

// Send initialization request
func (client *Client) SendInitializeRequest(rootURI string) error {
	params := map[string]interface{}{
		"processId": os.Getpid(),
		"rootUri":   rootURI,
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"synchronization": map[string]bool{
					"didSave":   true,
					"didChange": true,
					"willSave":  false,
					"openClose": true,
				},
				"completion": map[string]interface{}{
					"completionItem": map[string]bool{
						"snippetSupport": true,
					},
				},
				"hover": map[string]interface{}{
					"dynamicRegistration": true,
				},
				"definition": map[string]interface{}{
					"dynamicRegistration": true,
				},
			},
			"workspace": map[string]interface{}{
				"workspaceFolders": true,
			},
		},
		"workspaceFolders": []map[string]interface{}{
			{
				"uri":  rootURI,
				"name": filepath.Base(rootURI),
			},
		},
	}

	err := client.sendRequest("initialize", params)
	if err != nil {
		logger.Error("Failed to send initialize request: %v", err)
	} else {
		logger.Debug("Sent initialize request")
	}
	return err
}

func (client *Client) SendDidOpenRequest(uri string, languageID string, version int, content string) error {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        uri,
			"languageId": languageID,
			"version":    version,
			"text":       content,
		},
	}

	err := client.sendRequest("textDocument/didOpen", params)
	if err != nil {
		logger.Error("Failed to send didOpen request: %v", err)
	} else {
		logger.Debug("Sent didOpen request for %s", uri)
	}
	return err
}

func (client *Client) SendDidChangeRequest(uri string, version int, changes []map[string]interface{}) error {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":     uri,
			"version": version,
		},
		"contentChanges": changes,
	}

	err := client.sendRequest("textDocument/didChange", params)
	if err != nil {
		logger.Error("Failed to send didChange request: %v", err)
	} else {
		logger.Debug("Sent didChange request for %s", uri)
	}
	return err
}
