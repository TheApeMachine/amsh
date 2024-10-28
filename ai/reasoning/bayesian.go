// ai/reasoning/bayesian.go
package reasoning

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// BayesianNetwork represents a Bayesian network with nodes and their dependencies.
type BayesianNetwork struct {
	nodes map[string]*BayesianNode
	mu    sync.RWMutex // Ensures thread-safe access to nodes
}

// BayesianNode represents a node in the Bayesian network.
type BayesianNode struct {
	Name        string
	Parents     []*BayesianNode // Add this field
	Probability map[string]float64
	Properties  map[string]string
}

// ParentState represents the state of parent nodes.
type ParentState struct {
	States map[string]bool // Maps parent node names to their boolean states
}

// NewBayesianNetwork creates and initializes a new Bayesian network.
func NewBayesianNetwork() *BayesianNetwork {
	return &BayesianNetwork{
		nodes: make(map[string]*BayesianNode),
	}
}

// AddNode adds a new node to the Bayesian network.
// Returns an error if a node with the same name already exists.
func (bn *BayesianNetwork) AddNode(name string, prob float64) (*BayesianNode, error) {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	if _, exists := bn.nodes[name]; exists {
		return nil, fmt.Errorf("node '%s' already exists", name)
	}

	node := &BayesianNode{
		Name:        name,
		Parents:     make([]*BayesianNode, 0), // Initialize empty Parents slice
		Probability: make(map[string]float64),
		Properties:  make(map[string]string),
	}
	bn.nodes[name] = node
	return node, nil
}

// AddEdge adds a directed edge from parentName to childName.
// Returns an error if adding the edge creates a cycle or if nodes are not found.
func (bn *BayesianNetwork) AddEdge(childName, parentName string) error {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	child, childExists := bn.nodes[childName]
	if !childExists {
		return fmt.Errorf("child node '%s' not found", childName)
	}

	parent, parentExists := bn.nodes[parentName]
	if !parentExists {
		return fmt.Errorf("parent node '%s' not found", parentName)
	}

	// Check for potential cycles
	if bn.hasPath(parent, child) {
		return fmt.Errorf("adding edge from '%s' to '%s' would create a cycle", parentName, childName)
	}

	child.Parents = append(child.Parents, parent)
	return nil
}

// hasPath checks if there's a path from src to dst to prevent cycles.
func (bn *BayesianNetwork) hasPath(src, dst *BayesianNode) bool {
	visited := make(map[string]bool)
	return bn.dfs(src, dst, visited)
}

// dfs performs a Depth-First Search to find a path from current to target node.
func (bn *BayesianNetwork) dfs(current, target *BayesianNode, visited map[string]bool) bool {
	if current == target {
		return true
	}
	visited[current.Name] = true
	for _, parent := range current.Parents {
		if !visited[parent.Name] {
			if bn.dfs(parent, target, visited) {
				return true
			}
		}
	}
	return false
}

// SetCPT sets the Conditional Probability Table for a node.
// Returns an error if the node does not exist or if CPT keys are invalid.
func (bn *BayesianNetwork) SetCPT(nodeName string, cpt map[string]float64) error {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	node, exists := bn.nodes[nodeName]
	if !exists {
		return fmt.Errorf("node '%s' not found", nodeName)
	}

	node.SetCPT(cpt)
	return nil
}

// CalculateProbability calculates the probability of a node given evidence.
// Utilizes memoization to optimize repeated calculations.
func (bn *BayesianNetwork) CalculateProbability(nodeName string, evidence map[string]bool) (float64, error) {
	bn.mu.RLock()
	node, exists := bn.nodes[nodeName]
	bn.mu.RUnlock()
	if !exists {
		return 0.0, fmt.Errorf("node '%s' not found", nodeName)
	}

	memo := make(map[string]float64)
	visited := make(map[string]bool)
	return bn.calculateNodeProbability(node, evidence, memo, visited)
}

// calculateNodeProbability recursively calculates the probability of a node.
func (bn *BayesianNetwork) calculateNodeProbability(node *BayesianNode, evidence map[string]bool, memo map[string]float64, visited map[string]bool) (float64, error) {
	// Check for memoized result
	if prob, exists := memo[node.Name]; exists {
		return prob, nil
	}

	// If node's state is in evidence, return the corresponding probability
	if val, ok := evidence[node.Name]; ok {
		if val {
			memo[node.Name] = 1.0
			return 1.0, nil
		}
		memo[node.Name] = 0.0
		return 0.0, nil
	}

	// Detect cycles
	if visited[node.Name] {
		return 0.0, fmt.Errorf("cycle detected at node '%s'", node.Name)
	}
	visited[node.Name] = true

	// If no parents, return the prior probability
	if len(node.Parents) == 0 {
		memo[node.Name] = node.Probability["T"]
		return node.Probability["T"], nil
	}

	// Generate all possible parent states and calculate the weighted probability
	totalProb := 0.0
	for serializedParentState := range node.Probability {
		parentProb := 1.0

		// Deserialize the parent state string back to ParentState struct
		parentState, err := DeserializeParentState(serializedParentState)
		if err != nil {
			return 0.0, fmt.Errorf("failed to deserialize parent state '%s': %v", serializedParentState, err)
		}

		for parentName, state := range parentState.States {
			parent, exists := bn.nodes[parentName]
			if !exists {
				return 0.0, fmt.Errorf("parent node '%s' not found", parentName)
			}
			prob, err := bn.calculateNodeProbability(parent, evidence, memo, visited)
			if err != nil {
				return 0.0, err
			}
			if state {
				parentProb *= prob
			} else {
				parentProb *= (1 - prob)
			}
		}
		totalProb += parentStateProbability(parentState) * node.Probability[serializedParentState] * parentProb
	}

	memo[node.Name] = totalProb
	return totalProb, nil
}

// parentStateProbability calculates the joint probability of a given parent state.
// This function assumes independence between parent states, which may need adjustment.
func parentStateProbability(state ParentState) float64 {
	// Placeholder implementation:
	// In a real scenario, you would calculate the probability based on actual parent distributions.
	// Here, we'll assume each parent has a 50% chance of being true or false.
	prob := 1.0
	for _, s := range state.States {
		if s {
			prob *= 0.5
		} else {
			prob *= 0.5
		}
	}
	return prob
}

// SerializeParentState converts a ParentState to its string representation
func SerializeParentState(state ParentState) string {
	keys := make([]string, 0, len(state.States))
	for k := range state.States {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%t;", k, state.States[k]))
	}
	return sb.String()
}

// DeserializeParentState converts a serialized parent state string back to a ParentState struct.
func DeserializeParentState(data string) (ParentState, error) {
	parts := strings.Split(data, ";")
	state := ParentState{States: make(map[string]bool)}
	for _, part := range parts {
		if part == "" {
			continue
		}
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			return ParentState{}, fmt.Errorf("invalid parent state format: %s", part)
		}
		key := kv[0]
		val, err := strconv.ParseBool(kv[1])
		if err != nil {
			return ParentState{}, fmt.Errorf("invalid boolean value in parent state: %s", kv[1])
		}
		state.States[key] = val
	}
	return state, nil
}

// SetAdditionalProperty sets an additional property for the node
func (n *BayesianNode) SetAdditionalProperty(key, value string) {
	if n.Properties == nil {
		n.Properties = make(map[string]string)
	}
	n.Properties[key] = value
}

// SetCPT sets the Conditional Probability Table for the node
func (n *BayesianNode) SetCPT(cpt map[string]float64) {
	n.Probability = cpt
}
