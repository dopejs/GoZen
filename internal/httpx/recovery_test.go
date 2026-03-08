package httpx

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverReturnsInternalServerErrorOnPanic(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	handler := Recover(logger, "test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}
	if body := rec.Body.String(); body == "" {
		t.Fatal("expected error body")
	}
}

// TestRecoverDoesNotCrashDaemon verifies panic recovery prevents daemon crash
func TestRecoverDoesNotCrashDaemon(t *testing.T) {
	logger := log.New(io.Discard, "", 0)

	// Create handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("simulated crash")
	})

	// Wrap with recovery middleware
	handler := Recover(logger, "daemon", panicHandler)

	// First request triggers panic
	req1 := httptest.NewRequest(http.MethodGet, "/crash", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusInternalServerError {
		t.Fatalf("first request: status = %d, want %d", rec1.Code, http.StatusInternalServerError)
	}

	// Second request should still work (daemon didn't crash)
	req2 := httptest.NewRequest(http.MethodGet, "/crash", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("second request: status = %d, want %d", rec2.Code, http.StatusInternalServerError)
	}

	// If we got here, daemon survived multiple panics
}
