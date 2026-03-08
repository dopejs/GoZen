package httpx

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
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
				if logger != nil {
					logger.Printf("[%s] recovered panic: %v method=%s path=%s\n%s", component, rec, r.Method, r.URL.Path, debug.Stack())
				}

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
