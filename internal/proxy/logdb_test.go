package proxy

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestLogDBInsertAndQuery(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	now := time.Now()
	entries := []LogEntry{
		{Timestamp: now, Level: LogLevelInfo, Provider: "p1", Message: "ok", Method: "POST", Path: "/v1/messages", StatusCode: 200},
		{Timestamp: now, Level: LogLevelError, Provider: "p2", Message: "fail", Method: "POST", Path: "/v1/messages", StatusCode: 500, Error: "server error"},
		{Timestamp: now, Level: LogLevelWarn, Provider: "p1", Message: "rate limited", StatusCode: 429},
	}
	for _, e := range entries {
		db.Insert(e)
	}

	// Wait for flush
	time.Sleep(700 * time.Millisecond)

	// Query all
	results, err := db.Query(LogFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d entries, want 3", len(results))
	}

	// Query by provider
	results, err = db.Query(LogFilter{Provider: "p1", Limit: 100})
	if err != nil {
		t.Fatalf("Query provider: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d entries for p1, want 2", len(results))
	}

	// Query errors only
	results, err = db.Query(LogFilter{ErrorsOnly: true, Limit: 100})
	if err != nil {
		t.Fatalf("Query errors: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d error entries, want 2 (error + warn)", len(results))
	}

	// Query by level
	results, err = db.Query(LogFilter{Level: LogLevelError, Limit: 100})
	if err != nil {
		t.Fatalf("Query level: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d error-level entries, want 1", len(results))
	}

	// Query by status code
	results, err = db.Query(LogFilter{StatusCode: 500, Limit: 100})
	if err != nil {
		t.Fatalf("Query status: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d entries with status 500, want 1", len(results))
	}

	// Query by status range
	results, err = db.Query(LogFilter{StatusMin: 400, StatusMax: 599, Limit: 100})
	if err != nil {
		t.Fatalf("Query status range: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d entries in 4xx-5xx, want 2", len(results))
	}

	// Query with limit
	results, err = db.Query(LogFilter{Limit: 1})
	if err != nil {
		t.Fatalf("Query limit: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d entries with limit 1, want 1", len(results))
	}
}

func TestLogDBNewestFirst(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.Insert(LogEntry{Timestamp: now.Add(-2 * time.Second), Level: LogLevelInfo, Message: "first"})
	db.Insert(LogEntry{Timestamp: now.Add(-1 * time.Second), Level: LogLevelInfo, Message: "second"})
	db.Insert(LogEntry{Timestamp: now, Level: LogLevelInfo, Message: "third"})

	time.Sleep(700 * time.Millisecond)

	results, err := db.Query(LogFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d entries, want 3", len(results))
	}
	if results[0].Message != "third" {
		t.Errorf("first result = %q, want 'third' (newest first)", results[0].Message)
	}
	if results[2].Message != "first" {
		t.Errorf("last result = %q, want 'first' (oldest last)", results[2].Message)
	}
}

func TestLogDBGetProviders(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	db.Insert(LogEntry{Timestamp: time.Now(), Level: LogLevelInfo, Provider: "alpha"})
	db.Insert(LogEntry{Timestamp: time.Now(), Level: LogLevelInfo, Provider: "beta"})
	db.Insert(LogEntry{Timestamp: time.Now(), Level: LogLevelInfo, Provider: "alpha"}) // duplicate
	db.Insert(LogEntry{Timestamp: time.Now(), Level: LogLevelInfo, Provider: ""})      // empty

	time.Sleep(700 * time.Millisecond)

	providers, err := db.GetProviders()
	if err != nil {
		t.Fatalf("GetProviders: %v", err)
	}
	if len(providers) != 2 {
		t.Errorf("got %d providers, want 2", len(providers))
	}

	providerSet := make(map[string]bool)
	for _, p := range providers {
		providerSet[p] = true
	}
	if !providerSet["alpha"] || !providerSet["beta"] {
		t.Errorf("providers = %v, want [alpha, beta]", providers)
	}
}

func TestLogDBBatchFlush(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	// Insert more than 50 entries to trigger batch flush by count
	for i := 0; i < 60; i++ {
		db.Insert(LogEntry{Timestamp: time.Now(), Level: LogLevelInfo, Message: "batch"})
	}

	// Give time for batch to flush
	time.Sleep(200 * time.Millisecond)

	results, err := db.Query(LogFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	// At least 50 should be flushed (the batch threshold)
	if len(results) < 50 {
		t.Errorf("got %d entries, want >= 50 (batch threshold)", len(results))
	}
}

func TestLogDBCreatesFile(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	db.Close()

	dbPath := filepath.Join(dir, "logs.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("logs.db should exist after OpenLogDB")
	}
}

func TestLogDBCloseFlushesRemaining(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}

	// Insert a few entries (less than batch threshold)
	for i := 0; i < 5; i++ {
		db.Insert(LogEntry{Timestamp: time.Now(), Level: LogLevelInfo, Message: "flush-on-close"})
	}

	// Close should flush remaining entries
	db.Close()

	// Reopen and verify
	db2, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB reopen: %v", err)
	}
	defer db2.Close()

	results, err := db2.Query(LogFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("got %d entries after close+reopen, want 5", len(results))
	}
}

func TestLogDBResponseBody(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	db.Insert(LogEntry{
		Timestamp:    time.Now(),
		Level:        LogLevelError,
		Provider:     "p1",
		Message:      "error with body",
		StatusCode:   500,
		ResponseBody: `{"error":"internal server error"}`,
	})

	time.Sleep(700 * time.Millisecond)

	results, err := db.Query(LogFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d entries, want 1", len(results))
	}
	if results[0].ResponseBody != `{"error":"internal server error"}` {
		t.Errorf("response_body = %q, want error JSON", results[0].ResponseBody)
	}
	if results[0].Error != "" {
		t.Errorf("error = %q, want empty", results[0].Error)
	}
}

func TestLogDBSessionAndClientType(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.Insert(LogEntry{Timestamp: now, Level: LogLevelInfo, Provider: "p1", Message: "claude req", SessionID: "default:abc123", ClientType: "claude"})
	db.Insert(LogEntry{Timestamp: now, Level: LogLevelInfo, Provider: "p1", Message: "codex req", SessionID: "work:def456", ClientType: "codex"})
	db.Insert(LogEntry{Timestamp: now, Level: LogLevelInfo, Provider: "p2", Message: "no session", SessionID: "", ClientType: ""})

	time.Sleep(700 * time.Millisecond)

	// Query by client type
	results, err := db.Query(LogFilter{ClientType: "claude", Limit: 100})
	if err != nil {
		t.Fatalf("Query client_type: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d entries for claude, want 1", len(results))
	}
	if results[0].SessionID != "default:abc123" {
		t.Errorf("session_id = %q, want 'default:abc123'", results[0].SessionID)
	}

	// Query by session ID
	results, err = db.Query(LogFilter{SessionID: "work:def456", Limit: 100})
	if err != nil {
		t.Fatalf("Query session_id: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d entries for session, want 1", len(results))
	}
	if results[0].ClientType != "codex" {
		t.Errorf("client_type = %q, want 'codex'", results[0].ClientType)
	}

	// Query all — should return 3
	results, err = db.Query(LogFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Query all: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("got %d entries, want 3", len(results))
	}
}

func TestLogDBSchemaMigration(t *testing.T) {
	// Create an old-schema DB without session_id and client_type columns
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "logs.db")

	oldDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open old db: %v", err)
	}
	oldDB.Exec("PRAGMA journal_mode=WAL")
	oldDB.Exec(`CREATE TABLE logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		level TEXT NOT NULL,
		provider TEXT DEFAULT '',
		message TEXT DEFAULT '',
		status_code INTEGER DEFAULT 0,
		method TEXT DEFAULT '',
		path TEXT DEFAULT '',
		error TEXT DEFAULT '',
		response_body TEXT DEFAULT ''
	)`)
	oldDB.Exec(`INSERT INTO logs (timestamp, level, provider, message, status_code)
		VALUES ('2024-01-01T00:00:00Z', 'info', 'p1', 'old entry', 200)`)
	oldDB.Close()

	// Open with new code — should auto-migrate
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB migration: %v", err)
	}
	defer db.Close()

	// Insert a new entry with session_id and client_type
	db.Insert(LogEntry{
		Timestamp:  time.Now(),
		Level:      LogLevelInfo,
		Provider:   "p2",
		Message:    "new entry",
		SessionID:  "default:abc",
		ClientType: "claude",
	})
	time.Sleep(700 * time.Millisecond)

	// Query all — should see both old and new entries
	results, err := db.Query(LogFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d entries, want 2", len(results))
	}

	// Old entry should have empty session_id and client_type
	for _, r := range results {
		if r.Message == "old entry" {
			if r.SessionID != "" || r.ClientType != "" {
				t.Errorf("old entry should have empty session/client, got %q/%q", r.SessionID, r.ClientType)
			}
		}
		if r.Message == "new entry" {
			if r.SessionID != "default:abc" || r.ClientType != "claude" {
				t.Errorf("new entry session=%q client=%q, want default:abc/claude", r.SessionID, r.ClientType)
			}
		}
	}
}
