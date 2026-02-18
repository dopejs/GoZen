package agent

import (
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// Observatory monitors all agent sessions.
type Observatory struct {
	config   *config.ObservatoryConfig
	sessions map[string]*ObservedSession
	mu       sync.RWMutex
}

// Global observatory instance
var (
	globalObservatory     *Observatory
	globalObservatoryOnce sync.Once
	globalObservatoryMu   sync.RWMutex
)

// InitGlobalObservatory initializes the global observatory.
func InitGlobalObservatory() {
	globalObservatoryOnce.Do(func() {
		cfg := config.GetAgent()
		var obsCfg *config.ObservatoryConfig
		if cfg != nil {
			obsCfg = cfg.Observatory
		}
		globalObservatoryMu.Lock()
		globalObservatory = NewObservatory(obsCfg)
		globalObservatoryMu.Unlock()
	})
}

// GetGlobalObservatory returns the global observatory.
func GetGlobalObservatory() *Observatory {
	globalObservatoryMu.RLock()
	defer globalObservatoryMu.RUnlock()
	return globalObservatory
}

// NewObservatory creates a new observatory.
func NewObservatory(cfg *config.ObservatoryConfig) *Observatory {
	if cfg == nil {
		cfg = &config.ObservatoryConfig{
			Enabled:        false,
			StuckThreshold: 5,
			IdleTimeoutMin: 30,
		}
	}
	if cfg.StuckThreshold == 0 {
		cfg.StuckThreshold = 5
	}
	if cfg.IdleTimeoutMin == 0 {
		cfg.IdleTimeoutMin = 30
	}
	return &Observatory{
		config:   cfg,
		sessions: make(map[string]*ObservedSession),
	}
}

// IsEnabled returns whether the observatory is enabled.
func (o *Observatory) IsEnabled() bool {
	return o.config != nil && o.config.Enabled
}

// UpdateConfig updates the observatory configuration.
func (o *Observatory) UpdateConfig(cfg *config.ObservatoryConfig) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.config = cfg
}

// RegisterSession registers a new session for monitoring.
func (o *Observatory) RegisterSession(id, profile, client, projectPath string) *ObservedSession {
	o.mu.Lock()
	defer o.mu.Unlock()

	session := &ObservedSession{
		ID:           id,
		Profile:      profile,
		Client:       client,
		ProjectPath:  projectPath,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		Status:       SessionStatusActive,
		LastErrors:   make([]string, 0),
	}
	o.sessions[id] = session
	return session
}

// GetSession returns a session by ID.
func (o *Observatory) GetSession(id string) *ObservedSession {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.sessions[id]
}

// GetAllSessions returns all sessions.
func (o *Observatory) GetAllSessions() []*ObservedSession {
	o.mu.RLock()
	defer o.mu.RUnlock()

	sessions := make([]*ObservedSession, 0, len(o.sessions))
	for _, s := range o.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// GetActiveSessions returns only active sessions.
func (o *Observatory) GetActiveSessions() []*ObservedSession {
	o.mu.RLock()
	defer o.mu.RUnlock()

	sessions := make([]*ObservedSession, 0)
	for _, s := range o.sessions {
		if s.Status == SessionStatusActive || s.Status == SessionStatusIdle {
			sessions = append(sessions, s)
		}
	}
	return sessions
}

// RemoveSession removes a session from monitoring.
func (o *Observatory) RemoveSession(id string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.sessions, id)
}

// RecordRequest records a request for a session.
func (o *Observatory) RecordRequest(sessionID string, tokens int, cost float64, err error) {
	o.mu.RLock()
	session, ok := o.sessions[sessionID]
	o.mu.RUnlock()

	if !ok {
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	session.LastActivity = time.Now()
	session.RequestCount++
	session.TotalTokens += tokens
	session.TotalCost += cost

	if err != nil {
		session.ErrorCount++
		session.LastErrors = append(session.LastErrors, err.Error())
		if len(session.LastErrors) > 10 {
			session.LastErrors = session.LastErrors[1:]
		}
		session.RetryCount++

		// Check for stuck condition
		if session.RetryCount >= o.config.StuckThreshold {
			session.Status = SessionStatusStuck
		}
	} else {
		session.RetryCount = 0
		session.Status = SessionStatusActive
	}
}

// SetSessionTask updates the current task for a session.
func (o *Observatory) SetSessionTask(sessionID, task string) {
	o.mu.RLock()
	session, ok := o.sessions[sessionID]
	o.mu.RUnlock()

	if !ok {
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()
	session.CurrentTask = task
	session.LastActivity = time.Now()
}

// KillSession marks a session as killed.
func (o *Observatory) KillSession(sessionID string) bool {
	o.mu.RLock()
	session, ok := o.sessions[sessionID]
	o.mu.RUnlock()

	if !ok {
		return false
	}

	session.mu.Lock()
	defer session.mu.Unlock()
	session.Status = SessionStatusKilled
	return true
}

// PauseSession marks a session as paused.
func (o *Observatory) PauseSession(sessionID string) bool {
	o.mu.RLock()
	session, ok := o.sessions[sessionID]
	o.mu.RUnlock()

	if !ok {
		return false
	}

	session.mu.Lock()
	defer session.mu.Unlock()
	session.Status = SessionStatusPaused
	return true
}

// ResumeSession resumes a paused session.
func (o *Observatory) ResumeSession(sessionID string) bool {
	o.mu.RLock()
	session, ok := o.sessions[sessionID]
	o.mu.RUnlock()

	if !ok {
		return false
	}

	session.mu.Lock()
	defer session.mu.Unlock()
	if session.Status == SessionStatusPaused {
		session.Status = SessionStatusActive
		return true
	}
	return false
}

// IsSessionKilled checks if a session has been killed.
func (o *Observatory) IsSessionKilled(sessionID string) bool {
	o.mu.RLock()
	session, ok := o.sessions[sessionID]
	o.mu.RUnlock()

	if !ok {
		return false
	}

	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.Status == SessionStatusKilled
}

// IsSessionPaused checks if a session has been paused.
func (o *Observatory) IsSessionPaused(sessionID string) bool {
	o.mu.RLock()
	session, ok := o.sessions[sessionID]
	o.mu.RUnlock()

	if !ok {
		return false
	}

	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.Status == SessionStatusPaused
}

// CheckIdleSessions marks sessions as idle if inactive.
func (o *Observatory) CheckIdleSessions() {
	o.mu.RLock()
	defer o.mu.RUnlock()

	idleTimeout := time.Duration(o.config.IdleTimeoutMin) * time.Minute
	now := time.Now()

	for _, session := range o.sessions {
		session.mu.Lock()
		if session.Status == SessionStatusActive && now.Sub(session.LastActivity) > idleTimeout {
			session.Status = SessionStatusIdle
		}
		session.mu.Unlock()
	}
}

// GetStats returns aggregate statistics.
func (o *Observatory) GetStats() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var totalTokens int
	var totalCost float64
	var activeCount, idleCount, stuckCount, pausedCount int

	for _, s := range o.sessions {
		s.mu.RLock()
		totalTokens += s.TotalTokens
		totalCost += s.TotalCost
		switch s.Status {
		case SessionStatusActive:
			activeCount++
		case SessionStatusIdle:
			idleCount++
		case SessionStatusStuck:
			stuckCount++
		case SessionStatusPaused:
			pausedCount++
		}
		s.mu.RUnlock()
	}

	return map[string]interface{}{
		"total_sessions": len(o.sessions),
		"active":         activeCount,
		"idle":           idleCount,
		"stuck":          stuckCount,
		"paused":         pausedCount,
		"total_tokens":   totalTokens,
		"total_cost":     totalCost,
	}
}
