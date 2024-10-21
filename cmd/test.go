package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3/client"
	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/mastercomputer"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE:  runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	// Initialize the messaging queue
	queue := twoface.NewQueue()

	// Initialize the worker manager
	builder := mastercomputer.NewBuilder()

	for _, agent := range []mastercomputer.WorkerType{
		mastercomputer.WorkerTypeManager,
		mastercomputer.WorkerTypeReasoner,
		mastercomputer.WorkerTypeVerifier,
		mastercomputer.WorkerTypeCommunicator,
		mastercomputer.WorkerTypeResearcher,
		mastercomputer.WorkerTypeExecutor,
	} {
		worker := builder.NewWorker(agent)
		worker.Start()
	}

	message := data.New(utils.NewName(), "task", "managing", []byte("How many times do we find the letter r in the word strawberry?"))
	message.Poke("chain", "test")

	queue.PubCh <- *message

	log.Println("Waiting for workers to finish...")
	builder.Wait()

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
}
