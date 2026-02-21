package proxy

import (
	"time"
)

// ProviderMetric represents a single provider health metric.
type ProviderMetric struct {
	Timestamp   time.Time
	Provider    string
	LatencyMs   int
	StatusCode  int
	IsError     bool
	IsRateLimit bool
}

// ProviderMetrics holds aggregated metrics for a provider.
type ProviderMetrics struct {
	Provider        string  `json:"provider"`
	TotalRequests   int     `json:"total_requests"`
	SuccessCount    int     `json:"success_count"`
	ErrorCount      int     `json:"error_count"`
	RateLimitCount  int     `json:"rate_limit_count"`
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
	MinLatencyMs    int     `json:"min_latency_ms"`
	MaxLatencyMs    int     `json:"max_latency_ms"`
	SuccessRate     float64 `json:"success_rate"`
	LastSuccess     *time.Time `json:"last_success,omitempty"`
	LastError       *time.Time `json:"last_error,omitempty"`
}

// RecordMetric stores a provider metric in the database.
func (ldb *LogDB) RecordMetric(provider string, latencyMs int, statusCode int, isError, isRateLimit bool) error {
	if ldb == nil || ldb.db == nil {
		return nil
	}

	_, err := ldb.db.Exec(`
		INSERT INTO provider_metrics (timestamp, provider, latency_ms, status_code, is_error, is_rate_limit)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		time.Now().UTC().Format(time.RFC3339Nano),
		provider,
		latencyMs,
		statusCode,
		boolToInt(isError),
		boolToInt(isRateLimit),
	)
	return err
}

// GetProviderMetrics returns aggregated metrics for a provider since the given time.
func (ldb *LogDB) GetProviderMetrics(provider string, since time.Time) (*ProviderMetrics, error) {
	if ldb == nil || ldb.db == nil {
		return &ProviderMetrics{Provider: provider}, nil
	}

	metrics := &ProviderMetrics{Provider: provider}

	// Get aggregated stats
	err := ldb.db.QueryRow(`
		SELECT
			COUNT(*),
			SUM(CASE WHEN is_error = 0 THEN 1 ELSE 0 END),
			SUM(CASE WHEN is_error = 1 THEN 1 ELSE 0 END),
			SUM(CASE WHEN is_rate_limit = 1 THEN 1 ELSE 0 END),
			COALESCE(AVG(latency_ms), 0),
			COALESCE(MIN(latency_ms), 0),
			COALESCE(MAX(latency_ms), 0)
		FROM provider_metrics
		WHERE provider = ? AND timestamp >= ?
	`, provider, since.Format(time.RFC3339Nano)).Scan(
		&metrics.TotalRequests,
		&metrics.SuccessCount,
		&metrics.ErrorCount,
		&metrics.RateLimitCount,
		&metrics.AvgLatencyMs,
		&metrics.MinLatencyMs,
		&metrics.MaxLatencyMs,
	)
	if err != nil {
		return nil, err
	}

	if metrics.TotalRequests > 0 {
		metrics.SuccessRate = float64(metrics.SuccessCount) / float64(metrics.TotalRequests) * 100
	}

	// Get last success time
	var lastSuccessStr string
	err = ldb.db.QueryRow(`
		SELECT timestamp FROM provider_metrics
		WHERE provider = ? AND is_error = 0
		ORDER BY timestamp DESC LIMIT 1
	`, provider).Scan(&lastSuccessStr)
	if err == nil {
		if t, err := time.Parse(time.RFC3339Nano, lastSuccessStr); err == nil {
			metrics.LastSuccess = &t
		}
	}

	// Get last error time
	var lastErrorStr string
	err = ldb.db.QueryRow(`
		SELECT timestamp FROM provider_metrics
		WHERE provider = ? AND is_error = 1
		ORDER BY timestamp DESC LIMIT 1
	`, provider).Scan(&lastErrorStr)
	if err == nil {
		if t, err := time.Parse(time.RFC3339Nano, lastErrorStr); err == nil {
			metrics.LastError = &t
		}
	}

	return metrics, nil
}

// GetAllProviderMetrics returns metrics for all providers since the given time.
func (ldb *LogDB) GetAllProviderMetrics(since time.Time) (map[string]*ProviderMetrics, error) {
	if ldb == nil || ldb.db == nil {
		return make(map[string]*ProviderMetrics), nil
	}

	result := make(map[string]*ProviderMetrics)

	rows, err := ldb.db.Query(`
		SELECT
			provider,
			COUNT(*),
			SUM(CASE WHEN is_error = 0 THEN 1 ELSE 0 END),
			SUM(CASE WHEN is_error = 1 THEN 1 ELSE 0 END),
			SUM(CASE WHEN is_rate_limit = 1 THEN 1 ELSE 0 END),
			COALESCE(AVG(latency_ms), 0),
			COALESCE(MIN(latency_ms), 0),
			COALESCE(MAX(latency_ms), 0)
		FROM provider_metrics
		WHERE timestamp >= ?
		GROUP BY provider
	`, since.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m ProviderMetrics
		if err := rows.Scan(
			&m.Provider,
			&m.TotalRequests,
			&m.SuccessCount,
			&m.ErrorCount,
			&m.RateLimitCount,
			&m.AvgLatencyMs,
			&m.MinLatencyMs,
			&m.MaxLatencyMs,
		); err != nil {
			continue
		}
		if m.TotalRequests > 0 {
			m.SuccessRate = float64(m.SuccessCount) / float64(m.TotalRequests) * 100
		}
		result[m.Provider] = &m
	}

	return result, nil
}

// GetLatencyHistory returns latency data points for a provider (for charts).
func (ldb *LogDB) GetLatencyHistory(provider string, since time.Time, bucketMinutes int) ([]LatencyPoint, error) {
	if ldb == nil || ldb.db == nil {
		return nil, nil
	}

	if bucketMinutes <= 0 {
		bucketMinutes = 5
	}

	// SQLite doesn't have great time bucketing, so we'll do it in Go
	rows, err := ldb.db.Query(`
		SELECT timestamp, latency_ms, is_error
		FROM provider_metrics
		WHERE provider = ? AND timestamp >= ?
		ORDER BY timestamp ASC
	`, provider, since.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Bucket the data
	buckets := make(map[time.Time]*LatencyPoint)
	bucketDuration := time.Duration(bucketMinutes) * time.Minute

	for rows.Next() {
		var tsStr string
		var latency int
		var isError int
		if err := rows.Scan(&tsStr, &latency, &isError); err != nil {
			continue
		}

		ts, err := time.Parse(time.RFC3339Nano, tsStr)
		if err != nil {
			continue
		}

		// Round to bucket
		bucketTime := ts.Truncate(bucketDuration)

		if _, ok := buckets[bucketTime]; !ok {
			buckets[bucketTime] = &LatencyPoint{
				Timestamp: bucketTime,
			}
		}

		bp := buckets[bucketTime]
		bp.Count++
		bp.TotalLatency += latency
		if latency < bp.MinLatency || bp.MinLatency == 0 {
			bp.MinLatency = latency
		}
		if latency > bp.MaxLatency {
			bp.MaxLatency = latency
		}
		if isError == 1 {
			bp.ErrorCount++
		}
	}

	// Convert to slice and calculate averages
	var result []LatencyPoint
	for _, bp := range buckets {
		if bp.Count > 0 {
			bp.AvgLatency = float64(bp.TotalLatency) / float64(bp.Count)
		}
		result = append(result, *bp)
	}

	// Sort by timestamp
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Timestamp.After(result[j].Timestamp) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

// LatencyPoint represents a single data point for latency charts.
type LatencyPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	AvgLatency   float64   `json:"avg_latency"`
	MinLatency   int       `json:"min_latency"`
	MaxLatency   int       `json:"max_latency"`
	Count        int       `json:"count"`
	ErrorCount   int       `json:"error_count"`
	TotalLatency int       `json:"-"`
}

// CleanupOldMetrics removes metrics older than the specified duration.
func (ldb *LogDB) CleanupOldMetrics(maxAge time.Duration) (int64, error) {
	if ldb == nil || ldb.db == nil {
		return 0, nil
	}

	cutoff := time.Now().UTC().Add(-maxAge)

	result, err := ldb.db.Exec(`
		DELETE FROM provider_metrics WHERE timestamp < ?
	`, cutoff.Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
