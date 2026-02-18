package agent

import (
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

func TestObservatory_RegisterSession(t *testing.T) {
	obs := NewObservatory(&config.ObservatoryConfig{
		Enabled:        true,
		StuckThreshold: 5,
		IdleTimeoutMin: 30,
	})

	session := obs.RegisterSession("test-session", "default", "claude", "/test/project")

	if session.ID != "test-session" {
		t.Errorf("Expected session ID test-session, got %s", session.ID)
	}
	if session.Status != SessionStatusActive {
		t.Errorf("Expected status active, got %s", session.Status)
	}

	// Get session
	retrieved := obs.GetSession("test-session")
	if retrieved == nil {
		t.Fatal("Expected to retrieve session")
	}
	if retrieved.Profile != "default" {
		t.Errorf("Expected profile default, got %s", retrieved.Profile)
	}
}

func TestObservatory_KillSession(t *testing.T) {
	obs := NewObservatory(&config.ObservatoryConfig{Enabled: true})
	obs.RegisterSession("test-session", "default", "claude", "")

	if !obs.KillSession("test-session") {
		t.Error("Expected KillSession to return true")
	}

	if !obs.IsSessionKilled("test-session") {
		t.Error("Expected session to be killed")
	}
}

func TestGuardrails_CheckRequest(t *testing.T) {
	gr := NewGuardrails(&config.GuardrailsConfig{
		Enabled:            true,
		SessionSpendingCap: 10.0,
		RequestRateLimit:   60,
		AutoPauseOnCap:     true,
	})

	// Should allow request
	allowed, reason := gr.CheckRequest("test-session")
	if !allowed {
		t.Errorf("Expected request to be allowed, got reason: %s", reason)
	}

	// Record spending up to cap
	gr.RecordSpending("test-session", 10.0)

	// Should block request
	allowed, reason = gr.CheckRequest("test-session")
	if allowed {
		t.Error("Expected request to be blocked after spending cap")
	}
	if reason != "session spending cap reached" {
		t.Errorf("Expected spending cap reason, got: %s", reason)
	}
}

func TestGuardrails_CheckSensitiveOperation(t *testing.T) {
	gr := NewGuardrails(&config.GuardrailsConfig{
		Enabled:            true,
		SensitiveOpsDetect: true,
	})

	// Test file deletion detection
	ops := gr.CheckSensitiveOperation("test-session", []byte(`rm -rf /some/path`))
	if len(ops) == 0 {
		t.Error("Expected to detect file deletion operation")
	}

	// Test database operation detection
	ops = gr.CheckSensitiveOperation("test-session", []byte(`DROP TABLE users`))
	if len(ops) == 0 {
		t.Error("Expected to detect database operation")
	}
}

func TestCoordinator_AcquireLock(t *testing.T) {
	coord := NewCoordinator(&config.CoordinatorConfig{
		Enabled:        true,
		LockTimeoutSec: 300,
	})

	// Acquire lock
	success, holder := coord.AcquireLock("/test/file.go", "session-1")
	if !success {
		t.Error("Expected to acquire lock")
	}
	if holder != "" {
		t.Errorf("Expected no holder, got %s", holder)
	}

	// Try to acquire same lock from different session
	success, holder = coord.AcquireLock("/test/file.go", "session-2")
	if success {
		t.Error("Expected lock acquisition to fail")
	}
	if holder != "session-1" {
		t.Errorf("Expected holder session-1, got %s", holder)
	}

	// Same session can extend lock
	success, _ = coord.AcquireLock("/test/file.go", "session-1")
	if !success {
		t.Error("Expected same session to extend lock")
	}

	// Release lock
	coord.ReleaseLock("/test/file.go", "session-1")

	// Now session-2 can acquire
	success, _ = coord.AcquireLock("/test/file.go", "session-2")
	if !success {
		t.Error("Expected session-2 to acquire lock after release")
	}
}

func TestCoordinator_RecordChange(t *testing.T) {
	coord := NewCoordinator(&config.CoordinatorConfig{Enabled: true})

	coord.RecordChange("/test/file.go", "session-1", ChangeTypeModify, "Updated function")

	changes := coord.GetRecentChanges(10)
	if len(changes) != 1 {
		t.Fatalf("Expected 1 change, got %d", len(changes))
	}
	if changes[0].Path != "/test/file.go" {
		t.Errorf("Expected path /test/file.go, got %s", changes[0].Path)
	}
}

func TestTaskQueue_AddAndGetTask(t *testing.T) {
	tq := NewTaskQueue(&config.TaskQueueConfig{
		Enabled:    true,
		MaxRetries: 3,
	})

	task := tq.AddTask("Test task", 5)
	if task.ID == "" {
		t.Error("Expected task ID to be set")
	}
	if task.Status != TaskStatusPending {
		t.Errorf("Expected status pending, got %s", task.Status)
	}

	// Get task
	retrieved := tq.GetTask(task.ID)
	if retrieved == nil {
		t.Fatal("Expected to retrieve task")
	}
	if retrieved.Description != "Test task" {
		t.Errorf("Expected description 'Test task', got %s", retrieved.Description)
	}
}

func TestTaskQueue_Priority(t *testing.T) {
	tq := NewTaskQueue(&config.TaskQueueConfig{Enabled: true})

	tq.AddTask("Low priority", 1)
	tq.AddTask("High priority", 10)
	tq.AddTask("Medium priority", 5)

	// Get next task should return highest priority
	next := tq.GetNextTask("worker-1")
	if next.Description != "High priority" {
		t.Errorf("Expected high priority task, got %s", next.Description)
	}
}

func TestTaskQueue_CompleteAndFail(t *testing.T) {
	tq := NewTaskQueue(&config.TaskQueueConfig{
		Enabled:    true,
		MaxRetries: 2,
	})

	task := tq.AddTask("Test task", 5)

	// Complete task
	tq.CompleteTask(task.ID, &TaskResult{Success: true, Output: "Done"})

	retrieved := tq.GetTask(task.ID)
	if retrieved.Status != TaskStatusCompleted {
		t.Errorf("Expected status completed, got %s", retrieved.Status)
	}

	// Test failure with retry
	task2 := tq.AddTask("Failing task", 5)
	tq.FailTask(task2.ID, &TaskResult{Success: false, Error: "Error 1"})

	retrieved2 := tq.GetTask(task2.ID)
	if retrieved2.Status != TaskStatusPending {
		t.Errorf("Expected status pending after first failure, got %s", retrieved2.Status)
	}
	if retrieved2.RetryCount != 1 {
		t.Errorf("Expected retry count 1, got %d", retrieved2.RetryCount)
	}

	// Fail again - should still be pending
	tq.FailTask(task2.ID, &TaskResult{Success: false, Error: "Error 2"})
	retrieved2 = tq.GetTask(task2.ID)
	if retrieved2.Status != TaskStatusFailed {
		t.Errorf("Expected status failed after max retries, got %s", retrieved2.Status)
	}
}

func TestRuntime_NewRuntime(t *testing.T) {
	rt := NewRuntime(&config.RuntimeConfig{
		Enabled:        true,
		PlanningModel:  "claude-opus",
		ExecutionModel: "claude-sonnet",
		MaxTurns:       50,
	}, 19841)

	if !rt.IsEnabled() {
		t.Error("Expected runtime to be enabled")
	}

	if rt.config.PlanningModel != "claude-opus" {
		t.Errorf("Expected planning model claude-opus, got %s", rt.config.PlanningModel)
	}
}

func TestObservatory_StuckDetection(t *testing.T) {
	obs := NewObservatory(&config.ObservatoryConfig{
		Enabled:        true,
		StuckThreshold: 3,
	})

	obs.RegisterSession("test-session", "default", "claude", "")

	// Record errors
	for i := 0; i < 3; i++ {
		obs.RecordRequest("test-session", 100, 0.01, &testError{"error"})
	}

	session := obs.GetSession("test-session")
	if session.Status != SessionStatusStuck {
		t.Errorf("Expected status stuck after %d errors, got %s", 3, session.Status)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestCoordinator_GenerateContextWarning(t *testing.T) {
	coord := NewCoordinator(&config.CoordinatorConfig{
		Enabled:        true,
		InjectWarnings: true,
		LockTimeoutSec: 300,
	})

	// Acquire lock from session-1
	coord.AcquireLock("/test/file.go", "session-1")

	// Record change from session-1
	coord.RecordChange("/test/other.go", "session-1", ChangeTypeModify, "Updated")

	// Generate warning for session-2
	warning := coord.GenerateContextWarning("session-2")

	if warning == "" {
		t.Error("Expected warning to be generated")
	}
	if !contains(warning, "file.go") {
		t.Error("Expected warning to mention locked file")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestObservatory_IdleTimeout(t *testing.T) {
	obs := NewObservatory(&config.ObservatoryConfig{
		Enabled:        true,
		IdleTimeoutMin: 0, // Will be set to default 30
	})

	session := obs.RegisterSession("test-session", "default", "claude", "")

	// Manually set last activity to past
	session.mu.Lock()
	session.LastActivity = time.Now().Add(-31 * time.Minute)
	session.mu.Unlock()

	obs.CheckIdleSessions()

	if session.Status != SessionStatusIdle {
		t.Errorf("Expected status idle, got %s", session.Status)
	}
}
