package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3/client"
	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/berrt"
	"github.com/theapemachine/amsh/mastercomputer"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE:  runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	// Example user prompt input
	userPrompt := `
	You have been hired by a Dutch company called Fan Factory, which delivers products and services
	in relation to employee wellbeing. Your job is to research the company, their domain, and present
	a comprehensive plan to scale their business.
	`

	seq := mastercomputer.NewSequencer(cmd.Context(), userPrompt)
	events := mastercomputer.NewEvents()

	var wg sync.WaitGroup
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		for event := range events.Stream() {
			// For now, just log the events. This can be replaced with frontend integration.
			if event.WorkerID != "" {
				berrt.Info(event.WorkerID, event.Message)
			} else {
				berrt.Info(event.Type, event.Message)
			}
		}
	}(&wg)

	seq.Start()

	wg.Wait()

	// Delete Qdrant collections
	if err := deleteQdrantCollections(); err != nil {
		log.Printf("Error deleting Qdrant collections: %v", err)
	}

	return nil
}

func deleteQdrantCollections() error {
	// List collections
	resp, err := client.Get("http://localhost:6333/collections")
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	var listResp struct {
		Result struct {
			Collections []struct {
				Name string `json:"name"`
			} `json:"collections"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp.Body(), &listResp); err != nil {
		return fmt.Errorf("failed to unmarshal collection list: %w", err)
	}

	// Delete collections
	for _, collection := range listResp.Result.Collections {
		if strings.Contains(collection.Name, "marvin") || strings.Contains(collection.Name, "hive") {
			log.Printf("Skipping deletion of protected collection: %s", collection.Name)
			continue
		}

		resp, err := client.Delete(fmt.Sprintf("http://localhost:6333/collections/%s", collection.Name))
		if err != nil {
			log.Printf("Failed to delete collection %s: %v", collection.Name, err)
			continue
		}

		if resp.StatusCode() != http.StatusOK {
			log.Printf("Failed to delete collection %s: status code %d", collection.Name, resp.StatusCode())
		} else {
			log.Printf("Successfully deleted collection: %s", collection.Name)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(testCmd)
	os.Setenv("LOGFILE", "test.log")
	os.Setenv("QDRANT_URL", "http://localhost:6333")
	os.Setenv("NEO4J_URL", "neo4j://localhost:7474")
}
