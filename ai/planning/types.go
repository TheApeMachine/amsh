package planning

import (
	"fmt"
	"sync"
	"time"
)

type PlanStatus string
type GoalStatus string
type ObjectiveStatus string
type TaskStatus string
type MilestoneStatus string

const (
	PlanStatusCreated   PlanStatus = "created"
	PlanStatusActive    PlanStatus = "active"
	PlanStatusBlocked   PlanStatus = "blocked"
	PlanStatusComplete  PlanStatus = "complete"
	PlanStatusCancelled PlanStatus = "cancelled"

	GoalStatusPending   GoalStatus = "pending"
	GoalStatusActive    GoalStatus = "active"
	GoalStatusBlocked   GoalStatus = "blocked"
	GoalStatusComplete  GoalStatus = "complete"
	GoalStatusCancelled GoalStatus = "cancelled"

	ObjectiveStatusPending   ObjectiveStatus = "pending"
	ObjectiveStatusActive    ObjectiveStatus = "active"
	ObjectiveStatusBlocked   ObjectiveStatus = "blocked"
	ObjectiveStatusComplete  ObjectiveStatus = "complete"
	ObjectiveStatusCancelled ObjectiveStatus = "cancelled"

	TaskStatusPending   TaskStatus = "pending"
	TaskStatusActive    TaskStatus = "active"
	TaskStatusBlocked   TaskStatus = "blocked"
	TaskStatusComplete  TaskStatus = "complete"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type CreatePlanRequest struct {
	Name        string
	Description string
	EndTime     time.Time
	Goals       []CreateGoalRequest
}

type CreateGoalRequest struct {
	Name        string
	Description string
	Priority    int
	Deadline    time.Time
	Objectives  []CreateObjectiveRequest
}

type CreateObjectiveRequest struct {
	Name        string
	Description string
	Deadline    time.Time
}

type PlanUpdates struct {
	TaskUpdates []TaskUpdate
}

type TaskUpdate struct {
	TaskID   string
	Progress float64
	Status   TaskStatus
}

type ResourceRequirement struct {
	CPU     float64
	Memory  float64
	Storage float64
	Custom  map[string]float64
}

// Resource management types
type ResourcePool struct {
	Available map[string]float64
	Reserved  map[string]float64
	mu        sync.RWMutex
}

func NewResourcePool() *ResourcePool {
	return &ResourcePool{
		Available: make(map[string]float64),
		Reserved:  make(map[string]float64),
	}
}

// Add these methods
func (rp *ResourcePool) GetAvailable(resource string) float64 {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.Available[resource]
}

func (rp *ResourcePool) SetAvailable(resource string, amount float64) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.Available[resource] = amount
}

// Timeline related types
type Dependency struct {
	FromID string
	ToID   string
	Type   DependencyType
	Lag    time.Duration
}

type DependencyType string

const (
	FinishToStart  DependencyType = "finish_to_start"
	StartToStart   DependencyType = "start_to_start"
	FinishToFinish DependencyType = "finish_to_finish"
	StartToFinish  DependencyType = "start_to_finish"
)

type BufferPeriod struct {
	StartTime time.Time
	Duration  time.Duration
	Purpose   string
}

// Success criteria types
type SuccessCriterion struct {
	ID          string
	Description string
	Metric      string
	Target      float64
	Current     float64
}

// Error types
var (
	ErrPlanNotFound = fmt.Errorf("plan not found")
)

// Helper functions
func generateID() string {
	return fmt.Sprintf("id_%d", time.Now().UnixNano())
}
