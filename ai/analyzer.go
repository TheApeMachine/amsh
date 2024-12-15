package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/theapemachine/amsh/ai/process/persona"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/utils"
	"github.com/theapemachine/errnie"
)

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// PostMortemAnalysis evaluates and collects outputs as training examples
func (analyzer *Analyzer) PostMortemAnalysis(agent *Agent) {
	for _, sidekick := range agent.Sidekicks {
		errnie.Info("executing sidekick %s from team %s with role %s", sidekick.Name, agent.Name, sidekick.Role)

		var accumulator string

		// Collect sidekick's analysis
		for event := range sidekick.Execute(utils.JoinWith("\n\n",
			utils.JoinWith("\n",
				"<message buffer>",
				agent.Buffer.String(), // Pass full buffer for context
				"</message buffer>",
			),
			utils.JoinWith("\n",
				"<current parameters>",
				string(errnie.SafeMust(func() ([]byte, error) {
					return json.Marshal(agent.params)
				})),
				"</current parameters>",
			),
		)) {
			accumulator += event.Content
		}

		// Extract JSON from code blocks
		codeBlocks := utils.ExtractCodeBlocks(accumulator)
		if jsonBlocks, ok := codeBlocks["json"]; ok {
			for _, jsonBlock := range jsonBlocks {
				var optimizerResponse persona.Optimizer
				if err := json.Unmarshal([]byte(jsonBlock), &optimizerResponse); err != nil {
					errnie.Error(err)
					continue
				}

				// If quality sufficient, save as training example
				if optimizerResponse.AggregatedScore >= 0.8 {
					trainExample := TrainingExample{
						Prompt:   agent.Buffer.GetMessages()[1].Content,
						Response: agent.Buffer.GetMessages()[2].Content,
						Params:   agent.params,
					}

					// Copy assessment details
					for _, assessment := range optimizerResponse.Assessment {
						trainExample.Metadata.Assessment = append(
							trainExample.Metadata.Assessment,
							AssessmentDetail{
								Category: assessment.Category,
								Score:    assessment.Score,
								Reason:   assessment.Reasoning,
							},
						)
					}

					trainExample.Metadata.Score = optimizerResponse.AggregatedScore

					// Copy optimization details
					for _, opt := range optimizerResponse.Optimizations {
						detail := OptimizationDetail{
							Type: opt.Type,
						}
						if opt.Type == "parameter" {
							detail.Parameter = opt.Parameter
							detail.Value = opt.NewValue
						} else {
							detail.Suggestion = opt.Suggestion
						}
						trainExample.Metadata.Optimizations = append(
							trainExample.Metadata.Optimizations,
							detail,
						)
					}

					if err := trainExample.Save(); err != nil {
						errnie.Error(err)
					}
				}

				// Apply parameter optimizations
				for _, opt := range optimizerResponse.Optimizations {
					if opt.Type == "parameter" {
						switch opt.Parameter {
						case "temperature":
							agent.params.Temperature = opt.NewValue
						case "frequency_penalty":
							agent.params.FrequencyPenalty = opt.NewValue
						case "presence_penalty":
							agent.params.PresencePenalty = opt.NewValue
						}
					}
				}
			}
		}
	}
}

// TrainingExample represents a high-quality example for training
type TrainingExample struct {
	Prompt   string                    `json:"prompt"`
	Response string                    `json:"response"`
	Params   provider.GenerationParams `json:"params"`
	Metadata struct {
		Score         float64              `json:"score"`
		Assessment    []AssessmentDetail   `json:"assessment"`
		Optimizations []OptimizationDetail `json:"optimizations"`
	} `json:"metadata"`
}

type AssessmentDetail struct {
	Category string  `json:"category"`
	Score    float64 `json:"score"`
	Reason   string  `json:"reason"`
}

type OptimizationDetail struct {
	Type       string  `json:"type"`
	Suggestion string  `json:"suggestion,omitempty"`
	Parameter  string  `json:"parameter,omitempty"`
	Value      float64 `json:"value,omitempty"`
}

// Save writes the training example to a file in ~/.amsh/train
func (te *TrainingExample) Save() error {
	trainPath := filepath.Join(os.Getenv("HOME"), ".amsh", "train")
	if err := os.MkdirAll(trainPath, 0755); err != nil {
		return fmt.Errorf("failed to create training directory: %v", err)
	}

	filename := fmt.Sprintf("%s-%s.json",
		time.Now().Format("20060102-150405"),
		utils.NewName(),
	)
	filepath := filepath.Join(trainPath, filename)

	data, err := json.MarshalIndent(te, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal training example: %v", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write training example: %v", err)
	}

	return nil
}
