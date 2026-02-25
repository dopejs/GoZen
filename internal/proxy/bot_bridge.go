package proxy

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/bot"
)

// BotBridge connects proxy sessions to the bot gateway.
// It registers proxy sessions as "virtual" processes so the bot can
// report their status.
type BotBridge struct {
	mu       sync.RWMutex
	client   *bot.Client
	sessions map[string]*bridgeSession // sessionID -> session info
	enabled  bool
}

type bridgeSession struct {
	SessionID   string
	ClientType  string
	LastMessage string
	MessageRole string
	Status      string
	WaitingFor  string
	TokensUsed  int
	TurnCount   int
	LastUpdate  time.Time
}

var (
	globalBridge     *BotBridge
	globalBridgeOnce sync.Once
)

// InitBotBridge initializes the global bot bridge.
func InitBotBridge(gatewayPath string) error {
	var initErr error
	globalBridgeOnce.Do(func() {
		if gatewayPath == "" {
			gatewayPath = filepath.Join("/tmp", "zen-gateway.sock")
		}
		globalBridge = &BotBridge{
			sessions: make(map[string]*bridgeSession),
			enabled:  true,
		}
	})
	return initErr
}

// GetBotBridge returns the global bot bridge.
func GetBotBridge() *BotBridge {
	return globalBridge
}

// UpdateSession updates a session's status in the bridge.
// This is called by the proxy when processing requests.
func (b *BotBridge) UpdateSession(sessionID, clientType string, usage *SessionUsage, lastMessage, messageRole, waitingFor string) {
	if !b.enabled || sessionID == "" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	sess, exists := b.sessions[sessionID]
	if !exists {
		sess = &bridgeSession{
			SessionID:  sessionID,
			ClientType: clientType,
		}
		b.sessions[sessionID] = sess
	}

	sess.ClientType = clientType
	sess.LastUpdate = time.Now()

	if lastMessage != "" {
		sess.LastMessage = lastMessage
		sess.MessageRole = messageRole
	}

	if waitingFor != "" {
		sess.WaitingFor = waitingFor
	}

	if usage != nil {
		sess.TokensUsed = usage.InputTokens + usage.OutputTokens
		sess.TurnCount = usage.TurnCount
	}

	// Determine status based on state
	if waitingFor != "" {
		sess.Status = "waiting"
	} else if usage != nil && usage.TurnCount > 0 {
		sess.Status = "active"
	} else {
		sess.Status = "idle"
	}
}

// MarkSessionBusy marks a session as busy (processing a request).
func (b *BotBridge) MarkSessionBusy(sessionID, clientType string) {
	if !b.enabled || sessionID == "" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	sess, exists := b.sessions[sessionID]
	if !exists {
		sess = &bridgeSession{
			SessionID:  sessionID,
			ClientType: clientType,
		}
		b.sessions[sessionID] = sess
	}

	sess.Status = "busy"
	sess.WaitingFor = ""
	sess.LastUpdate = time.Now()
}

// MarkSessionIdle marks a session as idle (waiting for user input).
func (b *BotBridge) MarkSessionIdle(sessionID string) {
	if !b.enabled || sessionID == "" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if sess, exists := b.sessions[sessionID]; exists {
		sess.Status = "idle"
		sess.WaitingFor = "input"
		sess.LastUpdate = time.Now()
	}
}

// RemoveSession removes a session from the bridge.
func (b *BotBridge) RemoveSession(sessionID string) {
	if !b.enabled || sessionID == "" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.sessions, sessionID)
}

// GetSessions returns all active sessions for the bot to display.
func (b *BotBridge) GetSessions() []*bridgeSession {
	if !b.enabled {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	sessions := make([]*bridgeSession, 0, len(b.sessions))
	for _, sess := range b.sessions {
		sessions = append(sessions, sess)
	}
	return sessions
}

// CleanupStale removes sessions that haven't been updated recently.
func (b *BotBridge) CleanupStale(maxAge time.Duration) int {
	if !b.enabled {
		return 0
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	count := 0
	for id, sess := range b.sessions {
		if now.Sub(sess.LastUpdate) > maxAge {
			delete(b.sessions, id)
			count++
		}
	}
	return count
}

// GetProcessInfo implements bot.SessionProvider interface.
func (b *BotBridge) GetProcessInfo() []*bot.ProcessInfo {
	return b.ToProcessInfo()
}

// ToProcessInfo converts bridge sessions to ProcessInfo for the bot.
func (b *BotBridge) ToProcessInfo() []*bot.ProcessInfo {
	if !b.enabled {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	infos := make([]*bot.ProcessInfo, 0, len(b.sessions))
	for _, sess := range b.sessions {
		name := sess.ClientType
		if name == "" {
			name = "claude"
		}
		if sess.SessionID != "" {
			// Use short session ID for display
			shortID := sess.SessionID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}
			name = name + "-" + shortID
		}

		info := &bot.ProcessInfo{
			ID:            sess.SessionID,
			Name:          name,
			Status:        sess.Status,
			WaitingFor:    sess.WaitingFor,
			LastMessage:   sess.LastMessage,
			MessageRole:   sess.MessageRole,
			TokensUsed:    sess.TokensUsed,
			TurnCount:     sess.TurnCount,
			StartTime:     sess.LastUpdate, // Approximate
			LastSeen:      sess.LastUpdate,
		}
		infos = append(infos, info)
	}
	return infos
}
