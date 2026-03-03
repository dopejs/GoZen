package proxy

import (
	"sync"
	"time"
)

// RequestRecord represents a single proxied API request with complete metadata
// for monitoring and debugging.
type RequestRecord struct {
	ID            string            `json:"id"`
	Timestamp     time.Time         `json:"timestamp"`
	SessionID     string            `json:"session_id"`
	ClientType    string            `json:"client_type"`
	Provider      string            `json:"provider"`
	Model         string            `json:"model"`
	RequestFormat string            `json:"request_format"`
	StatusCode    int               `json:"status_code"`
	Duration      time.Duration     `json:"duration_ms"`
	InputTokens   int               `json:"input_tokens"`
	OutputTokens  int               `json:"output_tokens"`
	Cost          float64           `json:"cost_usd"`
	RequestSize   int               `json:"request_size"`
	FailoverChain []ProviderAttempt `json:"failover_chain,omitempty"`
	ErrorMessage  string            `json:"error_message,omitempty"`
}

// ProviderAttempt represents a single attempt to forward a request to a provider
// (part of failover chain).
type ProviderAttempt struct {
	Provider     string        `json:"provider"`
	StatusCode   int           `json:"status_code"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Duration     time.Duration `json:"duration_ms"`
	Skipped      bool          `json:"skipped"`
	SkipReason   string        `json:"skip_reason,omitempty"`
}

// RequestMonitor is a thread-safe ring buffer for storing recent request records.
type RequestMonitor struct {
	mu         sync.RWMutex
	records    []RequestRecord
	maxRecords int
}

// RequestFilter defines criteria for filtering request records.
type RequestFilter struct {
	Provider  string    // Filter by provider name (empty = all)
	SessionID string    // Filter by session ID (empty = all)
	MinStatus int       // Minimum status code (0 = no filter)
	MaxStatus int       // Maximum status code (0 = no filter)
	StartTime time.Time // Start of time range (zero = no filter)
	EndTime   time.Time // End of time range (zero = no filter)
	Model     string    // Filter by model name (empty = all)
}
