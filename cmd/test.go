// cmd/test.go
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/ai/mastercomputer"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/qpool"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI system integration test",
	Long:  `Run a test that demonstrates the integration between agents, communication, and VM components.`,
	RunE:  runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	os.Setenv("NOCONSOLE", "false")
	ctx := context.Background()

	// Initialize communication system
	comm := mastercomputer.NewAgentCommunication(ctx)
	defer comm.Close()

	// Create test agents
	researcher := mastercomputer.NewAgent("researcher", "researcher")
	analyst := mastercomputer.NewAgent("analyst", "analyst")

	// Register agents
	comm.RegisterAgent(researcher)
	comm.RegisterAgent(analyst)

	// Start a discussion
	discussionID, err := comm.StartDiscussion([]string{"researcher", "analyst"})
	if err != nil {
		return fmt.Errorf("failed to start discussion: %v", err)
	}

	// Subscribe both agents and monitoring channel to the discussion
	researcherMsgs, _ := comm.JoinDiscussion(discussionID)
	analystMsgs, _ := comm.JoinDiscussion(discussionID)
	monitorMsgs, _ := comm.JoinDiscussion(discussionID)

	// Start message handlers for agents
	go handleAgentMessages("researcher", researcherMsgs, comm)
	go handleAgentMessages("analyst", analystMsgs, comm)

	// Create and send initial message
	err = comm.SendMessage(discussionID, mastercomputer.Message{
		From:    "researcher",
		Content: "How many times do we find the letter r in the word strawberry?",
		Type:    "question",
	})
	if err != nil {
		return fmt.Errorf("failed to send initial message: %v", err)
	}

	// Create a test program for the VM
	testProgram := []mastercomputer.Instruction{
		{Op: mastercomputer.OpStore, Operands: []interface{}{"query", "How many times do we find the letter r in the word strawberry?"}},
		{Op: mastercomputer.OpStore, Operands: []interface{}{"result", 4}},
		{Op: mastercomputer.OpSend, Operands: []interface{}{"analyst", map[string]interface{}{
			"type": "analysis_request",
			"data": "Found 4 occurrences of 'r' in 'strawberry'",
		}}},
	}

	// Create and run VM
	vm := mastercomputer.NewAgentVM(comm.GetPool(), comm)

	errnie.Info("\nStarting VM execution...")
	if err = vm.Execute(ctx, testProgram); err != nil {
		return errnie.Error(fmt.Errorf("VM execution failed: %v", err))
	}
	errnie.Info("VM execution completed")

	// Monitor messages
	errnie.Info("Monitoring agent communication:")
	timeout := time.After(10 * time.Second)

	for {
		select {
		case msg := <-monitorMsgs:
			if msg.Error != nil {
				errnie.Error(fmt.Errorf("error: %v", msg.Error))
				continue
			}
			message := msg.Value.(mastercomputer.Message)
			errnie.Info(fmt.Sprintf("\nMessage:\n  From: %s\n  Type: %s\n  Content: %v\n",
				message.From,
				message.Type,
				message.Content,
			))
		case <-timeout:
			errnie.Warn("Test timed out")
			return nil
		}
	}
}

func handleAgentMessages(agentID string, msgs <-chan qpool.QuantumValue, comm *mastercomputer.AgentCommunication) {
	for msg := range msgs {
		if msg.Error != nil {
			errnie.Error(msg.Error)
			continue
		}

		message := msg.Value.(mastercomputer.Message)

		// Process message based on agent role
		switch agentID {
		case "analyst":
			if message.Type == "analysis_request" {
				// Analyst responds to analysis requests
				reply := mastercomputer.Message{
					From:    "analyst",
					To:      message.From,
					Content: fmt.Sprintf("Analysis confirmed: %v", message.Content),
					Type:    "analysis_response",
				}

				// Send response through discussion
				comm.SendMessage(message.ID, reply)
			}
		case "researcher":
			if message.Type == "analysis_response" {
				// Researcher processes analysis responses
				errnie.Info(fmt.Sprintf("\nResearcher received analysis: %v\n", message.Content))
			}
		}
	}
}

// func deleteQdrantCollections() error {
// 	// List collections
// 	resp, err := client.Get("http://localhost:6333/collections")
// 	if err != nil {
// 		return fmt.Errorf("failed to list collections: %w", err)
// 	}

// 	var listResp struct {
// 		Result struct {
// 			Collections []struct {
// 				Name string `json:"name"`
// 			} `json:"collections"`
// 		} `json:"result"`
// 	}

// 	if err := json.Unmarshal(resp.Body(), &listResp); err != nil {
// 		return fmt.Errorf("failed to unmarshal collection list: %w", err)
// 	}

// 	// Delete collections
// 	for _, collection := range listResp.Result.Collections {
// 		if strings.Contains(collection.Name, "marvin") || strings.Contains(collection.Name, "hive") {
// 			log.Printf("Skipping deletion of protected collection: %s", collection.Name)
// 			continue
// 		}

// 		resp, err := client.Delete(fmt.Sprintf("http://localhost:6333/collections/%s", collection.Name))
// 		if err != nil {
// 			log.Printf("Failed to delete collection %s: %v", collection.Name, err)
// 			continue
// 		}

// 		if resp.StatusCode() != http.StatusOK {
// 			log.Printf("Failed to delete collection %s: status code %d", collection.Name, resp.StatusCode())
// 		} else {
// 			log.Printf("Successfully deleted collection: %s", collection.Name)
// 		}
// 	}

// 	return nil
// }

func init() {
	rootCmd.AddCommand(testCmd)
	os.Setenv("LOGFILE", "test.log")
	os.Setenv("QDRANT_URL", "http://localhost:6333")
	os.Setenv("NEO4J_URL", "neo4j://localhost:7474")
}
