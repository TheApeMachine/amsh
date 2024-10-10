package mastercomputer

type WorkerState int

const (
	WaitingForPrompt WorkerState = iota
	CommandInProgress
	ReadyToExit
)
