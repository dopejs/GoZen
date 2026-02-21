package bot

import (
	"sync"
	"time"
)

// Session represents a user's chat session.
type Session struct {
	UserID       string    `json:"user_id"`
	Platform     Platform  `json:"platform"`
	ChatID       string    `json:"chat_id"`
	BoundProcess string    `json:"bound_process,omitempty"`
	LastActive   time.Time `json:"last_active"`
}

// SessionManager manages user sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session // key: platform:user_id
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func sessionKey(platform Platform, userID string) string {
	return string(platform) + ":" + userID
}

// Get returns a session for a user.
func (m *SessionManager) Get(platform Platform, userID string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionKey(platform, userID)]
}

// GetOrCreate returns an existing session or creates a new one.
func (m *SessionManager) GetOrCreate(platform Platform, userID, chatID string) *Session {
	key := sessionKey(platform, userID)

	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[key]; ok {
		s.LastActive = time.Now()
		s.ChatID = chatID // Update chat ID in case it changed
		return s
	}

	s := &Session{
		UserID:     userID,
		Platform:   platform,
		ChatID:     chatID,
		LastActive: time.Now(),
	}
	m.sessions[key] = s
	return s
}

// Bind binds a process to a user's session.
func (m *SessionManager) Bind(platform Platform, userID, processName string) {
	key := sessionKey(platform, userID)

	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[key]; ok {
		s.BoundProcess = processName
		s.LastActive = time.Now()
	} else {
		m.sessions[key] = &Session{
			UserID:       userID,
			Platform:     platform,
			BoundProcess: processName,
			LastActive:   time.Now(),
		}
	}
}

// Unbind removes the process binding from a user's session.
func (m *SessionManager) Unbind(platform Platform, userID string) {
	key := sessionKey(platform, userID)

	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[key]; ok {
		s.BoundProcess = ""
	}
}

// GetBoundProcess returns the bound process for a user.
func (m *SessionManager) GetBoundProcess(platform Platform, userID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if s, ok := m.sessions[sessionKey(platform, userID)]; ok {
		return s.BoundProcess
	}
	return ""
}

// Cleanup removes stale sessions.
func (m *SessionManager) Cleanup(maxAge time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	count := 0
	for key, s := range m.sessions {
		if now.Sub(s.LastActive) > maxAge {
			delete(m.sessions, key)
			count++
		}
	}
	return count
}

// PendingApproval tracks a pending approval request.
type PendingApproval struct {
	ID        string       `json:"id"`
	ProcessID string       `json:"process_id"`
	ReplyTo   ReplyContext `json:"reply_to"`
	MessageID string       `json:"message_id"` // The message with buttons
	CreatedAt time.Time    `json:"created_at"`
	Timeout   time.Time    `json:"timeout,omitempty"`
}

// ApprovalManager manages pending approval requests.
type ApprovalManager struct {
	mu       sync.RWMutex
	pending  map[string]*PendingApproval // approval ID -> pending
	byMsgID  map[string]string           // message ID -> approval ID
}

// NewApprovalManager creates a new approval manager.
func NewApprovalManager() *ApprovalManager {
	return &ApprovalManager{
		pending: make(map[string]*PendingApproval),
		byMsgID: make(map[string]string),
	}
}

// Add adds a pending approval.
func (m *ApprovalManager) Add(approval *PendingApproval) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pending[approval.ID] = approval
	if approval.MessageID != "" {
		m.byMsgID[approval.MessageID] = approval.ID
	}
}

// Get returns a pending approval by ID.
func (m *ApprovalManager) Get(id string) *PendingApproval {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pending[id]
}

// GetByMessageID returns a pending approval by message ID.
func (m *ApprovalManager) GetByMessageID(msgID string) *PendingApproval {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if id, ok := m.byMsgID[msgID]; ok {
		return m.pending[id]
	}
	return nil
}

// Remove removes a pending approval.
func (m *ApprovalManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if approval, ok := m.pending[id]; ok {
		if approval.MessageID != "" {
			delete(m.byMsgID, approval.MessageID)
		}
		delete(m.pending, id)
	}
}

// Cleanup removes expired approvals.
func (m *ApprovalManager) Cleanup() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var expired []string
	for id, approval := range m.pending {
		if !approval.Timeout.IsZero() && now.After(approval.Timeout) {
			if approval.MessageID != "" {
				delete(m.byMsgID, approval.MessageID)
			}
			delete(m.pending, id)
			expired = append(expired, id)
		}
	}
	return expired
}
