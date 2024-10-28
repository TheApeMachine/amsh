package planning

import (
	"context"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

// Plan represents a hierarchical execution plan
type Plan struct {
	ID          string
	Name        string
	Description string
	Goals       []Goal
	Timeline    Timeline
	Resources   ResourcePool
	Status      PlanStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Goal struct {
	ID           string
	Name         string
	Description  string
	Priority     int
	Objectives   []Objective
	Dependencies []string // IDs of dependent goals
	Status       GoalStatus
	Deadline     time.Time
}

type Objective struct {
	ID           string
	Name         string
	Description  string
	Tasks        []Task
	Dependencies []string // IDs of dependent objectives
	Status       ObjectiveStatus
	Deadline     time.Time
}

type Task struct {
	ID           string
	Name         string
	Description  string
	AssignedTo   string // Agent ID
	Dependencies []string
	Resources    ResourceRequirement
	Status       TaskStatus
	Duration     time.Duration
	Progress     float64
	StartTime    time.Time // Added field
	EndTime      time.Time // Added field
}

type Timeline struct {
	StartTime     time.Time
	EndTime       time.Time
	Milestones    []Milestone
	Dependencies  []Dependency
	CriticalPath  []string // IDs of critical tasks
	BufferPeriods []BufferPeriod
}

type Milestone struct {
	ID          string
	Name        string
	Description string
	Deadline    time.Time
	Status      MilestoneStatus
	Criteria    []SuccessCriterion
}

// Planner manages long-term planning and execution
type Planner struct {
	plans     map[string]*Plan
	resources *ResourcePool
	mu        sync.RWMutex
}

func NewPlanner() *Planner {
	return &Planner{
		plans:     make(map[string]*Plan),
		resources: NewResourcePool(),
	}
}

// CreatePlan creates a new hierarchical plan
func (p *Planner) CreatePlan(ctx context.Context, req CreatePlanRequest) (*Plan, error) {
	log.Info("Creating plan", "name", req.Name)
	p.mu.Lock()
	defer p.mu.Unlock()

	plan := &Plan{
		ID:          generateID(),
		Name:        req.Name,
		Description: req.Description,
		Goals:       make([]Goal, 0),
		Timeline: Timeline{
			StartTime: time.Now(),
			EndTime:   req.EndTime,
		},
		Resources: *NewResourcePool(), // Create a new ResourcePool instance
		Status:    PlanStatusCreated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Initialize goals and objectives
	for _, goalReq := range req.Goals {
		goal := Goal{
			ID:          generateID(),
			Name:        goalReq.Name,
			Description: goalReq.Description,
			Priority:    goalReq.Priority,
			Status:      GoalStatusPending,
			Deadline:    goalReq.Deadline,
		}

		// Create objectives for each goal
		for _, objReq := range goalReq.Objectives {
			objective := Objective{
				ID:          generateID(),
				Name:        objReq.Name,
				Description: objReq.Description,
				Status:      ObjectiveStatusPending,
				Deadline:    objReq.Deadline,
			}
			goal.Objectives = append(goal.Objectives, objective)
		}

		plan.Goals = append(plan.Goals, goal)
	}

	p.plans[plan.ID] = plan
	return plan, nil
}

// UpdatePlan updates plan progress and adjusts as needed
func (p *Planner) UpdatePlan(ctx context.Context, planID string, updates PlanUpdates) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	plan, exists := p.plans[planID]
	if !exists {
		return ErrPlanNotFound
	}

	// Update progress and status
	for _, update := range updates.TaskUpdates {
		p.updateTaskProgress(plan, update)
	}

	// Check for blocked tasks and resource conflicts
	p.identifyBlockedTasks(plan)
	p.resolveResourceConflicts(plan)

	// Update timeline and critical path
	p.updateTimeline(plan)
	p.recalculateCriticalPath(plan)

	plan.UpdatedAt = time.Now()
	return nil
}

// Helper functions for plan management
func (p *Planner) updateTaskProgress(plan *Plan, update TaskUpdate) {
	for _, goal := range plan.Goals {
		for _, objective := range goal.Objectives {
			for i, task := range objective.Tasks {
				if task.ID == update.TaskID {
					task.Progress = update.Progress
					task.Status = update.Status
					objective.Tasks[i] = task
					break
				}
			}
		}
	}
}

func (p *Planner) identifyBlockedTasks(plan *Plan) {
	for _, goal := range plan.Goals {
		for _, objective := range goal.Objectives {
			for i, task := range objective.Tasks {
				blocked := false
				for _, depID := range task.Dependencies {
					if !p.isTaskComplete(plan, depID) {
						blocked = true
						break
					}
				}
				if blocked {
					task.Status = TaskStatusBlocked
				} else if task.Status == TaskStatusBlocked {
					task.Status = TaskStatusPending
				}
				objective.Tasks[i] = task
			}
		}
	}
}

func (p *Planner) isTaskComplete(plan *Plan, taskID string) bool {
	for _, goal := range plan.Goals {
		for _, objective := range goal.Objectives {
			for _, task := range objective.Tasks {
				if task.ID == taskID {
					return task.Status == TaskStatusComplete
				}
			}
		}
	}
	return false
}

func (p *Planner) resolveResourceConflicts(plan *Plan) {
	// Build a map of resource allocations
	resourceUsage := make(map[string]float64)
	for _, goal := range plan.Goals {
		for _, objective := range goal.Objectives {
			for _, task := range objective.Tasks {
				if task.Status == TaskStatusActive {
					for res, amount := range task.Resources.Custom {
						resourceUsage[res] += amount
					}
				}
			}
		}
	}

	// Check for over-allocations and adjust
	for res, used := range resourceUsage {
		available := plan.Resources.GetAvailable(res)
		if used > available {
			// Find tasks to adjust
			for _, goal := range plan.Goals {
				for _, objective := range goal.Objectives {
					for i, task := range objective.Tasks {
						if task.Status == TaskStatusActive && task.Resources.Custom[res] > 0 {
							// Reduce resource usage or postpone task
							task.Status = TaskStatusBlocked
							objective.Tasks[i] = task
							used -= task.Resources.Custom[res]
							if used <= available {
								break
							}
						}
					}
					if used <= available {
						break
					}
				}
				if used <= available {
					break
				}
			}
		}
	}
}

func (p *Planner) updateTimeline(plan *Plan) {
	// Update task start and end times based on dependencies and progress
	for _, goal := range plan.Goals {
		for _, objective := range goal.Objectives {
			for i, task := range objective.Tasks {
				if task.Status == TaskStatusPending || task.Status == TaskStatusActive {
					earliestStart := plan.Timeline.StartTime
					for _, depID := range task.Dependencies {
						depTask := p.findTaskByID(plan, depID)
						if depTask != nil {
							depEndTime := depTask.StartTime.Add(depTask.Duration)
							if depEndTime.After(earliestStart) {
								earliestStart = depEndTime
							}
						}
					}
					task.StartTime = earliestStart
					task.EndTime = task.StartTime.Add(task.Duration)
					objective.Tasks[i] = task
				}
			}
		}
	}
}

func (p *Planner) recalculateCriticalPath(plan *Plan) {
	// Use Critical Path Method (CPM)
	// Calculate earliest and latest start times
	// Identify tasks with zero slack
	criticalTasks := []string{}
	projectEndTime := plan.Timeline.EndTime

	for _, goal := range plan.Goals {
		for _, objective := range goal.Objectives {
			for _, task := range objective.Tasks {
				// For simplicity, assume tasks with the latest end time are critical
				if task.EndTime.Equal(projectEndTime) || task.EndTime.After(projectEndTime) {
					criticalTasks = append(criticalTasks, task.ID)
				}
			}
		}
	}

	plan.Timeline.CriticalPath = criticalTasks
}

// Helper method to find a task by its ID
func (p *Planner) findTaskByID(plan *Plan, taskID string) *Task {
	for _, goal := range plan.Goals {
		for _, objective := range goal.Objectives {
			for _, task := range objective.Tasks {
				if task.ID == taskID {
					return &task
				}
			}
		}
	}
	return nil
}
