package proxy

import (
	"testing"
	"time"
)

func TestBotBridge_UpdateSession(t *testing.T) {
	bridge := &BotBridge{
		sessions: make(map[string]*bridgeSession),
		enabled:  true,
	}

	// Test updating a new session
	bridge.UpdateSession("sess-1", "claude-code", nil, "Hello", "user", "")

	if len(bridge.sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(bridge.sessions))
	}

	sess := bridge.sessions["sess-1"]
	if sess.LastMessage != "Hello" {
		t.Errorf("expected LastMessage 'Hello', got '%s'", sess.LastMessage)
	}
	if sess.MessageRole != "user" {
		t.Errorf("expected MessageRole 'user', got '%s'", sess.MessageRole)
	}
	if sess.ClientType != "claude-code" {
		t.Errorf("expected ClientType 'claude-code', got '%s'", sess.ClientType)
	}
}

func TestBotBridge_MarkSessionBusy(t *testing.T) {
	bridge := &BotBridge{
		sessions: make(map[string]*bridgeSession),
		enabled:  true,
	}

	bridge.MarkSessionBusy("sess-1", "claude-code")

	sess := bridge.sessions["sess-1"]
	if sess.Status != "busy" {
		t.Errorf("expected status 'busy', got '%s'", sess.Status)
	}
}

func TestBotBridge_MarkSessionIdle(t *testing.T) {
	bridge := &BotBridge{
		sessions: make(map[string]*bridgeSession),
		enabled:  true,
	}

	// First create a session
	bridge.MarkSessionBusy("sess-1", "claude-code")
	bridge.MarkSessionIdle("sess-1")

	sess := bridge.sessions["sess-1"]
	if sess.Status != "idle" {
		t.Errorf("expected status 'idle', got '%s'", sess.Status)
	}
	if sess.WaitingFor != "input" {
		t.Errorf("expected WaitingFor 'input', got '%s'", sess.WaitingFor)
	}
}

func TestBotBridge_ToProcessInfo(t *testing.T) {
	bridge := &BotBridge{
		sessions: make(map[string]*bridgeSession),
		enabled:  true,
	}

	bridge.sessions["sess-12345678"] = &bridgeSession{
		SessionID:   "sess-12345678",
		ClientType:  "claude-code",
		Status:      "idle",
		WaitingFor:  "input",
		LastMessage: "What files are in this directory?",
		MessageRole: "user",
		TokensUsed:  1500,
		TurnCount:   3,
		LastUpdate:  time.Now(),
	}

	infos := bridge.ToProcessInfo()
	if len(infos) != 1 {
		t.Fatalf("expected 1 process info, got %d", len(infos))
	}

	info := infos[0]
	if info.Status != "idle" {
		t.Errorf("expected status 'idle', got '%s'", info.Status)
	}
	if info.WaitingFor != "input" {
		t.Errorf("expected WaitingFor 'input', got '%s'", info.WaitingFor)
	}
	if info.TurnCount != 3 {
		t.Errorf("expected TurnCount 3, got %d", info.TurnCount)
	}
}

func TestBotBridge_CleanupStale(t *testing.T) {
	bridge := &BotBridge{
		sessions: make(map[string]*bridgeSession),
		enabled:  true,
	}

	// Add an old session
	bridge.sessions["old-sess"] = &bridgeSession{
		SessionID:  "old-sess",
		LastUpdate: time.Now().Add(-2 * time.Hour),
	}

	// Add a recent session
	bridge.sessions["new-sess"] = &bridgeSession{
		SessionID:  "new-sess",
		LastUpdate: time.Now(),
	}

	removed := bridge.CleanupStale(1 * time.Hour)
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	if len(bridge.sessions) != 1 {
		t.Errorf("expected 1 session remaining, got %d", len(bridge.sessions))
	}

	if _, exists := bridge.sessions["new-sess"]; !exists {
		t.Error("expected new-sess to remain")
	}
}

func TestBotBridge_Disabled(t *testing.T) {
	bridge := &BotBridge{
		sessions: make(map[string]*bridgeSession),
		enabled:  false,
	}

	// Operations should be no-ops when disabled
	bridge.UpdateSession("sess-1", "claude-code", nil, "Hello", "user", "")
	bridge.MarkSessionBusy("sess-2", "claude-code")

	if len(bridge.sessions) != 0 {
		t.Errorf("expected 0 sessions when disabled, got %d", len(bridge.sessions))
	}

	infos := bridge.ToProcessInfo()
	if infos != nil {
		t.Errorf("expected nil process info when disabled, got %v", infos)
	}
}
