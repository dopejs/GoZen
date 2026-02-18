// Package agent provides agent infrastructure for GoZen.
// [BETA] This feature is experimental and disabled by default.
package agent

import (
	"sync"
	"time"
)

// FileLock represents a lock on a file by an agent session.
type FileLock struct {
	Path      string    `json:"path"`
	SessionID string    `json:"session_id"`
	LockedAt  time.Time `json:"locked_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// FileChange represents a file change made by an agent.
type FileChange struct {
	Path       string    `json:"path"`
	SessionID  string    `json:"session_id"`
	ChangeType string    `json:"change_type"` // "create", "modify", "delete"
	Summary    string    `json:"summary"`
	Timestamp  time.Time `json:"timestamp"`
}

// ObservedSession represents a monitored agent session.
type ObservedSession struct {
	ID           string    `json:"id"`
	Profile      string    `json:"profile"`
	Client       string    `json:"client"`
	ProjectPath  string    `json:"project_path"`
	StartTime    time.Time `json:"start_time"`
	LastActivity time.Time `json:"last_activity"`

	// Metrics
	TotalTokens  int     `json:"total_tokens"`
	TotalCost    float64 `json:"total_cost"`
	RequestCount int     `json:"request_count"`
	ErrorCount   int     `json:"error_count"`

	// State
	CurrentTask string `json:"current_task"`
	Status      string `json:"status"` // "active", "idle", "stuck", "paused", "killed"

	// Stuck detection
	LastErrors []string `json:"last_errors,omitempty"`
	RetryCount int      `json:"retry_count"`

	// Internal
	mu sync.RWMutex `json:"-"`
}

// AgentTask represents a task in the queue.
type AgentTask struct {
	ID          string      `json:"id"`
	Description string      `json:"description"`
	Priority    int         `json:"priority"`
	Status      string      `json:"status"` // "pending", "running", "completed", "failed", "cancelled"
	AssignedTo  string      `json:"assigned_to,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	StartedAt   time.Time   `json:"started_at,omitempty"`
	CompletedAt time.Time   `json:"completed_at,omitempty"`
	Result      *TaskResult `json:"result,omitempty"`
	RetryCount  int         `json:"retry_count"`
	MaxRetries  int         `json:"max_retries"`
}

// TaskResult holds the result of a completed task.
type TaskResult struct {
	Success bool    `json:"success"`
	Output  string  `json:"output"`
	Error   string  `json:"error,omitempty"`
	Tokens  int     `json:"tokens"`
	Cost    float64 `json:"cost"`
}

// RuntimeTask represents an autonomous agent task.
type RuntimeTask struct {
	ID          string       `json:"id"`
	Description string       `json:"description"`
	Status      string       `json:"status"` // "planning", "executing", "validating", "completed", "failed", "cancelled"
	Plan        *TaskPlan    `json:"plan,omitempty"`
	Turns       []*AgentTurn `json:"turns,omitempty"`
	Result      *TaskResult  `json:"result,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	StartedAt   time.Time    `json:"started_at,omitempty"`
	CompletedAt time.Time    `json:"completed_at,omitempty"`
	TotalTokens int          `json:"total_tokens"`
	TotalCost   float64      `json:"total_cost"`
}

// TaskPlan holds the execution plan for a task.
type TaskPlan struct {
	Steps       []string `json:"steps"`
	CurrentStep int      `json:"current_step"`
}

// AgentTurn represents a single turn in an agent conversation.
type AgentTurn struct {
	Model     string    `json:"model"`
	Phase     string    `json:"phase"` // "planning", "execution", "validation"
	Request   []byte    `json:"request,omitempty"`
	Response  []byte    `json:"response,omitempty"`
	Tokens    int       `json:"tokens"`
	Cost      float64   `json:"cost"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// SensitiveOperation represents a detected sensitive operation.
type SensitiveOperation struct {
	SessionID   string    `json:"session_id"`
	Type        string    `json:"type"` // "file_delete", "config_modify", "db_operation", "network_call"
	Description string    `json:"description"`
	Path        string    `json:"path,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Blocked     bool      `json:"blocked"`
}

// SessionStatus constants
const (
	SessionStatusActive = "active"
	SessionStatusIdle   = "idle"
	SessionStatusStuck  = "stuck"
	SessionStatusPaused = "paused"
	SessionStatusKilled = "killed"
)

// TaskStatus constants
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
	TaskStatusCancelled = "cancelled"
)

// RuntimeStatus constants
const (
	RuntimeStatusPlanning   = "planning"
	RuntimeStatusExecuting  = "executing"
	RuntimeStatusValidating = "validating"
	RuntimeStatusCompleted  = "completed"
	RuntimeStatusFailed     = "failed"
	RuntimeStatusCancelled  = "cancelled"
)

// ChangeType constants
const (
	ChangeTypeCreate = "create"
	ChangeTypeModify = "modify"
	ChangeTypeDelete = "delete"
)

// SensitiveOpType constants
const (
	SensitiveOpFileDelete   = "file_delete"
	SensitiveOpConfigModify = "config_modify"
	SensitiveOpDBOperation  = "db_operation"
	SensitiveOpNetworkCall  = "network_call"
)
