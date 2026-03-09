package daemon

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestStructuredLoggerJSONFormat verifies that logs are output in JSON format
func TestStructuredLoggerJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(&buf)

	logger.Info("test_event", map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})

	output := buf.String()
	if output == "" {
		t.Fatal("expected log output, got empty string")
	}

	// Verify it's valid JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("log output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Verify required fields exist
	requiredFields := []string{"timestamp", "level", "event"}
	for _, field := range requiredFields {
		if _, ok := logEntry[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}
}

// TestStructuredLoggerEventFields verifies that all event fields are logged correctly
func TestStructuredLoggerEventFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(&buf)

	testFields := map[string]interface{}{
		"session_id": "test-session",
		"provider":   "test-provider",
		"duration":   123,
		"error":      "test error",
	}

	logger.Error("test_error_event", testFields)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log JSON: %v", err)
	}

	// Verify timestamp format (ISO 8601)
	timestamp, ok := logEntry["timestamp"].(string)
	if !ok {
		t.Error("timestamp field is not a string")
	}
	if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
		t.Errorf("timestamp is not in RFC3339 format: %s", timestamp)
	}

	// Verify level
	if level := logEntry["level"]; level != "error" {
		t.Errorf("level = %v, want 'error'", level)
	}

	// Verify event
	if event := logEntry["event"]; event != "test_error_event" {
		t.Errorf("event = %v, want 'test_error_event'", event)
	}

	// Verify custom fields
	if sessionID := logEntry["session_id"]; sessionID != "test-session" {
		t.Errorf("session_id = %v, want 'test-session'", sessionID)
	}
	if provider := logEntry["provider"]; provider != "test-provider" {
		t.Errorf("provider = %v, want 'test-provider'", provider)
	}
	if duration := logEntry["duration"]; duration != float64(123) {
		t.Errorf("duration = %v, want 123", duration)
	}
	if errMsg := logEntry["error"]; errMsg != "test error" {
		t.Errorf("error = %v, want 'test error'", errMsg)
	}
}

// TestStructuredLoggerLevels verifies that different log levels work correctly
func TestStructuredLoggerLevels(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(*StructuredLogger, string, map[string]interface{})
		wantLevel string
	}{
		{
			name: "info level",
			logFunc: func(l *StructuredLogger, event string, fields map[string]interface{}) {
				l.Info(event, fields)
			},
			wantLevel: "info",
		},
		{
			name: "warn level",
			logFunc: func(l *StructuredLogger, event string, fields map[string]interface{}) {
				l.Warn(event, fields)
			},
			wantLevel: "warn",
		},
		{
			name: "error level",
			logFunc: func(l *StructuredLogger, event string, fields map[string]interface{}) {
				l.Error(event, fields)
			},
			wantLevel: "error",
		},
		{
			name: "debug level",
			logFunc: func(l *StructuredLogger, event string, fields map[string]interface{}) {
				l.Debug(event, fields)
			},
			wantLevel: "debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewStructuredLogger(&buf)

			tt.logFunc(logger, "test_event", map[string]interface{}{"key": "value"})

			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Fatalf("failed to parse log JSON: %v", err)
			}

			if level := logEntry["level"]; level != tt.wantLevel {
				t.Errorf("level = %v, want %v", level, tt.wantLevel)
			}
		})
	}
}

// TestStructuredLoggerSelectiveLogging verifies that only errors and slow requests are logged
// This test documents the expected behavior - actual filtering happens at call sites
func TestStructuredLoggerSelectiveLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(&buf)

	// Simulate fast successful request (should NOT be logged in production)
	// This test just verifies the logger CAN log it if called
	logger.Info("request_completed", map[string]interface{}{
		"duration_ms": 50,
		"status":      200,
	})

	output := buf.String()
	if output == "" {
		t.Fatal("expected log output")
	}

	// Verify it was logged (filtering happens at call site, not in logger)
	if !strings.Contains(output, "request_completed") {
		t.Error("expected request_completed event in log")
	}

	buf.Reset()

	// Simulate slow request (>1s) - SHOULD be logged
	logger.Warn("request_slow", map[string]interface{}{
		"duration_ms": 1500,
		"status":      200,
	})

	output = buf.String()
	if !strings.Contains(output, "request_slow") {
		t.Error("expected request_slow event in log")
	}

	buf.Reset()

	// Simulate error request - SHOULD be logged
	logger.Error("request_failed", map[string]interface{}{
		"duration_ms": 100,
		"status":      500,
		"error":       "internal server error",
	})

	output = buf.String()
	if !strings.Contains(output, "request_failed") {
		t.Error("expected request_failed event in log")
	}
}

// TestStructuredLoggerNilFields verifies that nil fields don't cause panics
func TestStructuredLoggerNilFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(&buf)

	// Should not panic with nil fields
	logger.Info("test_event", nil)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log JSON: %v", err)
	}

	// Should still have required fields
	if _, ok := logEntry["timestamp"]; !ok {
		t.Error("missing timestamp field")
	}
	if _, ok := logEntry["level"]; !ok {
		t.Error("missing level field")
	}
	if _, ok := logEntry["event"]; !ok {
		t.Error("missing event field")
	}
}

// TestStructuredLoggerEmptyEvent verifies behavior with empty event name
func TestStructuredLoggerEmptyEvent(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(&buf)

	logger.Info("", map[string]interface{}{"key": "value"})

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log JSON: %v", err)
	}

	// Event field should exist (even if empty)
	if _, ok := logEntry["event"]; !ok {
		t.Error("missing event field")
	}
}
