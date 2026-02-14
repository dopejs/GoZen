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
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open log database: %w", err)
	}

	// Configure for concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	// Create table and indexes
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
			response_body TEXT DEFAULT ''
		)
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("create logs table: %w", err)
	}

	for _, idx := range []string{
		"CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_logs_provider ON logs(provider)",
		"CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level)",
	} {
		if _, err := db.Exec(idx); err != nil {
			db.Close()
			return nil, fmt.Errorf("create index: %w", err)
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
				// Channel closed — flush remaining and signal done.
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
		INSERT INTO logs (timestamp, level, provider, message, status_code, method, path, error, response_body)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return
	}
	defer stmt.Close()

	for _, e := range batch {
		stmt.Exec(
			e.Timestamp.UTC().Format(time.RFC3339Nano),
			string(e.Level),
			e.Provider,
			e.Message,
			e.StatusCode,
			e.Method,
			e.Path,
			e.Error,
			e.ResponseBody,
		)
	}

	tx.Commit()
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

	query := "SELECT timestamp, level, provider, message, status_code, method, path, error, response_body FROM logs"
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
		if err := rows.Scan(&tsStr, &level, &e.Provider, &e.Message, &e.StatusCode, &e.Method, &e.Path, &e.Error, &e.ResponseBody); err != nil {
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
