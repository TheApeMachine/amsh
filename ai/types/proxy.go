package types

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/theapemachine/amsh/errnie"
)

// Proxy handles routing memory operations to the appropriate store
type Proxy struct {
	vector    *Qdrant
	graph     *Neo4j
	store     string
	operation string
	data      string
}

// NewProxy creates a new memory proxy with the given parameters
func NewProxy(parameters map[string]any) *Proxy {
	store, _ := parameters["store"].(string)
	operation, _ := parameters["operation"].(string)
	data, _ := parameters["data"].(string)

	proxy := &Proxy{
		store:     store,
		operation: operation,
		data:      data,
	}

	// Initialize the appropriate store based on the parameters
	if store == "vector" {
		proxy.vector = NewQdrant("hive", 1536)
	} else if store == "graph" {
		proxy.graph = NewNeo4j()
	}

	return proxy
}

// Start executes the memory operation and returns the result
func (proxy *Proxy) Start() string {
	var result string
	var err error

	switch proxy.store {
	case "vector":
		result, err = proxy.handleVectorOperation()
	case "graph":
		result, err = proxy.handleGraphOperation()
	default:
		result = "Invalid store type specified"
	}

	if err != nil {
		errnie.Error(err)
		result = "Error: " + err.Error()
	}

	return result
}

func (proxy *Proxy) handleVectorOperation() (string, error) {
	switch proxy.operation {
	case "add":
		docs := []string{proxy.data}
		ids, err := proxy.vector.Add(docs)
		if err != nil {
			return "", err
		}
		return "Successfully added document with ID: " + strings.Join(ids, ", "), nil

	case "search":
		results, err := proxy.vector.Query(proxy.data)
		if err != nil {
			return "", err
		}
		jsonResult, err := json.Marshal(results)
		if err != nil {
			return "", err
		}
		return string(jsonResult), nil

	case "update":
		// For update, we'll treat it as an add operation since vectors are immutable
		docs := []string{proxy.data}
		ids, err := proxy.vector.Add(docs)
		if err != nil {
			return "", err
		}
		return "Successfully updated document with ID: " + strings.Join(ids, ", "), nil
	}

	return "", nil
}

func (proxy *Proxy) handleGraphOperation() (string, error) {
	switch proxy.operation {
	case "add", "update":
		result, err := proxy.graph.Write(proxy.data)
		if err != nil {
			return "", err
		}
		summary, err := result.Consume(context.Background())
		if err != nil {
			return "", err
		}
		return "Successfully executed graph operation. Nodes affected: " +
			strconv.Itoa(summary.Counters().NodesCreated()), nil

	case "search":
		results, err := proxy.graph.Query(proxy.data)
		if err != nil {
			return "", err
		}
		jsonResult, err := json.Marshal(results)
		if err != nil {
			return "", err
		}
		return string(jsonResult), nil
	}

	return "", nil
}
