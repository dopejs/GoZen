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
	db.Insert(LogEntry{Timestamp: time.Now(), Level: LogLevelInfo, Provider: "alpha"})
	db.Insert(LogEntry{Timestamp: time.Now(), Level: LogLevelInfo, Provider: ""})

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

	for i := 0; i < 60; i++ {
		db.Insert(LogEntry{Timestamp: time.Now(), Level: LogLevelInfo, Message: "batch"})
	}

	time.Sleep(200 * time.Millisecond)

	results, err := db.Query(LogFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
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

	for i := 0; i < 5; i++ {
		db.Insert(LogEntry{Timestamp: time.Now(), Level: LogLevelInfo, Message: "flush-on-close"})
	}

	db.Close()

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
	db.Insert(LogEntry{Timestamp: now, Level: LogLevelInfo, Provider: "p2", Message: "no session"})

	time.Sleep(700 * time.Millisecond)

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

	results, err = db.Query(LogFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Query all: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("got %d entries, want 3", len(results))
	}
}

// --- Schema migration tests ---

// createV1Database creates a v1 schema database (no schema_version table, no session_id/client_type).
func createV1Database(t *testing.T, dir string) {
	t.Helper()
	dbPath := filepath.Join(dir, "logs.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open v1 db: %v", err)
	}
	defer db.Close()
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec(`CREATE TABLE logs (
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
	db.Exec(`INSERT INTO logs (timestamp, level, provider, message, status_code)
		VALUES ('2024-01-01T00:00:00Z', 'info', 'p1', 'old entry', 200)`)
}

func TestSchemaMigrationV1ToV2(t *testing.T) {
	dir := t.TempDir()
	createV1Database(t, dir)

	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB migration: %v", err)
	}
	defer db.Close()

	// Insert a new entry with v2 fields
	db.Insert(LogEntry{
		Timestamp:  time.Now(),
		Level:      LogLevelInfo,
		Provider:   "p2",
		Message:    "new entry",
		SessionID:  "default:abc",
		ClientType: "claude",
	})
	time.Sleep(700 * time.Millisecond)

	results, err := db.Query(LogFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d entries, want 2", len(results))
	}

	for _, r := range results {
		if r.Message == "old entry" {
			if r.SessionID != "" || r.ClientType != "" {
				t.Errorf("old entry should have empty session/client, got %q/%q", r.SessionID, r.ClientType)
			}
		}
		if r.Message == "new entry" {
			if r.SessionID != "default:abc" || r.ClientType != "claude" {
				t.Errorf("new entry session=%q client=%q", r.SessionID, r.ClientType)
			}
		}
	}
}

func TestSchemaVersionSetOnFreshDB(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	db.Close()

	// Verify schema_version table has currentSchemaVersion
	rawDB, _ := sql.Open("sqlite", filepath.Join(dir, "logs.db"))
	defer rawDB.Close()

	ver := getSchemaVersion(rawDB)
	if ver != currentSchemaVersion {
		t.Errorf("fresh DB version = %d, want %d", ver, currentSchemaVersion)
	}
}

func TestSchemaVersionSetAfterMigration(t *testing.T) {
	dir := t.TempDir()
	createV1Database(t, dir)

	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	db.Close()

	rawDB, _ := sql.Open("sqlite", filepath.Join(dir, "logs.db"))
	defer rawDB.Close()

	ver := getSchemaVersion(rawDB)
	if ver != currentSchemaVersion {
		t.Errorf("migrated DB version = %d, want %d", ver, currentSchemaVersion)
	}
}

func TestSchemaVersionAlreadyCurrent(t *testing.T) {
	dir := t.TempDir()

	// First open creates fresh DB at current version
	db1, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	db1.Close()

	// Second open should be a no-op (no migration needed)
	db2, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	db2.Close()

	rawDB, _ := sql.Open("sqlite", filepath.Join(dir, "logs.db"))
	defer rawDB.Close()

	ver := getSchemaVersion(rawDB)
	if ver != currentSchemaVersion {
		t.Errorf("version after reopen = %d, want %d", ver, currentSchemaVersion)
	}
}

func TestSchemaVersionTooNew(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "logs.db")

	// Create a DB with a future schema version
	rawDB, _ := sql.Open("sqlite", dbPath)
	rawDB.Exec("CREATE TABLE schema_version (version INTEGER NOT NULL)")
	rawDB.Exec("INSERT INTO schema_version (version) VALUES (999)")
	rawDB.Exec(`CREATE TABLE logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		level TEXT NOT NULL
	)`)
	rawDB.Exec(`INSERT INTO logs (timestamp, level) VALUES ('2024-01-01T00:00:00Z', 'info')`)
	rawDB.Close()

	// Should auto-rebuild instead of failing
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("expected auto-rebuild, got error: %v", err)
	}
	defer db.Close()

	// Old data should be gone (rebuilt from scratch)
	results, _ := db.Query(LogFilter{Limit: 100})
	if len(results) != 0 {
		t.Errorf("expected 0 entries after rebuild, got %d", len(results))
	}

	// Version should be current
	ver := getSchemaVersion(db.db)
	if ver != currentSchemaVersion {
		t.Errorf("version after rebuild = %d, want %d", ver, currentSchemaVersion)
	}
}

func TestCorruptDatabaseAutoRebuilds(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "logs.db")

	// Write garbage to the file
	os.WriteFile(dbPath, []byte("this is not a sqlite database"), 0600)

	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("expected auto-rebuild for corrupt file, got error: %v", err)
	}
	defer db.Close()

	// Should work normally after rebuild
	db.Insert(LogEntry{Timestamp: time.Now(), Level: LogLevelInfo, Message: "after rebuild"})
	time.Sleep(700 * time.Millisecond)

	results, err := db.Query(LogFilter{Limit: 100})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 1 || results[0].Message != "after rebuild" {
		t.Errorf("got %v, want 1 entry 'after rebuild'", results)
	}
}

func TestBrokenSchemaAutoRebuilds(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "logs.db")

	// Create a DB with logs table that has wrong columns
	rawDB, _ := sql.Open("sqlite", dbPath)
	rawDB.Exec("CREATE TABLE schema_version (version INTEGER NOT NULL)")
	rawDB.Exec("INSERT INTO schema_version (version) VALUES (1)")
	rawDB.Exec("CREATE TABLE logs (id INTEGER, garbage TEXT)")
	rawDB.Close()

	// Migration will fail because ALTER TABLE on a mangled table — should rebuild
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("expected auto-rebuild, got error: %v", err)
	}
	defer db.Close()

	ver := getSchemaVersion(db.db)
	if ver != currentSchemaVersion {
		t.Errorf("version = %d, want %d", ver, currentSchemaVersion)
	}
}

func TestTableExists(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, _ := sql.Open("sqlite", dbPath)
	defer db.Close()

	if tableExists(db, "nonexistent") {
		t.Error("nonexistent table should not exist")
	}

	db.Exec("CREATE TABLE mytable (id INTEGER)")
	if !tableExists(db, "mytable") {
		t.Error("mytable should exist after creation")
	}
}
