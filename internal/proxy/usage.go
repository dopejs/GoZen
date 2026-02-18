package proxy

import (
	"strings"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// UsageEntry represents a single API usage record.
type UsageEntry struct {
	Timestamp    time.Time
	SessionID    string
	Provider     string
	Model        string
	InputTokens  int
	OutputTokens int
	CostUSD      float64
	LatencyMs    int
	ProjectPath  string
	ClientType   string
}

// UsageSummary provides aggregated usage statistics.
type UsageSummary struct {
	TotalInputTokens  int     `json:"total_input_tokens"`
	TotalOutputTokens int     `json:"total_output_tokens"`
	TotalCost         float64 `json:"total_cost"`
	RequestCount      int     `json:"request_count"`
	ByProvider        map[string]*UsageStats `json:"by_provider,omitempty"`
	ByModel           map[string]*UsageStats `json:"by_model,omitempty"`
	ByProject         map[string]*UsageStats `json:"by_project,omitempty"`
}

// UsageStats holds usage statistics for a single dimension.
type UsageStats struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	Cost         float64 `json:"cost"`
	RequestCount int     `json:"request_count"`
}

// UsageTracker tracks API usage and calculates costs.
type UsageTracker struct {
	db      *LogDB
	pricing map[string]*config.ModelPricing
}

// NewUsageTracker creates a new usage tracker.
func NewUsageTracker(db *LogDB) *UsageTracker {
	return &UsageTracker{
		db:      db,
		pricing: config.GetPricing(),
	}
}

// ReloadPricing refreshes the pricing data from config.
func (t *UsageTracker) ReloadPricing() {
	t.pricing = config.GetPricing()
}

// CalculateCost calculates the cost for a given model and token counts.
func (t *UsageTracker) CalculateCost(model string, inputTokens, outputTokens int) float64 {
	pricing := t.findPricing(model)
	if pricing == nil {
		return 0
	}

	inputCost := float64(inputTokens) / 1_000_000 * pricing.InputPerMillion
	outputCost := float64(outputTokens) / 1_000_000 * pricing.OutputPerMillion
	return inputCost + outputCost
}

// findPricing finds the pricing for a model, supporting partial matches.
func (t *UsageTracker) findPricing(model string) *config.ModelPricing {
	// Exact match first
	if p, ok := t.pricing[model]; ok {
		return p
	}

	// Try partial match (model name contains key)
	modelLower := strings.ToLower(model)
	for key, p := range t.pricing {
		if strings.Contains(modelLower, strings.ToLower(key)) {
			return p
		}
	}

	// Try matching by model family
	if strings.Contains(modelLower, "opus") {
		if p, ok := t.pricing["claude-3-opus-20240229"]; ok {
			return p
		}
	}
	if strings.Contains(modelLower, "sonnet") {
		if p, ok := t.pricing["claude-3-5-sonnet-20241022"]; ok {
			return p
		}
	}
	if strings.Contains(modelLower, "haiku") {
		if p, ok := t.pricing["claude-3-5-haiku-20241022"]; ok {
			return p
		}
	}

	return nil
}

// Record stores a usage entry in the database.
func (t *UsageTracker) Record(entry UsageEntry) error {
	if t.db == nil || t.db.db == nil {
		return nil
	}

	_, err := t.db.db.Exec(`
		INSERT INTO usage (timestamp, session_id, provider, model, input_tokens, output_tokens, cost_usd, latency_ms, project_path, client_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		entry.Timestamp.UTC().Format(time.RFC3339Nano),
		entry.SessionID,
		entry.Provider,
		entry.Model,
		entry.InputTokens,
		entry.OutputTokens,
		entry.CostUSD,
		entry.LatencyMs,
		entry.ProjectPath,
		entry.ClientType,
	)
	return err
}

// GetSummary returns usage summary for a time period.
// period can be "day", "week", "month", or "all".
// projectPath filters by project (empty string for all projects).
func (t *UsageTracker) GetSummary(period string, projectPath string) (*UsageSummary, error) {
	if t.db == nil || t.db.db == nil {
		return &UsageSummary{
			ByProvider: make(map[string]*UsageStats),
			ByModel:    make(map[string]*UsageStats),
			ByProject:  make(map[string]*UsageStats),
		}, nil
	}

	var since time.Time
	now := time.Now().UTC()

	switch period {
	case "day":
		since = now.AddDate(0, 0, -1)
	case "week":
		since = now.AddDate(0, 0, -7)
	case "month":
		since = now.AddDate(0, -1, 0)
	default:
		since = time.Time{} // all time
	}

	return t.querySummary(since, projectPath)
}

// GetDailyCost returns the total cost for today.
func (t *UsageTracker) GetDailyCost(projectPath string) (float64, error) {
	return t.getCostSince(time.Now().UTC().Truncate(24*time.Hour), projectPath)
}

// GetWeeklyCost returns the total cost for the current week.
func (t *UsageTracker) GetWeeklyCost(projectPath string) (float64, error) {
	now := time.Now().UTC()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = weekStart.Truncate(24 * time.Hour)
	return t.getCostSince(weekStart, projectPath)
}

// GetMonthlyCost returns the total cost for the current month.
func (t *UsageTracker) GetMonthlyCost(projectPath string) (float64, error) {
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return t.getCostSince(monthStart, projectPath)
}

func (t *UsageTracker) getCostSince(since time.Time, projectPath string) (float64, error) {
	if t.db == nil || t.db.db == nil {
		return 0, nil
	}

	var query string
	var args []interface{}

	if projectPath != "" {
		query = `SELECT COALESCE(SUM(cost_usd), 0) FROM usage WHERE timestamp >= ? AND project_path = ?`
		args = []interface{}{since.Format(time.RFC3339Nano), projectPath}
	} else {
		query = `SELECT COALESCE(SUM(cost_usd), 0) FROM usage WHERE timestamp >= ?`
		args = []interface{}{since.Format(time.RFC3339Nano)}
	}

	var cost float64
	err := t.db.db.QueryRow(query, args...).Scan(&cost)
	return cost, err
}

func (t *UsageTracker) querySummary(since time.Time, projectPath string) (*UsageSummary, error) {
	summary := &UsageSummary{
		ByProvider: make(map[string]*UsageStats),
		ByModel:    make(map[string]*UsageStats),
		ByProject:  make(map[string]*UsageStats),
	}

	var conditions []string
	var args []interface{}

	if !since.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, since.Format(time.RFC3339Nano))
	}
	if projectPath != "" {
		conditions = append(conditions, "project_path = ?")
		args = append(args, projectPath)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Query totals
	query := `SELECT COALESCE(SUM(input_tokens), 0), COALESCE(SUM(output_tokens), 0), COALESCE(SUM(cost_usd), 0), COUNT(*) FROM usage` + whereClause
	err := t.db.db.QueryRow(query, args...).Scan(&summary.TotalInputTokens, &summary.TotalOutputTokens, &summary.TotalCost, &summary.RequestCount)
	if err != nil {
		return nil, err
	}

	// Query by provider
	query = `SELECT provider, SUM(input_tokens), SUM(output_tokens), SUM(cost_usd), COUNT(*) FROM usage` + whereClause + ` GROUP BY provider`
	rows, err := t.db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var provider string
		var stats UsageStats
		if err := rows.Scan(&provider, &stats.InputTokens, &stats.OutputTokens, &stats.Cost, &stats.RequestCount); err != nil {
			continue
		}
		summary.ByProvider[provider] = &stats
	}

	// Query by model
	query = `SELECT model, SUM(input_tokens), SUM(output_tokens), SUM(cost_usd), COUNT(*) FROM usage` + whereClause + ` GROUP BY model`
	rows, err = t.db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var model string
		var stats UsageStats
		if err := rows.Scan(&model, &stats.InputTokens, &stats.OutputTokens, &stats.Cost, &stats.RequestCount); err != nil {
			continue
		}
		summary.ByModel[model] = &stats
	}

	// Query by project
	query = `SELECT project_path, SUM(input_tokens), SUM(output_tokens), SUM(cost_usd), COUNT(*) FROM usage` + whereClause + ` GROUP BY project_path`
	rows, err = t.db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var project string
		var stats UsageStats
		if err := rows.Scan(&project, &stats.InputTokens, &stats.OutputTokens, &stats.Cost, &stats.RequestCount); err != nil {
			continue
		}
		if project != "" {
			summary.ByProject[project] = &stats
		}
	}

	return summary, nil
}

// AggregateHourly aggregates recent usage data into hourly buckets.
// This should be called periodically (e.g., every hour) to maintain dashboard performance.
func (t *UsageTracker) AggregateHourly() error {
	if t.db == nil || t.db.db == nil {
		return nil
	}

	// Aggregate last 2 hours of data (to catch any late entries)
	since := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Hour)

	_, err := t.db.db.Exec(`
		INSERT OR REPLACE INTO usage_hourly (hour, provider, model, project_path, total_input, total_output, total_cost, request_count)
		SELECT
			strftime('%Y-%m-%d %H:00:00', timestamp) as hour,
			provider,
			model,
			project_path,
			SUM(input_tokens),
			SUM(output_tokens),
			SUM(cost_usd),
			COUNT(*)
		FROM usage
		WHERE timestamp >= ?
		GROUP BY hour, provider, model, project_path
	`, since.Format(time.RFC3339Nano))

	return err
}

// GetHourlySummary returns hourly aggregated data for charts.
func (t *UsageTracker) GetHourlySummary(hours int) ([]HourlyUsage, error) {
	if t.db == nil || t.db.db == nil {
		return nil, nil
	}

	since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)

	rows, err := t.db.db.Query(`
		SELECT hour, SUM(total_input), SUM(total_output), SUM(total_cost), SUM(request_count)
		FROM usage_hourly
		WHERE hour >= ?
		GROUP BY hour
		ORDER BY hour ASC
	`, since.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []HourlyUsage
	for rows.Next() {
		var h HourlyUsage
		var hourStr string
		if err := rows.Scan(&hourStr, &h.InputTokens, &h.OutputTokens, &h.Cost, &h.RequestCount); err != nil {
			continue
		}
		if t, err := time.Parse("2006-01-02 15:04:05", hourStr); err == nil {
			h.Hour = t
		}
		result = append(result, h)
	}

	return result, nil
}

// HourlyUsage represents aggregated usage for one hour.
type HourlyUsage struct {
	Hour         time.Time `json:"hour"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	Cost         float64   `json:"cost"`
	RequestCount int       `json:"request_count"`
}

// GetRecentUsage returns recent usage entries.
func (t *UsageTracker) GetRecentUsage(limit int) ([]UsageEntry, error) {
	if t.db == nil || t.db.db == nil {
		return nil, nil
	}

	if limit <= 0 {
		limit = 100
	}

	rows, err := t.db.db.Query(`
		SELECT timestamp, session_id, provider, model, input_tokens, output_tokens, cost_usd, latency_ms, project_path, client_type
		FROM usage
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []UsageEntry
	for rows.Next() {
		var e UsageEntry
		var tsStr string
		if err := rows.Scan(&tsStr, &e.SessionID, &e.Provider, &e.Model, &e.InputTokens, &e.OutputTokens, &e.CostUSD, &e.LatencyMs, &e.ProjectPath, &e.ClientType); err != nil {
			continue
		}
		if t, err := time.Parse(time.RFC3339Nano, tsStr); err == nil {
			e.Timestamp = t
		}
		entries = append(entries, e)
	}

	return entries, nil
}

// --- Global usage tracker ---

var globalUsageTracker *UsageTracker

// InitGlobalUsageTracker initializes the global usage tracker.
func InitGlobalUsageTracker(db *LogDB) {
	globalUsageTracker = NewUsageTracker(db)
}

// GetGlobalUsageTracker returns the global usage tracker.
func GetGlobalUsageTracker() *UsageTracker {
	return globalUsageTracker
}
