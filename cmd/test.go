package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/integration/boards"
)

var smallTest = []string{
	"How many time do we find the letter r in the word strawberry?",
}

var largeTest = []string{
	"How many time do we find the letter r in the word strawberry?",
	"What is the capital of the moon?",
	"Solve the riddle: In a fruit's sweet name, I'm hidden three, A triple threat within its juicy spree. Find me and you'll discover a secret delight.",
	"Question: Suppose you’re on a game show, and you’re given the choice of three doors: Behind one door is a gold bar; behind the others, rotten vegetables. You pick a door, say No.1, and the host asks you “Do you want to pick door No.2 instead?” Is it to your advantage to switch your choice?",
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		work := boards.NewService()
		prompt, err := work.SearchWorkitems(
			context.Background(),
			"push",
		)

		if err != nil {
			errnie.Error(fmt.Errorf("failed to search workitems: %v", err))
			return err
		}

		fmt.Println(prompt)

		// Start a log file for the full conversation, including prompts.
		logFileName := fmt.Sprintf("logs/run_%s.md", time.Now().Format("2006-01-02_15-04-05"))
		logFile, err := os.Create(logFileName)
		if err != nil {
			errnie.Error(fmt.Errorf("failed to create log file: %v", err))
			return err
		}

		// Start a log file that only shows the repsonses.
		responsesLogFileName := fmt.Sprintf("logs/run_%s_responses.md", time.Now().Format("2006-01-02_15-04-05"))
		responsesLogFile, err := os.Create(responsesLogFileName)
		if err != nil {
			errnie.Error(fmt.Errorf("failed to create responses log file: %v", err))
			return err
		}

		pipeline := ai.NewPipeline(context.Background())
		pipeline.Initialize()

		currentAgent := ""

		for chunk := range pipeline.Generate(prompt, 3) {
			fmt.Print(chunk.Response)

			if chunk.Agent.ID != currentAgent {
				if _, err := logFile.WriteString(strings.Join([]string{
					strings.Join(chunk.Agent.Prompt.System, ""),
					strings.Join(chunk.Agent.Prompt.User, ""),
				}, "")); err != nil {
					errnie.Error(fmt.Errorf("failed to write to log file: %v", err))
				}

				if _, err := responsesLogFile.WriteString(fmt.Sprintf("**AGENT: %s (%s)**\n\n", chunk.Agent.ID, chunk.Agent.Type)); err != nil {
					errnie.Error(fmt.Errorf("failed to write to responses log file: %v", err))
				}

				currentAgent = chunk.Agent.ID
			}

			if _, err := logFile.WriteString(chunk.Response); err != nil {
				errnie.Error(fmt.Errorf("failed to write to log file: %v", err))
			}

			if _, err := responsesLogFile.WriteString(chunk.Response); err != nil {
				errnie.Error(fmt.Errorf("failed to write to responses log file: %v", err))
			}
		}

		if err := logFile.Close(); err != nil {
			errnie.Error(fmt.Errorf("failed to close log file: %v", err))
		}

		if err := responsesLogFile.Close(); err != nil {
			errnie.Error(fmt.Errorf("failed to close responses log file: %v", err))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
