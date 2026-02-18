package proxy

import (
	"sync"

	"github.com/dopejs/gozen/internal/config"
)

// BudgetStatus represents the current budget status.
type BudgetStatus struct {
	DailySpent     float64 `json:"daily_spent"`
	DailyLimit     float64 `json:"daily_limit,omitempty"`
	DailyRemaining float64 `json:"daily_remaining,omitempty"`
	DailyPercent   float64 `json:"daily_percent,omitempty"`

	WeeklySpent     float64 `json:"weekly_spent"`
	WeeklyLimit     float64 `json:"weekly_limit,omitempty"`
	WeeklyRemaining float64 `json:"weekly_remaining,omitempty"`
	WeeklyPercent   float64 `json:"weekly_percent,omitempty"`

	MonthlySpent     float64 `json:"monthly_spent"`
	MonthlyLimit     float64 `json:"monthly_limit,omitempty"`
	MonthlyRemaining float64 `json:"monthly_remaining,omitempty"`
	MonthlyPercent   float64 `json:"monthly_percent,omitempty"`

	ShouldWarn      bool                `json:"should_warn"`
	ShouldDowngrade bool                `json:"should_downgrade"`
	ShouldBlock     bool                `json:"should_block"`
	ActiveAction    config.BudgetAction `json:"active_action,omitempty"`
	Message         string              `json:"message,omitempty"`
}

// BudgetChecker checks spending against configured budget limits.
type BudgetChecker struct {
	tracker *UsageTracker
	mu      sync.RWMutex
	config  *config.BudgetConfig
}

// NewBudgetChecker creates a new budget checker.
func NewBudgetChecker(tracker *UsageTracker) *BudgetChecker {
	return &BudgetChecker{
		tracker: tracker,
		config:  config.GetBudgets(),
	}
}

// ReloadConfig refreshes the budget configuration.
func (c *BudgetChecker) ReloadConfig() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = config.GetBudgets()
}

// Check returns the current budget status for a project.
// If projectPath is empty and PerProject is false, checks global budget.
func (c *BudgetChecker) Check(projectPath string) (*BudgetStatus, error) {
	c.mu.RLock()
	cfg := c.config
	c.mu.RUnlock()

	status := &BudgetStatus{}

	if c.tracker == nil {
		return status, nil
	}

	// Determine which project to check
	checkProject := ""
	if cfg != nil && cfg.PerProject && projectPath != "" {
		checkProject = projectPath
	}

	// Get current spending
	var err error
	status.DailySpent, err = c.tracker.GetDailyCost(checkProject)
	if err != nil {
		return nil, err
	}

	status.WeeklySpent, err = c.tracker.GetWeeklyCost(checkProject)
	if err != nil {
		return nil, err
	}

	status.MonthlySpent, err = c.tracker.GetMonthlyCost(checkProject)
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		return status, nil
	}

	// Check daily limit
	if cfg.Daily != nil && cfg.Daily.Amount > 0 {
		status.DailyLimit = cfg.Daily.Amount
		status.DailyRemaining = cfg.Daily.Amount - status.DailySpent
		if status.DailyRemaining < 0 {
			status.DailyRemaining = 0
		}
		status.DailyPercent = (status.DailySpent / cfg.Daily.Amount) * 100

		c.checkLimit(status, status.DailySpent, cfg.Daily, "daily")
	}

	// Check weekly limit
	if cfg.Weekly != nil && cfg.Weekly.Amount > 0 {
		status.WeeklyLimit = cfg.Weekly.Amount
		status.WeeklyRemaining = cfg.Weekly.Amount - status.WeeklySpent
		if status.WeeklyRemaining < 0 {
			status.WeeklyRemaining = 0
		}
		status.WeeklyPercent = (status.WeeklySpent / cfg.Weekly.Amount) * 100

		c.checkLimit(status, status.WeeklySpent, cfg.Weekly, "weekly")
	}

	// Check monthly limit
	if cfg.Monthly != nil && cfg.Monthly.Amount > 0 {
		status.MonthlyLimit = cfg.Monthly.Amount
		status.MonthlyRemaining = cfg.Monthly.Amount - status.MonthlySpent
		if status.MonthlyRemaining < 0 {
			status.MonthlyRemaining = 0
		}
		status.MonthlyPercent = (status.MonthlySpent / cfg.Monthly.Amount) * 100

		c.checkLimit(status, status.MonthlySpent, cfg.Monthly, "monthly")
	}

	return status, nil
}

// checkLimit updates status based on a single limit check.
func (c *BudgetChecker) checkLimit(status *BudgetStatus, spent float64, limit *config.BudgetLimit, period string) {
	if limit == nil || limit.Amount <= 0 {
		return
	}

	percent := (spent / limit.Amount) * 100

	// Warning threshold at 80%
	if percent >= 80 && percent < 100 {
		status.ShouldWarn = true
		if status.Message == "" {
			status.Message = period + " budget at " + formatPercent(percent)
		}
	}

	// Over limit
	if percent >= 100 {
		switch limit.Action {
		case config.BudgetActionWarn:
			status.ShouldWarn = true
			status.ActiveAction = config.BudgetActionWarn
			status.Message = period + " budget exceeded"

		case config.BudgetActionDowngrade:
			status.ShouldDowngrade = true
			status.ActiveAction = config.BudgetActionDowngrade
			status.Message = period + " budget exceeded, downgrading model"

		case config.BudgetActionBlock:
			status.ShouldBlock = true
			status.ActiveAction = config.BudgetActionBlock
			status.Message = period + " budget exceeded, requests blocked"

		default:
			// Default to warn
			status.ShouldWarn = true
			status.ActiveAction = config.BudgetActionWarn
			status.Message = period + " budget exceeded"
		}
	}
}

// ShouldBlock returns true if requests should be blocked due to budget.
func (c *BudgetChecker) ShouldBlock(projectPath string) bool {
	status, err := c.Check(projectPath)
	if err != nil {
		return false
	}
	return status.ShouldBlock
}

// ShouldDowngrade returns true if model should be downgraded due to budget.
func (c *BudgetChecker) ShouldDowngrade(projectPath string) bool {
	status, err := c.Check(projectPath)
	if err != nil {
		return false
	}
	return status.ShouldDowngrade
}

// GetDowngradeModel returns a cheaper model to use when budget is exceeded.
func (c *BudgetChecker) GetDowngradeModel(currentModel string) string {
	// Downgrade to haiku as the cheapest option
	return "claude-3-5-haiku-20241022"
}

func formatPercent(p float64) string {
	if p >= 100 {
		return "100%"
	}
	return string(rune('0'+int(p/10))) + string(rune('0'+int(p)%10)) + "%"
}

// --- Global budget checker ---

var globalBudgetChecker *BudgetChecker

// InitGlobalBudgetChecker initializes the global budget checker.
func InitGlobalBudgetChecker(tracker *UsageTracker) {
	globalBudgetChecker = NewBudgetChecker(tracker)
}

// GetGlobalBudgetChecker returns the global budget checker.
func GetGlobalBudgetChecker() *BudgetChecker {
	return globalBudgetChecker
}
