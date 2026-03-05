package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/bot"
)

// --- chatSessionStore unit tests ---

func TestChatSessionStore_CreateAndGet(t *testing.T) {
	store := &chatSessionStore{sessions: make(map[string]*ChatSession)}

	session := store.Create()
	if session == nil {
		t.Fatal("Create returned nil")
	}
	if session.ID == "" {
		t.Error("session ID should not be empty")
	}
	if len(session.Messages) != 0 {
		t.Error("new session should have no messages")
	}

	got := store.Get(session.ID)
	if got == nil {
		t.Fatal("Get returned nil for existing session")
	}
	if got.ID != session.ID {
		t.Errorf("expected ID %s, got %s", session.ID, got.ID)
	}
}

func TestChatSessionStore_GetNonExistent(t *testing.T) {
	store := &chatSessionStore{sessions: make(map[string]*ChatSession)}
	if store.Get("nonexistent") != nil {
		t.Error("Get should return nil for nonexistent session")
	}
}

func TestChatSessionStore_AddMessage(t *testing.T) {
	store := &chatSessionStore{sessions: make(map[string]*ChatSession)}
	session := store.Create()

	store.AddMessage(session.ID, bot.ChatMessage{Role: "user", Content: "hello"})
	store.AddMessage(session.ID, bot.ChatMessage{Role: "assistant", Content: "hi"})

	got := store.Get(session.ID)
	if len(got.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(got.Messages))
	}
	if got.Messages[0].Content != "hello" {
		t.Errorf("expected first message 'hello', got %q", got.Messages[0].Content)
	}
}

func TestChatSessionStore_AddMessage_NonExistent(t *testing.T) {
	store := &chatSessionStore{sessions: make(map[string]*ChatSession)}
	// Should not panic
	store.AddMessage("nonexistent", bot.ChatMessage{Role: "user", Content: "hello"})
}

func TestChatSessionStore_Clear(t *testing.T) {
	store := &chatSessionStore{sessions: make(map[string]*ChatSession)}
	session := store.Create()
	store.AddMessage(session.ID, bot.ChatMessage{Role: "user", Content: "hello"})

	store.Clear(session.ID)

	got := store.Get(session.ID)
	if len(got.Messages) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(got.Messages))
	}
}

func TestChatSessionStore_Clear_NonExistent(t *testing.T) {
	store := &chatSessionStore{sessions: make(map[string]*ChatSession)}
	// Should not panic
	store.Clear("nonexistent")
}

func TestChatSessionStore_Cleanup(t *testing.T) {
	store := &chatSessionStore{sessions: make(map[string]*ChatSession)}
	old := store.Create()
	old.UpdatedAt = time.Now().Add(-2 * time.Hour)
	// Ensure distinct ID by sleeping briefly
	time.Sleep(time.Millisecond)
	recent := store.Create()

	store.Cleanup(1 * time.Hour)

	if store.Get(old.ID) != nil {
		t.Error("old session should have been cleaned up")
	}
	if store.Get(recent.ID) == nil {
		t.Error("recent session should not have been cleaned up")
	}
}

// --- sendSSE test ---

func TestSendSSE(t *testing.T) {
	w := httptest.NewRecorder()

	sendSSE(w, w, "delta", map[string]string{"content": "hello"})

	body := w.Body.String()
	if !strings.Contains(body, "event: delta\n") {
		t.Errorf("expected SSE event line, got %q", body)
	}
	if !strings.Contains(body, `"content":"hello"`) {
		t.Errorf("expected JSON data, got %q", body)
	}
}

// --- handleBotChat HTTP tests ---

func TestHandleBotChat_MethodNotAllowed(t *testing.T) {
	s := setupTestServerWithBot(t)
	w := doRequest(s, "GET", "/api/v1/bot/chat", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleBotChat_InvalidBody(t *testing.T) {
	s := setupTestServerWithBot(t)
	w := doRequestRaw(s, "POST", "/api/v1/bot/chat", []byte("not json"))
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleBotChat_ClearSession(t *testing.T) {
	s := setupTestServerWithBot(t)

	// Create a session first via the store
	session := globalChatSessions.Create()
	globalChatSessions.AddMessage(session.ID, bot.ChatMessage{Role: "user", Content: "hi"})

	body := map[string]interface{}{
		"session_id": session.ID,
		"clear":      true,
	}
	w := doRequest(s, "POST", "/api/v1/bot/chat", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "cleared" {
		t.Errorf("expected status 'cleared', got %q", resp["status"])
	}
	if resp["session_id"] != session.ID {
		t.Errorf("expected session_id %q, got %q", session.ID, resp["session_id"])
	}

	// Verify messages were cleared
	got := globalChatSessions.Get(session.ID)
	if len(got.Messages) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(got.Messages))
	}
}

func TestHandleBotChat_ClearNewSession(t *testing.T) {
	s := setupTestServerWithBot(t)

	// Clear without existing session_id — should create new session and clear it
	body := map[string]interface{}{
		"clear": true,
	}
	w := doRequest(s, "POST", "/api/v1/bot/chat", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "cleared" {
		t.Errorf("expected status 'cleared', got %q", resp["status"])
	}
	if resp["session_id"] == "" {
		t.Error("expected non-empty session_id")
	}
}

func TestHandleBotChat_MissingMessage(t *testing.T) {
	s := setupTestServerWithBot(t)

	body := map[string]interface{}{
		"message": "",
	}
	w := doRequest(s, "POST", "/api/v1/bot/chat", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleBotChat_BotNotConfigured(t *testing.T) {
	// Use regular test server (no bot config)
	s := setupTestServer(t)

	body := map[string]interface{}{
		"message": "hello",
	}
	w := doRequest(s, "POST", "/api/v1/bot/chat", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
