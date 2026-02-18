package agent

import (
	"crypto/rand"
	"encoding/hex"
	"sort"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// TaskQueue manages agent tasks.
type TaskQueue struct {
	config *config.TaskQueueConfig
	tasks  map[string]*AgentTask
	mu     sync.RWMutex
}

// Global task queue instance
var (
	globalTaskQueue     *TaskQueue
	globalTaskQueueOnce sync.Once
	globalTaskQueueMu   sync.RWMutex
)

// InitGlobalTaskQueue initializes the global task queue.
func InitGlobalTaskQueue() {
	globalTaskQueueOnce.Do(func() {
		cfg := config.GetAgent()
		var tqCfg *config.TaskQueueConfig
		if cfg != nil {
			tqCfg = cfg.TaskQueue
		}
		globalTaskQueueMu.Lock()
		globalTaskQueue = NewTaskQueue(tqCfg)
		globalTaskQueueMu.Unlock()
	})
}

// GetGlobalTaskQueue returns the global task queue.
func GetGlobalTaskQueue() *TaskQueue {
	globalTaskQueueMu.RLock()
	defer globalTaskQueueMu.RUnlock()
	return globalTaskQueue
}

// NewTaskQueue creates a new task queue.
func NewTaskQueue(cfg *config.TaskQueueConfig) *TaskQueue {
	if cfg == nil {
		cfg = &config.TaskQueueConfig{
			Enabled:    false,
			MaxRetries: 3,
			Workers:    1,
		}
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.Workers == 0 {
		cfg.Workers = 1
	}
	return &TaskQueue{
		config: cfg,
		tasks:  make(map[string]*AgentTask),
	}
}

// IsEnabled returns whether the task queue is enabled.
func (q *TaskQueue) IsEnabled() bool {
	return q.config != nil && q.config.Enabled
}

// UpdateConfig updates the task queue configuration.
func (q *TaskQueue) UpdateConfig(cfg *config.TaskQueueConfig) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.config = cfg
}

// AddTask adds a new task to the queue.
func (q *TaskQueue) AddTask(description string, priority int) *AgentTask {
	q.mu.Lock()
	defer q.mu.Unlock()

	task := &AgentTask{
		ID:          generateTaskID(),
		Description: description,
		Priority:    priority,
		Status:      TaskStatusPending,
		CreatedAt:   time.Now(),
		MaxRetries:  q.config.MaxRetries,
	}
	q.tasks[task.ID] = task
	return task
}

// GetTask returns a task by ID.
func (q *TaskQueue) GetTask(id string) *AgentTask {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.tasks[id]
}

// GetAllTasks returns all tasks.
func (q *TaskQueue) GetAllTasks() []*AgentTask {
	q.mu.RLock()
	defer q.mu.RUnlock()

	tasks := make([]*AgentTask, 0, len(q.tasks))
	for _, t := range q.tasks {
		tasks = append(tasks, t)
	}

	// Sort by priority (higher first), then by creation time
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Priority != tasks[j].Priority {
			return tasks[i].Priority > tasks[j].Priority
		}
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})

	return tasks
}

// GetPendingTasks returns all pending tasks sorted by priority.
func (q *TaskQueue) GetPendingTasks() []*AgentTask {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var tasks []*AgentTask
	for _, t := range q.tasks {
		if t.Status == TaskStatusPending {
			tasks = append(tasks, t)
		}
	}

	// Sort by priority (higher first), then by creation time
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Priority != tasks[j].Priority {
			return tasks[i].Priority > tasks[j].Priority
		}
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})

	return tasks
}

// GetRunningTasks returns all running tasks.
func (q *TaskQueue) GetRunningTasks() []*AgentTask {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var tasks []*AgentTask
	for _, t := range q.tasks {
		if t.Status == TaskStatusRunning {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

// GetNextTask returns the next pending task and marks it as running.
func (q *TaskQueue) GetNextTask(sessionID string) *AgentTask {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Find highest priority pending task
	var best *AgentTask
	for _, t := range q.tasks {
		if t.Status == TaskStatusPending {
			if best == nil || t.Priority > best.Priority ||
				(t.Priority == best.Priority && t.CreatedAt.Before(best.CreatedAt)) {
				best = t
			}
		}
	}

	if best != nil {
		best.Status = TaskStatusRunning
		best.AssignedTo = sessionID
		best.StartedAt = time.Now()
	}

	return best
}

// CompleteTask marks a task as completed.
func (q *TaskQueue) CompleteTask(id string, result *TaskResult) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, ok := q.tasks[id]
	if !ok {
		return false
	}

	task.Status = TaskStatusCompleted
	task.CompletedAt = time.Now()
	task.Result = result
	return true
}

// FailTask marks a task as failed.
func (q *TaskQueue) FailTask(id string, result *TaskResult) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, ok := q.tasks[id]
	if !ok {
		return false
	}

	task.RetryCount++
	if task.RetryCount < task.MaxRetries {
		// Reset to pending for retry
		task.Status = TaskStatusPending
		task.AssignedTo = ""
	} else {
		task.Status = TaskStatusFailed
		task.CompletedAt = time.Now()
		task.Result = result
	}
	return true
}

// CancelTask cancels a task.
func (q *TaskQueue) CancelTask(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, ok := q.tasks[id]
	if !ok {
		return false
	}

	if task.Status == TaskStatusPending || task.Status == TaskStatusRunning {
		task.Status = TaskStatusCancelled
		task.CompletedAt = time.Now()
		return true
	}
	return false
}

// RetryTask resets a failed task to pending.
func (q *TaskQueue) RetryTask(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, ok := q.tasks[id]
	if !ok {
		return false
	}

	if task.Status == TaskStatusFailed {
		task.Status = TaskStatusPending
		task.AssignedTo = ""
		task.RetryCount = 0
		task.Result = nil
		return true
	}
	return false
}

// DeleteTask removes a task from the queue.
func (q *TaskQueue) DeleteTask(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, ok := q.tasks[id]; ok {
		delete(q.tasks, id)
		return true
	}
	return false
}

// GetStats returns queue statistics.
func (q *TaskQueue) GetStats() map[string]int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	stats := map[string]int{
		"total":     len(q.tasks),
		"pending":   0,
		"running":   0,
		"completed": 0,
		"failed":    0,
		"cancelled": 0,
	}

	for _, t := range q.tasks {
		switch t.Status {
		case TaskStatusPending:
			stats["pending"]++
		case TaskStatusRunning:
			stats["running"]++
		case TaskStatusCompleted:
			stats["completed"]++
		case TaskStatusFailed:
			stats["failed"]++
		case TaskStatusCancelled:
			stats["cancelled"]++
		}
	}

	return stats
}

// CleanOldTasks removes completed/failed/cancelled tasks older than the given duration.
func (q *TaskQueue) CleanOldTasks(maxAge time.Duration) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	count := 0

	for id, t := range q.tasks {
		if (t.Status == TaskStatusCompleted || t.Status == TaskStatusFailed || t.Status == TaskStatusCancelled) &&
			t.CompletedAt.Before(cutoff) {
			delete(q.tasks, id)
			count++
		}
	}

	return count
}

// generateTaskID generates a unique task ID.
func generateTaskID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "task-" + hex.EncodeToString(b)
}
