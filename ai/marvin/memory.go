package marvin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/errnie"
)

// Memory represents an agent's memory system
type Memory struct {
	vectorStore *tools.Qdrant
	graphStore  *tools.Neo4j
}

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	Content   string                 `json:"content"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Relations []Relation             `json:"relations,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Relation represents a relationship between memories
type Relation struct {
	Target   string  `json:"target"`
	Type     string  `json:"type"`
	Strength float64 `json:"strength"`
}

// NewMemory creates a new memory system
func NewMemory() *Memory {
	return &Memory{
		vectorStore: tools.NewQdrant("agent_memories", 1536), // OpenAI embedding dimension
		graphStore:  tools.NewNeo4j(),
	}
}

// StoreMemory stores a new memory in both vector and graph stores if applicable
func (m *Memory) StoreMemory(ctx context.Context, entry MemoryEntry) error {
	// Store in vector store for semantic search
	m.vectorStore.Add([]string{entry.Content}, map[string]interface{}{
		"type":      entry.Type,
		"timestamp": entry.Timestamp,
	})

	// If there are relations, store in graph store
	if len(entry.Relations) > 0 {
		// Create node for the memory
		query := fmt.Sprintf(
			`CREATE (m:Memory {content: '%s', type: '%s', timestamp: '%s'})`,
			sanitizeCypher(entry.Content),
			entry.Type,
			entry.Timestamp.Format(time.RFC3339),
		)
		m.graphStore.Write(query)

		// Create relationships
		for _, rel := range entry.Relations {
			query = fmt.Sprintf(
				`MATCH (m1:Memory {content: '%s'}), (m2:Memory {content: '%s'})
				CREATE (m1)-[r:%s {strength: %f}]->(m2)`,
				sanitizeCypher(entry.Content),
				sanitizeCypher(rel.Target),
				rel.Type,
				rel.Strength,
			)
			m.graphStore.Write(query)
		}
	}

	return nil
}

// RetrieveContextualMemories finds relevant memories based on input
func (m *Memory) RetrieveContextualMemories(ctx context.Context, input provider.Message) ([]MemoryEntry, error) {
	var memories []MemoryEntry

	// Search vector store for semantic similarity
	vectorResults, err := m.vectorStore.Query(input.Content)
	if err != nil {
		return nil, err
	}

	// Convert vector results to memory entries
	for _, result := range vectorResults {
		memories = append(memories, MemoryEntry{
			Content:   result["content"].(string),
			Type:      result["metadata"].(map[string]interface{})["type"].(string),
			Timestamp: parseTime(result["metadata"].(map[string]interface{})["timestamp"].(string)),
		})
	}

	// Search graph store for related memories
	query := fmt.Sprintf(
		`MATCH (m:Memory)-[r]-(related)
		WHERE m.content CONTAINS '%s'
		RETURN related, type(r), r.strength
		ORDER BY r.strength DESC
		LIMIT 5`,
		sanitizeCypher(input.Content),
	)

	graphResults, err := m.graphStore.Query(query)
	if err != nil {
		return memories, nil // Return vector results if graph query fails
	}

	// Add graph results to memories
	for _, result := range graphResults {
		node := result["related"].(map[string]interface{})
		memories = append(memories, MemoryEntry{
			Content:   node["content"].(string),
			Type:      node["type"].(string),
			Timestamp: parseTime(node["timestamp"].(string)),
		})
	}

	return memories, nil
}

// AnalyzeForMemories analyzes content for potential memories and relationships
func (m *Memory) AnalyzeForMemories(ctx context.Context, content string) ([]MemoryEntry, error) {
	// Use the agent system to analyze content and extract potential memories
	agent := NewAgent(ctx, "memory_analyzer")
	agent.SetUserPrompt(fmt.Sprintf(`
		Analyze this content for key insights and relationships:
		---
		%s
		---
		Extract key memories and their relationships. Format as JSON array of MemoryEntry objects.
		Each memory should have: content, type (insight/fact/relationship), and optional relations array.
	`, content))

	var memories []MemoryEntry
	var jsonStr strings.Builder

	// Collect agent's response
	for event := range agent.Generate() {
		if event.Type == provider.EventToken {
			jsonStr.WriteString(event.Content)
		}
	}

	// Parse response into memory entries
	err := json.Unmarshal([]byte(jsonStr.String()), &memories)
	if err != nil {
		return nil, errnie.Error(err)
	}

	// Set timestamps
	now := time.Now()
	for i := range memories {
		memories[i].Timestamp = now
	}

	return memories, nil
}

// Helper functions

func sanitizeCypher(input string) string {
	return strings.ReplaceAll(input, "'", "\\'")
}

func parseTime(timestamp string) time.Time {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return time.Now()
	}
	return t
}
