package types

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// ChatInteractionType represents the type of chat interaction
type ChatInteractionType int

const (
	HumanInitiated ChatInteractionType = iota
	AIInitiated
)

// Message represents a single chat message
type Message struct {
	Content   string
	From      string // "human" or "ai"
	Timestamp time.Time
	Context   *MessageContext
	Priority  int // Higher priority messages are notifications/questions
}

// MessageContext holds relevant context for a message
type MessageContext struct {
	FilePath  string
	LineStart int
	LineEnd   int
	Selection string
	ChangeID  string
	LockID    string // Reference to file lock if relevant
}

// Chat manages the chat interface and AI interaction
type Chat struct {
	mode          ChatInteractionType
	messages      []Message
	activeSession bool
	aiContainer   *AIContainer
	changeTracker *ChangeTracker
	notifications []Message // Pending AI notifications/questions
	fileLocks     *LockManager
	mu            sync.RWMutex // Protects message and notification access
}

// GetMessages returns all chat messages
func (c *Chat) GetMessages() []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]Message{}, c.messages...)
}

// LockManager handles file locking between human and AI
type LockManager struct {
	locks map[string]*FileLock
	mu    sync.RWMutex
}

// FileLock represents a lock on a file
type FileLock struct {
	ID        string
	FilePath  string
	Owner     string // "human" or "ai"
	Timestamp time.Time
	ChangeID  string // Associated change if any
}

// AIContainer manages communication with the AI docker container
type AIContainer struct {
	containerID string
	workingDir  string
	changes     map[string]*FileChange
	activeFiles map[string]bool // Files AI is currently working on
	mu          sync.RWMutex
}

// NewChat creates a new chat instance
func NewChat() *Chat {
	return &Chat{
		messages:      make([]Message, 0),
		notifications: make([]Message, 0),
		changeTracker: &ChangeTracker{
			humanChanges: make(map[string]*FileChange),
			aiChanges:    make(map[string]*FileChange),
			conflicts:    make(map[string]*Conflict),
		},
		aiContainer: &AIContainer{
			changes:     make(map[string]*FileChange),
			activeFiles: make(map[string]bool),
		},
		fileLocks: &LockManager{
			locks: make(map[string]*FileLock),
		},
	}
}

// AcquireLock attempts to acquire a lock on a file
func (lm *LockManager) AcquireLock(filePath, owner string, changeID string) (*FileLock, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Check if file is already locked
	if existing, exists := lm.locks[filePath]; exists {
		return nil, &LockError{
			FilePath: filePath,
			Owner:    existing.Owner,
			Message:  "file is locked by " + existing.Owner,
		}
	}

	// Create new lock
	lock := &FileLock{
		ID:        GenerateID(),
		FilePath:  filePath,
		Owner:     owner,
		Timestamp: time.Now(),
		ChangeID:  changeID,
	}
	lm.locks[filePath] = lock
	return lock, nil
}

// ReleaseLock releases a lock on a file
func (lm *LockManager) ReleaseLock(filePath, owner string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lock, exists := lm.locks[filePath]; exists {
		if lock.Owner != owner {
			return &LockError{
				FilePath: filePath,
				Owner:    lock.Owner,
				Message:  "cannot release lock owned by " + lock.Owner,
			}
		}
		delete(lm.locks, filePath)
		return nil
	}
	return nil
}

// AddNotification adds an AI notification/question
func (c *Chat) AddNotification(content string, context *MessageContext) {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg := Message{
		Content:   content,
		From:      "ai",
		Timestamp: time.Now(),
		Context:   context,
		Priority:  1, // Higher priority for notifications
	}
	c.notifications = append(c.notifications, msg)
}

// GetPendingNotifications returns pending AI notifications
func (c *Chat) GetPendingNotifications() []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return append([]Message{}, c.notifications...)
}

// ClearNotification removes a notification after it's handled
func (c *Chat) ClearNotification(timestamp time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, msg := range c.notifications {
		if msg.Timestamp == timestamp {
			c.notifications = append(c.notifications[:i], c.notifications[i+1:]...)
			return
		}
	}
}

// NotifyHumanChange notifies the AI of a human change
func (ai *AIContainer) NotifyHumanChange(change *FileChange) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	// Update AI's view of the file
	// This would trigger the AI to review the change and potentially adapt its work
	if ai.activeFiles[change.FilePath] {
		// AI is working on this file - it should review the change
		// This would trigger the AI to evaluate if it needs to:
		// 1. Modify its current work
		// 2. Ask a question
		// 3. Suggest an alternative
		// Implementation would depend on AI system
	}
}

// LockError represents a file locking error
type LockError struct {
	FilePath string
	Owner    string
	Message  string
}

func (e *LockError) Error() string {
	return e.Message
}

// ChangeTracker tracks and manages changes from both human and AI
type ChangeTracker struct {
	humanChanges map[string]*FileChange
	aiChanges    map[string]*FileChange
	conflicts    map[string]*Conflict
}

// FileChange represents a change to a file
type FileChange struct {
	FilePath   string
	ChangeType string // "add", "modify", "delete"
	Content    string
	Timestamp  time.Time
	Author     string // "human" or "ai"
	ChangeID   string
	ConflictID string // If this change is part of a conflict
}

// Conflict represents a conflict between human and AI changes
type Conflict struct {
	ID          string
	FilePath    string
	HumanChange *FileChange
	AIChange    *FileChange
	Resolution  string // "human", "ai", "merged"
	Resolved    bool
}

// GenerateID creates a unique identifier for changes and conflicts
func GenerateID() string {
	return uuid.New().String()
}

// StartChat initiates a chat session
func (c *Chat) StartChat(mode ChatInteractionType) {
	c.mode = mode
	c.activeSession = true
}

// AddMessage adds a message to the chat
func (c *Chat) AddMessage(content, from string, context *MessageContext) {
	msg := Message{
		Content:   content,
		From:      from,
		Timestamp: time.Now(),
		Context:   context,
	}
	c.messages = append(c.messages, msg)
}

// TrackChange records a change from either human or AI
func (ct *ChangeTracker) TrackChange(change *FileChange) {
	if change.Author == "human" {
		ct.humanChanges[change.ChangeID] = change
	} else {
		ct.aiChanges[change.ChangeID] = change
	}

	// Check for conflicts
	ct.detectConflicts(change)
}

// detectConflicts checks if a new change conflicts with existing changes
func (ct *ChangeTracker) detectConflicts(newChange *FileChange) {
	var existingChanges map[string]*FileChange
	if newChange.Author == "human" {
		existingChanges = ct.aiChanges
	} else {
		existingChanges = ct.humanChanges
	}

	for _, existing := range existingChanges {
		if existing.FilePath == newChange.FilePath {
			// Create conflict
			conflict := &Conflict{
				ID:       GenerateID(), // Implement this
				FilePath: newChange.FilePath,
				Resolved: false,
			}
			if newChange.Author == "human" {
				conflict.HumanChange = newChange
				conflict.AIChange = existing
			} else {
				conflict.HumanChange = existing
				conflict.AIChange = newChange
			}
			ct.conflicts[conflict.ID] = conflict

			// Update changes to reference conflict
			newChange.ConflictID = conflict.ID
			existing.ConflictID = conflict.ID
		}
	}
}

// GetPendingConflicts returns unresolved conflicts
func (ct *ChangeTracker) GetPendingConflicts() []*Conflict {
	var pending []*Conflict
	for _, conflict := range ct.conflicts {
		if !conflict.Resolved {
			pending = append(pending, conflict)
		}
	}
	return pending
}

// ResolveConflict resolves a conflict with the given resolution
func (ct *ChangeTracker) ResolveConflict(conflictID, resolution string) {
	if conflict, exists := ct.conflicts[conflictID]; exists {
		conflict.Resolution = resolution
		conflict.Resolved = true
	}
}
