package ai

import "sync"

/*
Monitor is responsible for keeping track of various metrics and statistics
for both agents and teams.
*/
type Monitor struct {
	mu sync.RWMutex
}

/*
NewMonitor creates a new monitor instance.
*/
func NewMonitor() *Monitor {
	return &Monitor{}
}

func (monitor *Monitor) Report()
