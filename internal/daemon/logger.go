package daemon

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

// StructuredLogger provides JSON-formatted logging for daemon events
type StructuredLogger struct {
	mu     sync.Mutex
	writer io.Writer
}

// NewStructuredLogger creates a new structured logger that writes to the given writer
func NewStructuredLogger(w io.Writer) *StructuredLogger {
	return &StructuredLogger{
		writer: w,
	}
}

// logEntry represents a single log entry
type logEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Event     string                 `json:"event"`
	Fields    map[string]interface{} `json:",inline"`
}

// log writes a log entry with the given level, event, and fields
func (l *StructuredLogger) log(level, event string, fields map[string]interface{}) {
	entry := logEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Event:     event,
		Fields:    fields,
	}

	// Merge fields into the entry for inline JSON output
	data := make(map[string]interface{})
	data["timestamp"] = entry.Timestamp
	data["level"] = entry.Level
	data["event"] = entry.Event

	// Add custom fields
	if fields != nil {
		for k, v := range fields {
			data[k] = v
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Marshal to JSON and write
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Fallback: write error message
		l.writer.Write([]byte(`{"timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `","level":"error","event":"log_marshal_error","error":"` + err.Error() + `"}` + "\n"))
		return
	}

	l.writer.Write(jsonData)
	l.writer.Write([]byte("\n"))
}

// Info logs an informational event
func (l *StructuredLogger) Info(event string, fields map[string]interface{}) {
	l.log("info", event, fields)
}

// Warn logs a warning event
func (l *StructuredLogger) Warn(event string, fields map[string]interface{}) {
	l.log("warn", event, fields)
}

// Error logs an error event
func (l *StructuredLogger) Error(event string, fields map[string]interface{}) {
	l.log("error", event, fields)
}

// Debug logs a debug event
func (l *StructuredLogger) Debug(event string, fields map[string]interface{}) {
	l.log("debug", event, fields)
}

// --- Routing-specific logging functions ---

// LogRoutingDecision logs a routing decision with scenario, source, and reason
func (l *StructuredLogger) LogRoutingDecision(sessionID, scenario, source, reason string, confidence float64, provider string) {
	l.Info("routing_decision", map[string]interface{}{
		"session_id": sessionID,
		"scenario":   scenario,
		"source":     source,
		"reason":     reason,
		"confidence": confidence,
		"provider":   provider,
	})
}

// LogRoutingFallback logs when routing falls back to default behavior
func (l *StructuredLogger) LogRoutingFallback(sessionID, scenario, reason, fallbackProvider string) {
	l.Warn("routing_fallback", map[string]interface{}{
		"session_id":        sessionID,
		"scenario":          scenario,
		"reason":            reason,
		"fallback_provider": fallbackProvider,
	})
}

// LogProtocolDetection logs the detected API protocol for a request
func (l *StructuredLogger) LogProtocolDetection(sessionID, detectedProtocol, detectionMethod string) {
	l.Debug("protocol_detection", map[string]interface{}{
		"session_id":        sessionID,
		"detected_protocol": detectedProtocol,
		"detection_method":  detectionMethod,
	})
}

// LogRequestFeatures logs extracted request features for routing classification
func (l *StructuredLogger) LogRequestFeatures(sessionID string, features map[string]interface{}) {
	fields := map[string]interface{}{
		"session_id": sessionID,
	}
	for k, v := range features {
		fields[k] = v
	}
	l.Debug("request_features", fields)
}

// LogProviderSelection logs the final provider selection with strategy details
func (l *StructuredLogger) LogProviderSelection(sessionID, provider, strategy, reason string, candidates []string) {
	l.Info("provider_selection", map[string]interface{}{
		"session_id": sessionID,
		"provider":   provider,
		"strategy":   strategy,
		"reason":     reason,
		"candidates": candidates,
	})
}
