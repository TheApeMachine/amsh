package ai

import "time"

// Add these type definitions at the top of the file
type ProcessStatus string
type StepStatus string

const (
	ProcessStatusPending  ProcessStatus = "pending"
	ProcessStatusRunning  ProcessStatus = "running"
	ProcessStatusComplete ProcessStatus = "complete"
	ProcessStatusFailed   ProcessStatus = "failed"

	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
)

type Process struct {
	Name        string
	Steps       []ProcessStep
	CurrentStep int
	Status      ProcessStatus
}

type ProcessStep struct {
	Name      string
	Agents    []string
	Input     string
	Output    string
	Status    StepStatus
	StartTime time.Time
	EndTime   time.Time
}
