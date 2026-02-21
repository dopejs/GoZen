package proxy

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Schema version history:
//   v1: original schema (logs table with basic fields)
//   v2: add session_id and client_type columns + indexes
//   v3: add usage, provider_metrics, usage_hourly tables for v2.2 observability
const currentSchemaVersion = 3

// migrations is an ordered list of schema upgrade functions.
// migrations[0] upgrades v1 → v2, migrations[1] upgrades v2 → v3, etc.
var migrations = []func(tx *sql.Tx) error{
	migrateV1ToV2,
	migrateV2ToV3,
}

// LogDB provides SQLite-backed log storage with batched writes.
type LogDB struct {
	db      *sql.DB
	writeCh chan LogEntry
	done    chan struct{}
}

// OpenLogDB opens (or creates) the SQLite log database in logDir.
func OpenLogDB(logDir string) (*LogDB, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	dbPath := filepath.Join(logDir, "logs.db")
	db, err := openAndMigrate(dbPath)
	if err != nil {
		// Database is corrupt or has incompatible schema — rebuild.
		db, err = rebuildDatabase(dbPath)
		if err != nil {
			return nil, fmt.Errorf("rebuild log database: %w", err)
		}
	}

	ldb := &LogDB{
		db:      db,
		writeCh: make(chan LogEntry, 256),
		done:    make(chan struct{}),
	}
	go ldb.flushLoop()
	return ldb, nil
}

// openAndMigrate opens the database, configures it, and runs schema migration.
func openAndMigrate(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	// Set restrictive permissions - ignore error as file may not exist yet
	_ = os.Chmod(dbPath, 0600)

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, err
	}
	if err := migrateSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// rebuildDatabase removes the corrupt database and creates a fresh one.
func rebuildDatabase(dbPath string) (*sql.DB, error) {
	// Remove the main file and WAL/SHM sidecars - ignore errors
	for _, suffix := range []string{"", "-wal", "-shm"} {
		_ = os.Remove(dbPath + suffix)
	}
	return openAndMigrate(dbPath)
}

// migrateSchema ensures the database is at the latest schema version.
// New databases get the full schema at currentSchemaVersion.
// Existing databases are upgraded step by step.
func migrateSchema(db *sql.DB) error {
	// Ensure schema_version table exists
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)`); err != nil {
		return fmt.Errorf("create schema_version table: %w", err)
	}

	version := getSchemaVersion(db)

	if version == 0 {
		// Fresh database or pre-versioning database.
		// Check if logs table already exists (pre-versioning v1 DB).
		if tableExists(db, "logs") {
			// Existing v1 database — set version to 1 and run migrations.
			version = 1
		} else {
			// Brand new database — create full schema at latest version.
			return initFreshSchema(db)
		}
	}

	if version == currentSchemaVersion {
		return nil
	}
	if version > currentSchemaVersion {
		return fmt.Errorf("database schema version %d is newer than supported version %d", version, currentSchemaVersion)
	}

	// Run migrations sequentially: version N → N+1 uses migrations[N-1]
	for v := version; v < currentSchemaVersion; v++ {
		idx := v - 1 // migrations[0] = v1→v2
		if idx < 0 || idx >= len(migrations) {
			return fmt.Errorf("no migration defined for v%d → v%d", v, v+1)
		}
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration v%d→v%d: %w", v, v+1, err)
		}
		if err := migrations[idx](tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration v%d→v%d: %w", v, v+1, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration v%d→v%d: %w", v, v+1, err)
		}
	}

	setSchemaVersion(db, currentSchemaVersion)
	return nil
}

// initFreshSchema creates the full schema for a brand new database.
func initFreshSchema(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp     DATETIME NOT NULL,
			level         TEXT NOT NULL,
			provider      TEXT DEFAULT '',
			message       TEXT DEFAULT '',
			status_code   INTEGER DEFAULT 0,
			method        TEXT DEFAULT '',
			path          TEXT DEFAULT '',
			error         TEXT DEFAULT '',
			response_body TEXT DEFAULT '',
			session_id    TEXT DEFAULT '',
			client_type   TEXT DEFAULT ''
		)
	`); err != nil {
		return fmt.Errorf("create logs table: %w", err)
	}

	// Create usage table for tracking API usage and costs
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS usage (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp     DATETIME NOT NULL,
			session_id    TEXT NOT NULL,
			provider      TEXT NOT NULL,
			model         TEXT NOT NULL,
			input_tokens  INTEGER NOT NULL,
			output_tokens INTEGER NOT NULL,
			cost_usd      REAL NOT NULL,
			latency_ms    INTEGER DEFAULT 0,
			project_path  TEXT DEFAULT '',
			client_type   TEXT DEFAULT ''
		)
	`); err != nil {
		return fmt.Errorf("create usage table: %w", err)
	}

	// Create provider_metrics table for health monitoring
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS provider_metrics (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp     DATETIME NOT NULL,
			provider      TEXT NOT NULL,
			latency_ms    INTEGER NOT NULL,
			status_code   INTEGER NOT NULL,
			is_error      INTEGER DEFAULT 0,
			is_rate_limit INTEGER DEFAULT 0
		)
	`); err != nil {
		return fmt.Errorf("create provider_metrics table: %w", err)
	}

	// Create usage_hourly table for aggregated dashboard data
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS usage_hourly (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			hour          DATETIME NOT NULL,
			provider      TEXT NOT NULL,
			model         TEXT NOT NULL,
			project_path  TEXT DEFAULT '',
			total_input   INTEGER NOT NULL,
			total_output  INTEGER NOT NULL,
			total_cost    REAL NOT NULL,
			request_count INTEGER NOT NULL,
			UNIQUE(hour, provider, model, project_path)
		)
	`); err != nil {
		return fmt.Errorf("create usage_hourly table: %w", err)
	}

	for _, idx := range []string{
		"CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_logs_provider ON logs(provider)",
		"CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level)",
		"CREATE INDEX IF NOT EXISTS idx_logs_session_id ON logs(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_logs_client_type ON logs(client_type)",
		"CREATE INDEX IF NOT EXISTS idx_usage_timestamp ON usage(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_usage_session_id ON usage(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_usage_provider ON usage(provider)",
		"CREATE INDEX IF NOT EXISTS idx_usage_project_path ON usage(project_path)",
		"CREATE INDEX IF NOT EXISTS idx_provider_metrics_timestamp ON provider_metrics(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_provider_metrics_provider ON provider_metrics(provider)",
		"CREATE INDEX IF NOT EXISTS idx_usage_hourly_hour ON usage_hourly(hour)",
	} {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}

	setSchemaVersion(db, currentSchemaVersion)
	return nil
}

// --- Migrations ---

// migrateV1ToV2 adds session_id, client_type columns and their indexes.
func migrateV1ToV2(tx *sql.Tx) error {
	for _, stmt := range []string{
		"ALTER TABLE logs ADD COLUMN session_id TEXT DEFAULT ''",
		"ALTER TABLE logs ADD COLUMN client_type TEXT DEFAULT ''",
		"CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_logs_provider ON logs(provider)",
		"CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level)",
		"CREATE INDEX IF NOT EXISTS idx_logs_session_id ON logs(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_logs_client_type ON logs(client_type)",
	} {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// migrateV2ToV3 adds usage, provider_metrics, and usage_hourly tables.
func migrateV2ToV3(tx *sql.Tx) error {
	// Create usage table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS usage (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp     DATETIME NOT NULL,
			session_id    TEXT NOT NULL,
			provider      TEXT NOT NULL,
			model         TEXT NOT NULL,
			input_tokens  INTEGER NOT NULL,
			output_tokens INTEGER NOT NULL,
			cost_usd      REAL NOT NULL,
			latency_ms    INTEGER DEFAULT 0,
			project_path  TEXT DEFAULT '',
			client_type   TEXT DEFAULT ''
		)
	`); err != nil {
		return err
	}

	// Create provider_metrics table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS provider_metrics (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp     DATETIME NOT NULL,
			provider      TEXT NOT NULL,
			latency_ms    INTEGER NOT NULL,
			status_code   INTEGER NOT NULL,
			is_error      INTEGER DEFAULT 0,
			is_rate_limit INTEGER DEFAULT 0
		)
	`); err != nil {
		return err
	}

	// Create usage_hourly table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS usage_hourly (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			hour          DATETIME NOT NULL,
			provider      TEXT NOT NULL,
			model         TEXT NOT NULL,
			project_path  TEXT DEFAULT '',
			total_input   INTEGER NOT NULL,
			total_output  INTEGER NOT NULL,
			total_cost    REAL NOT NULL,
			request_count INTEGER NOT NULL,
			UNIQUE(hour, provider, model, project_path)
		)
	`); err != nil {
		return err
	}

	// Create indexes
	for _, idx := range []string{
		"CREATE INDEX IF NOT EXISTS idx_usage_timestamp ON usage(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_usage_session_id ON usage(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_usage_provider ON usage(provider)",
		"CREATE INDEX IF NOT EXISTS idx_usage_project_path ON usage(project_path)",
		"CREATE INDEX IF NOT EXISTS idx_provider_metrics_timestamp ON provider_metrics(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_provider_metrics_provider ON provider_metrics(provider)",
		"CREATE INDEX IF NOT EXISTS idx_usage_hourly_hour ON usage_hourly(hour)",
	} {
		if _, err := tx.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}

// --- Schema version helpers ---

func getSchemaVersion(db *sql.DB) int {
	var version int
	if err := db.QueryRow("SELECT version FROM schema_version LIMIT 1").Scan(&version); err != nil {
		return 0
	}
	return version
}

func setSchemaVersion(db *sql.DB, version int) {
	// Best-effort schema version update - errors are not critical
	_, _ = db.Exec("DELETE FROM schema_version")
	_, _ = db.Exec("INSERT INTO schema_version (version) VALUES (?)", version)
}

func tableExists(db *sql.DB, name string) bool {
	var n int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&n)
	return err == nil && n > 0
}

// Insert queues a log entry for batched writing.
func (ldb *LogDB) Insert(entry LogEntry) {
	select {
	case ldb.writeCh <- entry:
	default:
		// Channel full — drop entry to avoid blocking the caller.
	}
}

// flushLoop collects entries and flushes them periodically or when the batch is large enough.
func (ldb *LogDB) flushLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var batch []LogEntry
	for {
		select {
		case entry, ok := <-ldb.writeCh:
			if !ok {
				ldb.flushBatch(batch)
				close(ldb.done)
				return
			}
			batch = append(batch, entry)
			if len(batch) >= 50 {
				ldb.flushBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				ldb.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch inserts a slice of entries in a single transaction.
func (ldb *LogDB) flushBatch(batch []LogEntry) {
	if len(batch) == 0 {
		return
	}

	tx, err := ldb.db.Begin()
	if err != nil {
		return
	}

	stmt, err := tx.Prepare(`
		INSERT INTO logs (timestamp, level, provider, message, status_code, method, path, error, response_body, session_id, client_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return
	}
	defer stmt.Close()

	var execErr error
	for _, e := range batch {
		_, err := stmt.Exec(
			e.Timestamp.UTC().Format(time.RFC3339Nano),
			string(e.Level),
			e.Provider,
			e.Message,
			e.StatusCode,
			e.Method,
			e.Path,
			e.Error,
			e.ResponseBody,
			e.SessionID,
			e.ClientType,
		)
		if err != nil {
			execErr = err
		}
	}

	if execErr != nil {
		_ = tx.Rollback()
		return
	}
	_ = tx.Commit()
}

// Query returns log entries matching the filter, newest first.
func (ldb *LogDB) Query(filter LogFilter) ([]LogEntry, error) {
	var conditions []string
	var args []interface{}

	if filter.Provider != "" {
		conditions = append(conditions, "provider = ?")
		args = append(args, filter.Provider)
	}
	if filter.Level != "" {
		conditions = append(conditions, "level = ?")
		args = append(args, string(filter.Level))
	}
	if filter.ErrorsOnly {
		conditions = append(conditions, "level IN ('error', 'warn')")
	}

	if filter.StatusCode > 0 {
		conditions = append(conditions, "status_code = ?")
		args = append(args, filter.StatusCode)
	}
	if filter.StatusMin > 0 {
		conditions = append(conditions, "status_code >= ?")
		args = append(args, filter.StatusMin)
	}
	if filter.StatusMax > 0 {
		conditions = append(conditions, "status_code <= ?")
		args = append(args, filter.StatusMax)
	}
	if filter.SessionID != "" {
		conditions = append(conditions, "session_id = ?")
		args = append(args, filter.SessionID)
	}
	if filter.ClientType != "" {
		conditions = append(conditions, "client_type = ?")
		args = append(args, filter.ClientType)
	}

	query := "SELECT timestamp, level, provider, message, status_code, method, path, error, response_body, session_id, client_type FROM logs"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY timestamp DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	query += " LIMIT ?"
	args = append(args, limit)

	rows, err := ldb.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query logs: %w", err)
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var e LogEntry
		var tsStr string
		var level string
		if err := rows.Scan(&tsStr, &level, &e.Provider, &e.Message, &e.StatusCode, &e.Method, &e.Path, &e.Error, &e.ResponseBody, &e.SessionID, &e.ClientType); err != nil {
			continue
		}
		e.Level = LogLevel(level)
		if t, err := time.Parse(time.RFC3339Nano, tsStr); err == nil {
			e.Timestamp = t
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// GetProviders returns distinct provider names from the log database.
func (ldb *LogDB) GetProviders() ([]string, error) {
	rows, err := ldb.db.Query("SELECT DISTINCT provider FROM logs WHERE provider != '' ORDER BY provider")
	if err != nil {
		return nil, fmt.Errorf("query providers: %w", err)
	}
	defer rows.Close()

	var providers []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			continue
		}
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

// Close stops the background writer and closes the database.
func (ldb *LogDB) Close() error {
	close(ldb.writeCh)
	<-ldb.done
	return ldb.db.Close()
}
