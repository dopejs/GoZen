package proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestStructuredLoggerBasic(t *testing.T) {
	dir := t.TempDir()

	logger, err := NewStructuredLogger(dir, 100, nil)
	if err != nil {
		t.Fatalf("NewStructuredLogger() error: %v", err)
	}
	defer logger.Close()

	// Test basic logging
	logger.Info("test-provider", "test info message")
	logger.Warn("test-provider", "test warn message")
	logger.Error("test-provider", "test error message")

	// Test request logging
	logger.RequestLog("provider1", "POST", "/v1/messages", 200, "success")
	logger.RequestError("provider2", "POST", "/v1/messages", fmt.Errorf("server error"))
	logger.RequestErrorWithResponse("provider3", "POST", "/v1/messages", 429, "rate limited", []byte("retry later"))

	// Verify entries are stored
	if !logger.HasEntries() {
		t.Error("expected entries")
	}

	entries := logger.GetEntries(LogFilter{})
	if len(entries) != 6 {
		t.Errorf("expected 6 entries, got %d", len(entries))
	}

	// Verify providers
	providers := logger.GetProviders()
	if len(providers) < 2 {
		t.Errorf("expected at least 2 providers, got %d", len(providers))
	}

	// Verify log files exist
	if _, err := os.Stat(filepath.Join(dir, "proxy.log")); err != nil {
		t.Errorf("proxy.log should exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "err.log")); err != nil {
		t.Errorf("err.log should exist: %v", err)
	}
}

func TestStructuredLoggerFilter(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewStructuredLogger(dir, 100, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	logger.RequestLog("p1", "POST", "/v1/messages", 200, "ok")
	logger.RequestLog("p2", "POST", "/v1/messages", 500, "fail")
	logger.Info("p1", "info msg")

	// Filter by provider
	entries := logger.GetEntries(LogFilter{Provider: "p1"})
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for p1, got %d", len(entries))
	}

	// Filter by level
	entries = logger.GetEntries(LogFilter{Level: LogLevelError})
	if len(entries) != 1 {
		t.Errorf("expected 1 error entry, got %d", len(entries))
	}

	// Filter by status code
	entries = logger.GetEntries(LogFilter{StatusCode: 500})
	if len(entries) != 1 {
		t.Errorf("expected 1 entry with status 500, got %d", len(entries))
	}
}

func TestStructuredLoggerMaxEntries(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewStructuredLogger(dir, 5, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	for i := 0; i < 10; i++ {
		logger.Info("p", "msg")
	}

	entries := logger.GetEntries(LogFilter{})
	if len(entries) > 5 {
		t.Errorf("expected max 5 entries, got %d", len(entries))
	}
}

func TestLogEntryToJSON(t *testing.T) {
	entry := LogEntry{
		Level:    LogLevelInfo,
		Provider: "p1",
		Message:  "test",
	}
	data, err := entry.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func TestStructuredLoggerFormatEntry(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewStructuredLogger(dir, 100, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	entry := LogEntry{
		Level:    LogLevelInfo,
		Provider: "test",
		Message:  "hello",
	}
	formatted := logger.formatEntry(entry)
	if len(formatted) == 0 {
		t.Error("expected non-empty formatted entry")
	}
}

func TestLogFilterMatch(t *testing.T) {
	entry := LogEntry{
		Level:      LogLevelError,
		Provider:   "p1",
		StatusCode: 500,
	}

	if !(LogFilter{}).Match(entry) {
		t.Error("empty filter should match all")
	}
	if !(LogFilter{Provider: "p1"}).Match(entry) {
		t.Error("should match provider p1")
	}
	if (LogFilter{Provider: "p2"}).Match(entry) {
		t.Error("should not match provider p2")
	}
	if !(LogFilter{Level: LogLevelError}).Match(entry) {
		t.Error("should match error level")
	}
	if (LogFilter{Level: LogLevelInfo}).Match(entry) {
		t.Error("should not match info level")
	}
	if !(LogFilter{StatusCode: 500}).Match(entry) {
		t.Error("should match status 500")
	}
	if (LogFilter{StatusCode: 200}).Match(entry) {
		t.Error("should not match status 200")
	}
}

func TestProviderGetEnvVarsForCLI(t *testing.T) {
	p := &Provider{
		EnvVars:       map[string]string{"SHARED": "1"},
		ClaudeEnvVars: map[string]string{"C": "c"},
		CodexEnvVars:  map[string]string{"X": "x"},
	}

	vars := p.GetEnvVarsForCLI("claude")
	if vars["C"] != "c" {
		t.Errorf("claude vars = %v", vars)
	}

	vars = p.GetEnvVarsForCLI("codex")
	if vars["X"] != "x" {
		t.Errorf("codex vars = %v", vars)
	}

	// Fallback to shared
	p2 := &Provider{EnvVars: map[string]string{"S": "s"}}
	vars = p2.GetEnvVarsForCLI("claude")
	if vars["S"] != "s" {
		t.Errorf("fallback vars = %v", vars)
	}
}

func TestInitGlobalLogger(t *testing.T) {
	dir := t.TempDir()
	if err := InitGlobalLogger(dir); err != nil {
		t.Fatalf("InitGlobalLogger() error: %v", err)
	}

	logger := GetGlobalLogger()
	if logger == nil {
		t.Fatal("expected non-nil global logger")
	}

	// Don't close — other tests may use the global logger
}

func TestCleanupOldSessions(t *testing.T) {
	// Add some sessions
	UpdateSessionUsage("old-1", &SessionUsage{InputTokens: 100, OutputTokens: 10})
	UpdateSessionUsage("old-2", &SessionUsage{InputTokens: 200, OutputTokens: 20})

	size, _ := GetCacheStats()
	if size < 2 {
		t.Errorf("expected at least 2 sessions, got %d", size)
	}

	// Cleanup with 0 max age removes nothing (sessions are fresh)
	CleanupOldSessions(1000)

	// Clean up test sessions
	ClearSessionUsage("old-1")
	ClearSessionUsage("old-2")
}

func TestEstimateTokensFromChars(t *testing.T) {
	body := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Hello, this is a test message with some content.",
			},
		},
	}
	tokens := estimateTokensFromChars(body)
	if tokens <= 0 {
		t.Errorf("expected positive token estimate, got %d", tokens)
	}

	// Test with system prompt
	body2 := map[string]interface{}{
		"system": "You are a helpful assistant.",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "hi"},
		},
	}
	tokens2 := estimateTokensFromChars(body2)
	if tokens2 <= 0 {
		t.Errorf("expected positive token estimate with system, got %d", tokens2)
	}
}
