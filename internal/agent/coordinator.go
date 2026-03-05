package agent

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// Coordinator manages file locks and change awareness between agents.
type Coordinator struct {
	config  *config.CoordinatorConfig
	locks   map[string]*FileLock   // file path -> lock
	changes []*FileChange          // recent changes
	mu      sync.RWMutex
}

// Global coordinator instance
var (
	globalCoordinator     *Coordinator
	globalCoordinatorOnce sync.Once
	globalCoordinatorMu   sync.RWMutex
)

// InitGlobalCoordinator initializes the global coordinator.
func InitGlobalCoordinator() {
	globalCoordinatorOnce.Do(func() {
		cfg := config.GetAgent()
		var coordCfg *config.CoordinatorConfig
		if cfg != nil {
			coordCfg = cfg.Coordinator
		}
		globalCoordinatorMu.Lock()
		globalCoordinator = NewCoordinator(coordCfg)
		globalCoordinatorMu.Unlock()
	})
}

// GetGlobalCoordinator returns the global coordinator.
func GetGlobalCoordinator() *Coordinator {
	globalCoordinatorMu.RLock()
	defer globalCoordinatorMu.RUnlock()
	return globalCoordinator
}

// NewCoordinator creates a new coordinator.
func NewCoordinator(cfg *config.CoordinatorConfig) *Coordinator {
	if cfg == nil {
		cfg = &config.CoordinatorConfig{
			Enabled:        false,
			LockTimeoutSec: 300,
			InjectWarnings: true,
		}
	}
	if cfg.LockTimeoutSec == 0 {
		cfg.LockTimeoutSec = 300
	}
	return &Coordinator{
		config:  cfg,
		locks:   make(map[string]*FileLock),
		changes: make([]*FileChange, 0),
	}
}

// IsEnabled returns whether the coordinator is enabled.
func (c *Coordinator) IsEnabled() bool {
	return c.config != nil && c.config.Enabled
}

// UpdateConfig updates the coordinator configuration.
func (c *Coordinator) UpdateConfig(cfg *config.CoordinatorConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = cfg
}

// AcquireLock attempts to acquire a lock on a file.
// Returns (success, existing lock holder session ID).
func (c *Coordinator) AcquireLock(path, sessionID string) (bool, string) {
	if !c.IsEnabled() {
		return true, ""
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Clean expired locks first
	c.cleanExpiredLocksLocked()

	// Check if already locked
	if existing, ok := c.locks[path]; ok {
		if existing.SessionID == sessionID {
			// Extend the lock
			existing.ExpiresAt = time.Now().Add(time.Duration(c.config.LockTimeoutSec) * time.Second)
			return true, ""
		}
		return false, existing.SessionID
	}

	// Acquire new lock
	c.locks[path] = &FileLock{
		Path:      path,
		SessionID: sessionID,
		LockedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(c.config.LockTimeoutSec) * time.Second),
	}
	return true, ""
}

// ReleaseLock releases a lock on a file.
func (c *Coordinator) ReleaseLock(path, sessionID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if lock, ok := c.locks[path]; ok {
		if lock.SessionID == sessionID {
			delete(c.locks, path)
			return true
		}
	}
	return false
}

// ReleaseAllLocks releases all locks held by a session.
func (c *Coordinator) ReleaseAllLocks(sessionID string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for path, lock := range c.locks {
		if lock.SessionID == sessionID {
			delete(c.locks, path)
			count++
		}
	}
	return count
}

// GetLock returns the lock for a file, if any.
func (c *Coordinator) GetLock(path string) *FileLock {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.locks[path]
}

// GetAllLocks returns all current locks.
func (c *Coordinator) GetAllLocks() []*FileLock {
	c.mu.RLock()
	defer c.mu.RUnlock()

	locks := make([]*FileLock, 0, len(c.locks))
	for _, lock := range c.locks {
		locks = append(locks, lock)
	}
	return locks
}

// GetSessionLocks returns all locks held by a session.
func (c *Coordinator) GetSessionLocks(sessionID string) []*FileLock {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var locks []*FileLock
	for _, lock := range c.locks {
		if lock.SessionID == sessionID {
			locks = append(locks, lock)
		}
	}
	return locks
}

// RecordChange records a file change made by an agent.
func (c *Coordinator) RecordChange(path, sessionID, changeType, summary string) {
	if !c.IsEnabled() {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	change := &FileChange{
		Path:       path,
		SessionID:  sessionID,
		ChangeType: changeType,
		Summary:    summary,
		Timestamp:  time.Now(),
	}
	c.changes = append(c.changes, change)

	// Keep only last 500 changes
	if len(c.changes) > 500 {
		c.changes = c.changes[len(c.changes)-500:]
	}
}

// GetRecentChanges returns recent file changes.
func (c *Coordinator) GetRecentChanges(limit int) []*FileChange {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 || limit > len(c.changes) {
		limit = len(c.changes)
	}

	result := make([]*FileChange, limit)
	copy(result, c.changes[len(c.changes)-limit:])
	return result
}

// GetChangesForSession returns changes made by a specific session.
func (c *Coordinator) GetChangesForSession(sessionID string) []*FileChange {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*FileChange
	for _, change := range c.changes {
		if change.SessionID == sessionID {
			result = append(result, change)
		}
	}
	return result
}

// GetChangesForFile returns changes to a specific file.
func (c *Coordinator) GetChangesForFile(path string) []*FileChange {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*FileChange
	for _, change := range c.changes {
		if change.Path == path {
			result = append(result, change)
		}
	}
	return result
}

// GetRecentChangesExcluding returns recent changes excluding a session.
func (c *Coordinator) GetRecentChangesExcluding(sessionID string, limit int) []*FileChange {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*FileChange
	for i := len(c.changes) - 1; i >= 0 && len(result) < limit; i-- {
		if c.changes[i].SessionID != sessionID {
			result = append(result, c.changes[i])
		}
	}
	return result
}

// GenerateContextWarning generates a warning message about locks and changes.
func (c *Coordinator) GenerateContextWarning(sessionID string) string {
	if !c.IsEnabled() || !c.config.InjectWarnings {
		return ""
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	var warnings []string

	// Check for files locked by other sessions
	for path, lock := range c.locks {
		if lock.SessionID != sessionID {
			warnings = append(warnings, fmt.Sprintf("- File '%s' is being modified by another agent session", path))
		}
	}

	// Get recent changes by other sessions
	recentChanges := make([]*FileChange, 0)
	cutoff := time.Now().Add(-10 * time.Minute)
	for _, change := range c.changes {
		if change.SessionID != sessionID && change.Timestamp.After(cutoff) {
			recentChanges = append(recentChanges, change)
		}
	}

	if len(recentChanges) > 0 {
		warnings = append(warnings, "\nRecent changes by other agents:")
		for _, change := range recentChanges {
			warnings = append(warnings, fmt.Sprintf("- [%s] %s: %s", change.ChangeType, change.Path, change.Summary))
		}
	}

	if len(warnings) == 0 {
		return ""
	}

	return fmt.Sprintf("\n[Agent Coordinator Warning]\n%s\n[End Warning]\n", strings.Join(warnings, "\n"))
}

// cleanExpiredLocksLocked removes expired locks. Must be called with lock held.
func (c *Coordinator) cleanExpiredLocksLocked() {
	now := time.Now()
	for path, lock := range c.locks {
		if now.After(lock.ExpiresAt) {
			delete(c.locks, path)
		}
	}
}

// CleanExpiredLocks removes expired locks.
func (c *Coordinator) CleanExpiredLocks() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cleanExpiredLocksLocked()
}
