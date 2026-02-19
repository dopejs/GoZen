package bot

import (
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/bot/adapters"
)

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager()
	if sm == nil {
		t.Fatal("NewSessionManager returned nil")
	}
	if sm.sessions == nil {
		t.Error("sessions map is nil")
	}
}

func TestSessionManager_Get(t *testing.T) {
	sm := NewSessionManager()

	// Get non-existent session
	s := sm.Get(adapters.PlatformTelegram, "user1")
	if s != nil {
		t.Error("Get should return nil for non-existent session")
	}

	// Create session and get it
	sm.GetOrCreate(adapters.PlatformTelegram, "user1", "chat1")
	s = sm.Get(adapters.PlatformTelegram, "user1")
	if s == nil {
		t.Error("Get should return session after creation")
	}
}

func TestSessionManager_GetOrCreate(t *testing.T) {
	sm := NewSessionManager()

	// Create new session
	s1 := sm.GetOrCreate(adapters.PlatformTelegram, "user1", "chat1")
	if s1 == nil {
		t.Fatal("GetOrCreate returned nil")
	}
	if s1.UserID != "user1" {
		t.Errorf("expected UserID 'user1', got '%s'", s1.UserID)
	}
	if s1.Platform != adapters.PlatformTelegram {
		t.Errorf("expected Platform Telegram, got %v", s1.Platform)
	}

	// Get existing session
	s2 := sm.GetOrCreate(adapters.PlatformTelegram, "user1", "chat1")
	if s1 != s2 {
		t.Error("GetOrCreate should return same session for same user")
	}

	// Different user should get different session
	s3 := sm.GetOrCreate(adapters.PlatformTelegram, "user2", "chat1")
	if s1 == s3 {
		t.Error("Different users should get different sessions")
	}
}

func TestSessionManager_GetOrCreate_UpdatesChatID(t *testing.T) {
	sm := NewSessionManager()

	s1 := sm.GetOrCreate(adapters.PlatformTelegram, "user1", "chat1")
	if s1.ChatID != "chat1" {
		t.Errorf("expected ChatID 'chat1', got '%s'", s1.ChatID)
	}

	// Get with different chat ID should update
	s2 := sm.GetOrCreate(adapters.PlatformTelegram, "user1", "chat2")
	if s2.ChatID != "chat2" {
		t.Errorf("expected ChatID 'chat2', got '%s'", s2.ChatID)
	}
}

func TestSessionManager_Bind(t *testing.T) {
	sm := NewSessionManager()

	sm.GetOrCreate(adapters.PlatformTelegram, "user1", "chat1")
	sm.Bind(adapters.PlatformTelegram, "user1", "myprocess")

	s := sm.GetOrCreate(adapters.PlatformTelegram, "user1", "chat1")
	if s.BoundProcess != "myprocess" {
		t.Errorf("expected BoundProcess 'myprocess', got '%s'", s.BoundProcess)
	}
}

func TestSessionManager_Bind_CreatesSession(t *testing.T) {
	sm := NewSessionManager()

	// Bind without existing session should create one
	sm.Bind(adapters.PlatformDiscord, "user2", "process2")

	s := sm.Get(adapters.PlatformDiscord, "user2")
	if s == nil {
		t.Fatal("Bind should create session if not exists")
	}
	if s.BoundProcess != "process2" {
		t.Errorf("expected BoundProcess 'process2', got '%s'", s.BoundProcess)
	}
}

func TestSessionManager_Unbind(t *testing.T) {
	sm := NewSessionManager()

	sm.GetOrCreate(adapters.PlatformTelegram, "user1", "chat1")
	sm.Bind(adapters.PlatformTelegram, "user1", "myprocess")
	sm.Unbind(adapters.PlatformTelegram, "user1")

	s := sm.Get(adapters.PlatformTelegram, "user1")
	if s.BoundProcess != "" {
		t.Errorf("expected empty BoundProcess after unbind, got '%s'", s.BoundProcess)
	}
}

func TestSessionManager_Unbind_NonExistent(t *testing.T) {
	sm := NewSessionManager()

	// Unbind non-existent session should not panic
	sm.Unbind(adapters.PlatformTelegram, "nonexistent")
}

func TestSessionManager_GetBoundProcess(t *testing.T) {
	sm := NewSessionManager()

	// Non-existent user
	proc := sm.GetBoundProcess(adapters.PlatformTelegram, "user1")
	if proc != "" {
		t.Errorf("expected empty string for non-existent user, got '%s'", proc)
	}

	// Create and bind
	sm.GetOrCreate(adapters.PlatformTelegram, "user1", "chat1")
	sm.Bind(adapters.PlatformTelegram, "user1", "myprocess")

	proc = sm.GetBoundProcess(adapters.PlatformTelegram, "user1")
	if proc != "myprocess" {
		t.Errorf("expected 'myprocess', got '%s'", proc)
	}
}

func TestSessionManager_Cleanup(t *testing.T) {
	sm := NewSessionManager()

	// Create a session with old LastActive
	s := sm.GetOrCreate(adapters.PlatformTelegram, "user1", "chat1")
	s.LastActive = time.Now().Add(-48 * time.Hour)

	// Cleanup sessions older than 24 hours
	count := sm.Cleanup(24 * time.Hour)

	if count != 1 {
		t.Errorf("expected 1 cleaned up, got %d", count)
	}

	// Session should be removed
	if len(sm.sessions) != 0 {
		t.Errorf("expected 0 sessions after cleanup, got %d", len(sm.sessions))
	}
}

func TestSessionManager_Cleanup_KeepsRecent(t *testing.T) {
	sm := NewSessionManager()

	// Create a recent session
	sm.GetOrCreate(adapters.PlatformTelegram, "user1", "chat1")

	// Cleanup should not remove recent sessions
	count := sm.Cleanup(24 * time.Hour)

	if count != 0 {
		t.Errorf("expected 0 cleaned up, got %d", count)
	}

	if len(sm.sessions) != 1 {
		t.Errorf("expected 1 session to remain, got %d", len(sm.sessions))
	}
}

func TestSessionKey(t *testing.T) {
	key := sessionKey(adapters.PlatformTelegram, "user123")
	expected := "telegram:user123"
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestNewApprovalManager(t *testing.T) {
	am := NewApprovalManager()
	if am == nil {
		t.Fatal("NewApprovalManager returned nil")
	}
	if am.pending == nil {
		t.Error("pending map is nil")
	}
}

func TestApprovalManager_AddGetRemove(t *testing.T) {
	am := NewApprovalManager()

	approval := &PendingApproval{
		ID:        "approval-1",
		ProcessID: "process-1",
		MessageID: "msg-1",
		CreatedAt: time.Now(),
	}

	am.Add(approval)

	// Get by ID
	got := am.Get("approval-1")
	if got == nil {
		t.Fatal("Get returned nil for existing approval")
	}
	if got.ID != "approval-1" {
		t.Errorf("expected ID 'approval-1', got '%s'", got.ID)
	}

	// Get non-existent
	got = am.Get("non-existent")
	if got != nil {
		t.Error("Get should return nil for non-existent approval")
	}

	// Remove
	am.Remove("approval-1")
	got = am.Get("approval-1")
	if got != nil {
		t.Error("Get should return nil after removal")
	}
}

func TestApprovalManager_Add_NoMessageID(t *testing.T) {
	am := NewApprovalManager()

	approval := &PendingApproval{
		ID:        "approval-1",
		ProcessID: "process-1",
		MessageID: "", // No message ID
		CreatedAt: time.Now(),
	}

	am.Add(approval)

	// Should still be retrievable by ID
	got := am.Get("approval-1")
	if got == nil {
		t.Fatal("Get returned nil for existing approval")
	}

	// Should not be in byMsgID map
	got = am.GetByMessageID("")
	if got != nil {
		t.Error("GetByMessageID should return nil for empty message ID")
	}
}

func TestApprovalManager_Remove_NonExistent(t *testing.T) {
	am := NewApprovalManager()

	// Remove non-existent should not panic
	am.Remove("non-existent")
}

func TestApprovalManager_Remove_CleansUpMsgID(t *testing.T) {
	am := NewApprovalManager()

	approval := &PendingApproval{
		ID:        "approval-1",
		ProcessID: "process-1",
		MessageID: "msg-123",
		CreatedAt: time.Now(),
	}

	am.Add(approval)

	// Verify it's in byMsgID
	if am.GetByMessageID("msg-123") == nil {
		t.Fatal("approval should be findable by message ID")
	}

	am.Remove("approval-1")

	// Should be removed from byMsgID too
	if am.GetByMessageID("msg-123") != nil {
		t.Error("approval should be removed from byMsgID after Remove")
	}
}

func TestApprovalManager_GetByMessageID(t *testing.T) {
	am := NewApprovalManager()

	approval := &PendingApproval{
		ID:        "approval-1",
		ProcessID: "process-1",
		MessageID: "msg-123",
		CreatedAt: time.Now(),
	}

	am.Add(approval)

	got := am.GetByMessageID("msg-123")
	if got == nil {
		t.Fatal("GetByMessageID returned nil")
	}
	if got.ID != "approval-1" {
		t.Errorf("expected ID 'approval-1', got '%s'", got.ID)
	}

	got = am.GetByMessageID("non-existent")
	if got != nil {
		t.Error("GetByMessageID should return nil for non-existent message")
	}
}

func TestApprovalManager_Cleanup(t *testing.T) {
	am := NewApprovalManager()

	// Add expired approval
	expired := &PendingApproval{
		ID:        "expired-1",
		ProcessID: "process-1",
		MessageID: "msg-expired",
		CreatedAt: time.Now().Add(-time.Hour),
		Timeout:   time.Now().Add(-30 * time.Minute), // Already expired
	}
	am.Add(expired)

	// Add non-expired approval
	valid := &PendingApproval{
		ID:        "valid-1",
		ProcessID: "process-1",
		CreatedAt: time.Now(),
		Timeout:   time.Now().Add(time.Hour), // Not expired
	}
	am.Add(valid)

	// Add approval with no timeout (zero time)
	noTimeout := &PendingApproval{
		ID:        "no-timeout-1",
		ProcessID: "process-1",
		CreatedAt: time.Now(),
		Timeout:   time.Time{}, // No timeout
	}
	am.Add(noTimeout)

	expiredIDs := am.Cleanup()

	// Should return expired ID
	if len(expiredIDs) != 1 || expiredIDs[0] != "expired-1" {
		t.Errorf("expected [expired-1], got %v", expiredIDs)
	}

	// Expired should be removed
	if am.Get("expired-1") != nil {
		t.Error("Expired approval should be removed")
	}

	// Expired message ID should be cleaned up
	if am.GetByMessageID("msg-expired") != nil {
		t.Error("Expired approval message ID should be cleaned up")
	}

	// Valid should remain
	if am.Get("valid-1") == nil {
		t.Error("Valid approval should remain")
	}

	// No timeout should remain
	if am.Get("no-timeout-1") == nil {
		t.Error("Approval with no timeout should remain")
	}
}
