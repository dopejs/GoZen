package httpx

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"sync"
)

type statusWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	w.wroteHeader = true
	return w.ResponseWriter.Write(p)
}

// Recover wraps an HTTP handler and prevents panics from crashing the process.
func Recover(logger *log.Logger, component string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w}

		defer func() {
			if rec := recover(); rec != nil {
				stack := string(debug.Stack())
				if logger != nil {
					logger.Printf("[%s] recovered panic: %v method=%s path=%s\n%s", component, rec, r.Method, r.URL.Path, stack)
				}

				// Log panic_recovered event (T069)
				logPanicRecovered(rec, stack, r.URL.Path)

				if sw.wroteHeader {
					return
				}

				sw.Header().Set("Content-Type", "application/json")
				sw.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(sw).Encode(map[string]string{
					"error": "internal server error",
				})
			}
		}()

		next.ServeHTTP(sw, r)
	})
}

// logPanicRecovered logs panic_recovered event (T069)
func logPanicRecovered(panicValue interface{}, stack, path string) {
	// Get daemon structured logger if available
	daemonLogger := getDaemonLogger()
	if daemonLogger == nil {
		return
	}

	// Truncate stack trace to reasonable length
	stackStr := stack
	if len(stackStr) > 2000 {
		stackStr = stackStr[:2000] + "..."
	}

	daemonLogger.Error("panic_recovered", map[string]interface{}{
		"error": fmt.Sprintf("%v", panicValue),
		"stack": stackStr,
		"path":  path,
	})
}

// daemonStructuredLogger holds the daemon's structured logger
var (
	daemonStructuredLogger     *daemonLogger
	daemonStructuredLoggerOnce sync.Once
	daemonStructuredLoggerMu   sync.RWMutex
)

// daemonLogger interface matches daemon.StructuredLogger methods
type daemonLogger interface {
	Error(event string, fields map[string]interface{})
}

// SetDaemonLogger sets the daemon's structured logger for panic logging
func SetDaemonLogger(logger daemonLogger) {
	daemonStructuredLoggerMu.Lock()
	defer daemonStructuredLoggerMu.Unlock()
	daemonStructuredLogger = &logger
}

// getDaemonLogger returns the daemon's structured logger if available
func getDaemonLogger() daemonLogger {
	daemonStructuredLoggerMu.RLock()
	defer daemonStructuredLoggerMu.RUnlock()
	if daemonStructuredLogger != nil {
		return *daemonStructuredLogger
	}
	return nil
}
