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

// NewRequestMonitor creates a new RequestMonitor with the specified buffer size.
func NewRequestMonitor(maxRecords int) *RequestMonitor {
	return &RequestMonitor{
		records:    make([]RequestRecord, 0, maxRecords),
		maxRecords: maxRecords,
	}
}

// Add appends a new request record to the buffer.
// If the buffer is full, it evicts the oldest 20% of records (LRU eviction).
func (rm *RequestMonitor) Add(record RequestRecord) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// If buffer is full, evict oldest 20%
	if len(rm.records) >= rm.maxRecords {
		evictCount := rm.maxRecords / 5
		if evictCount < 1 {
			evictCount = 1
		}
		rm.records = rm.records[evictCount:]
	}

	rm.records = append(rm.records, record)
}

// GetRecent returns the most recent request records matching the filter criteria,
// in reverse chronological order (newest first).
func (rm *RequestMonitor) GetRecent(limit int, filter RequestFilter) []RequestRecord {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Filter records
	var filtered []RequestRecord
	for i := len(rm.records) - 1; i >= 0; i-- {
		record := rm.records[i]

		// Apply filters
		if filter.Provider != "" && record.Provider != filter.Provider {
			continue
		}
		if filter.SessionID != "" && record.SessionID != filter.SessionID {
			continue
		}
		if filter.Model != "" && record.Model != filter.Model {
			continue
		}
		if filter.MinStatus > 0 && record.StatusCode < filter.MinStatus {
			continue
		}
		if filter.MaxStatus > 0 && record.StatusCode > filter.MaxStatus {
			continue
		}
		if !filter.StartTime.IsZero() && record.Timestamp.Before(filter.StartTime) {
			continue
		}
		if !filter.EndTime.IsZero() && record.Timestamp.After(filter.EndTime) {
			continue
		}

		filtered = append(filtered, record)

		// Apply limit
		if limit > 0 && len(filtered) >= limit {
			break
		}
	}

	return filtered
}

// GetByID returns a single request record by ID, or nil if not found.
func (rm *RequestMonitor) GetByID(id string) *RequestRecord {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for i := len(rm.records) - 1; i >= 0; i-- {
		if rm.records[i].ID == id {
			record := rm.records[i]
			return &record
		}
	}

	return nil
}

// Global request monitor singleton
var (
	globalMonitor     *RequestMonitor
	globalMonitorOnce sync.Once
)

// GetGlobalRequestMonitor returns the global request monitor singleton.
func GetGlobalRequestMonitor() *RequestMonitor {
	globalMonitorOnce.Do(func() {
		globalMonitor = NewRequestMonitor(1000) // Default buffer size
	})
	return globalMonitor
}

// InitGlobalRequestMonitor initializes the global monitor with a custom buffer size.
// This should be called before any requests are processed.
func InitGlobalRequestMonitor(maxRecords int) {
	globalMonitorOnce.Do(func() {
		globalMonitor = NewRequestMonitor(maxRecords)
	})
}
