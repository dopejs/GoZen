package middleware

import (
	"encoding/json"
	"log"
	"os"
	"testing"
)

func TestPipeline_AddRemove(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	pipeline := NewPipeline(logger)

	// Add middleware
	m := NewRequestLogger()
	if err := m.Init(nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	pipeline.Add(m)

	if len(pipeline.List()) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(pipeline.List()))
	}

	// Remove middleware
	pipeline.Remove("request-logger")
	if len(pipeline.List()) != 0 {
		t.Errorf("Expected 0 middlewares, got %d", len(pipeline.List()))
	}
}

func TestPipeline_Priority(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	pipeline := NewPipeline(logger)

	// Add middlewares in reverse priority order
	logger1 := NewRequestLogger()
	logger1.Init(nil)
	pipeline.Add(logger1)

	injection := NewContextInjection()
	injection.Init(nil)
	pipeline.Add(injection)

	list := pipeline.List()
	if len(list) != 2 {
		t.Fatalf("Expected 2 middlewares, got %d", len(list))
	}

	// Context injection (priority 10) should come before request logger (priority 20)
	if list[0].Name() != "context-injection" {
		t.Errorf("Expected context-injection first, got %s", list[0].Name())
	}
	if list[1].Name() != "request-logger" {
		t.Errorf("Expected request-logger second, got %s", list[1].Name())
	}
}

func TestPipeline_ProcessRequest(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	pipeline := NewPipeline(logger)
	pipeline.SetEnabled(true)

	m := NewRequestLogger()
	m.Init(nil)
	pipeline.Add(m)

	ctx := NewRequestContext()
	ctx.SessionID = "test-session"
	ctx.Method = "POST"
	ctx.Path = "/v1/messages"
	ctx.Body = []byte(`{"model": "claude-3", "messages": []}`)

	result, err := pipeline.ProcessRequest(ctx)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if result.SessionID != "test-session" {
		t.Errorf("Expected session ID test-session, got %s", result.SessionID)
	}
}

func TestPipeline_Disabled(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	pipeline := NewPipeline(logger)
	pipeline.SetEnabled(false)

	ctx := NewRequestContext()
	result, err := pipeline.ProcessRequest(ctx)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	// Should return same context when disabled
	if result != ctx {
		t.Error("Expected same context when pipeline is disabled")
	}
}

func TestRequestContext_Clone(t *testing.T) {
	ctx := NewRequestContext()
	ctx.SessionID = "original"
	ctx.Body = []byte("original body")
	ctx.Headers.Set("X-Test", "value")
	ctx.Metadata["key"] = "value"

	clone := ctx.Clone()

	// Modify original
	ctx.SessionID = "modified"
	ctx.Body[0] = 'X'
	ctx.Headers.Set("X-Test", "modified")
	ctx.Metadata["key"] = "modified"

	// Clone should be unchanged
	if clone.SessionID != "original" {
		t.Errorf("Clone SessionID was modified")
	}
	if string(clone.Body) != "original body" {
		t.Errorf("Clone Body was modified")
	}
	if clone.Headers.Get("X-Test") != "value" {
		t.Errorf("Clone Headers was modified")
	}
	if clone.Metadata["key"] != "value" {
		t.Errorf("Clone Metadata was modified")
	}
}

func TestRequestLoggerMiddleware(t *testing.T) {
	m := NewRequestLogger()

	if m.Name() != "request-logger" {
		t.Errorf("Expected name request-logger, got %s", m.Name())
	}
	if m.Priority() != 20 {
		t.Errorf("Expected priority 20, got %d", m.Priority())
	}

	// Test init with config
	cfg := json.RawMessage(`{"log_body": true, "max_body_size": 500}`)
	if err := m.Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test process request
	ctx := NewRequestContext()
	ctx.SessionID = "test"
	ctx.Body = []byte(`{"test": true}`)

	result, err := m.ProcessRequest(ctx)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}
	if result == nil {
		t.Fatal("ProcessRequest returned nil")
	}

	m.Close()
}

func TestContextInjectionMiddleware(t *testing.T) {
	m := NewContextInjection()

	if m.Name() != "context-injection" {
		t.Errorf("Expected name context-injection, got %s", m.Name())
	}
	if m.Priority() != 10 {
		t.Errorf("Expected priority 10, got %d", m.Priority())
	}

	if err := m.Init(nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test process request without project path (should pass through)
	ctx := NewRequestContext()
	ctx.Body = []byte(`{"messages": [{"role": "user", "content": "hello"}]}`)

	result, err := m.ProcessRequest(ctx)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}
	if result == nil {
		t.Fatal("ProcessRequest returned nil")
	}
}

func TestRegistry_AvailableBuiltins(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	registry := NewRegistry(logger)

	builtins := registry.AvailableBuiltins()
	if len(builtins) < 4 {
		t.Errorf("Expected at least 4 builtins, got %d", len(builtins))
	}

	// Check that expected builtins are registered
	found := make(map[string]bool)
	for _, name := range builtins {
		found[name] = true
	}

	if !found["request-logger"] {
		t.Error("Expected request-logger builtin")
	}
	if !found["context-injection"] {
		t.Error("Expected context-injection builtin")
	}
	if !found["session-memory"] {
		t.Error("Expected session-memory builtin")
	}
	if !found["orchestration"] {
		t.Error("Expected orchestration builtin")
	}
}

func TestSessionMemoryMiddleware(t *testing.T) {
	m := NewSessionMemory()

	if m.Name() != "session-memory" {
		t.Errorf("Expected name session-memory, got %s", m.Name())
	}
	if m.Priority() != 15 {
		t.Errorf("Expected priority 15, got %d", m.Priority())
	}

	if err := m.Init(nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test process request without project path (should pass through)
	ctx := NewRequestContext()
	ctx.Body = []byte(`{"messages": [{"role": "user", "content": "hello"}]}`)

	result, err := m.ProcessRequest(ctx)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}
	if result == nil {
		t.Fatal("ProcessRequest returned nil")
	}

	m.Close()
}

func TestOrchestrationMiddleware(t *testing.T) {
	m := NewOrchestration()

	if m.Name() != "orchestration" {
		t.Errorf("Expected name orchestration, got %s", m.Name())
	}
	if m.Priority() != 50 {
		t.Errorf("Expected priority 50, got %d", m.Priority())
	}

	// Test init with config
	cfg := json.RawMessage(`{"default_mode": "single", "timeout": 60}`)
	if err := m.Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test process request
	ctx := NewRequestContext()
	ctx.Body = []byte(`{"model": "claude-3", "messages": []}`)
	ctx.Metadata = make(map[string]interface{})

	result, err := m.ProcessRequest(ctx)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}
	if result == nil {
		t.Fatal("ProcessRequest returned nil")
	}

	// Check metadata was set
	if result.Metadata["orchestration_mode"] != "single" {
		t.Error("Expected orchestration_mode to be set in metadata")
	}

	m.Close()
}
