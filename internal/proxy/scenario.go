package proxy

import (
	"encoding/json"
	"strings"

	"github.com/dopejs/gozen/internal/config"
)

const (
	defaultLongContextThreshold = 32000
	// Minimum token count for current request when using session history
	// If current request is below this, assume context was cleared
	minCurrentTokensForSessionCheck = 5000
	// Ratio threshold: if current tokens are less than 20% of last session's input,
	// assume context was cleared or significantly compacted
	sessionClearRatio = 0.2
)

// DetectScenario examines a parsed request body and returns the matching scenario.
// Priority: webSearch > think > image > longContext > background > default.
func DetectScenario(body map[string]interface{}, threshold int, sessionID string) config.Scenario {
	if hasWebSearchTool(body) {
		return config.ScenarioWebSearch
	}
	if hasThinkingEnabled(body) {
		return config.ScenarioThink
	}
	if hasImageContent(body) {
		return config.ScenarioImage
	}
	if isLongContext(body, threshold, sessionID) {
		return config.ScenarioLongContext
	}
	if isBackgroundRequest(body) {
		return config.ScenarioBackground
	}
	return config.ScenarioDefault
}

// DetectScenarioFromJSON parses raw JSON and detects the scenario.
func DetectScenarioFromJSON(data []byte, threshold int, sessionID string) (config.Scenario, map[string]interface{}) {
	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return config.ScenarioDefault, nil
	}
	return DetectScenario(body, threshold, sessionID), body
}

// hasImageContent checks if any message contains an image content block.
func hasImageContent(body map[string]interface{}) bool {
	messages, ok := body["messages"].([]interface{})
	if !ok {
		return false
	}
	for _, msg := range messages {
		m, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}
		content, ok := m["content"].([]interface{})
		if !ok {
			continue
		}
		for _, block := range content {
			b, ok := block.(map[string]interface{})
			if !ok {
				continue
			}
			if t, ok := b["type"].(string); ok && t == "image" {
				return true
			}
		}
	}
	return false
}

// isLongContext checks if the total text content in messages exceeds the threshold.
// It uses tiktoken for accurate token counting and considers session history.
//
// Session history logic:
// - lastUsage.InputTokens represents the ACTUAL tokens sent to API (after any compaction)
// - If current request tokens are significantly lower than last session (< 20%),
//   assume context was cleared and DON'T use session history
// - This accounts for /clear commands and context resets
func isLongContext(body map[string]interface{}, threshold int, sessionID string) bool {
	if threshold <= 0 {
		threshold = defaultLongContextThreshold
	}

	// Calculate current request token count
	tokenCount, err := calculateTokenCount(body)
	if err != nil {
		// Fallback to character-based estimation on error
		tokenCount = estimateTokensFromChars(body)
	}

	// Check current request token count first
	if tokenCount >= threshold {
		return true
	}

	// Check session history for long context continuation
	// This helps maintain model consistency in long conversations
	if sessionID != "" {
		lastUsage := GetSessionUsage(sessionID)
		if lastUsage != nil && lastUsage.InputTokens > 0 {
			// Detect context clear: if current tokens are much smaller than last session,
			// the user likely cleared context or it was heavily compacted
			ratio := float64(tokenCount) / float64(lastUsage.InputTokens)
			if ratio < sessionClearRatio {
				// Context was likely cleared, don't use session history
				// Also clear the session cache to prevent future false positives
				ClearSessionUsage(sessionID)
				return false
			}

			// If last request's input exceeded threshold and current request
			// is substantial (not a tiny follow-up), continue using longContext
			if lastUsage.InputTokens > threshold && tokenCount > minCurrentTokensForSessionCheck {
				return true
			}
		}
	}

	return false
}

// hasWebSearchTool checks if the request includes web_search tools.
func hasWebSearchTool(body map[string]interface{}) bool {
	tools, ok := body["tools"].([]interface{})
	if !ok {
		return false
	}
	for _, tool := range tools {
		t, ok := tool.(map[string]interface{})
		if !ok {
			continue
		}
		if toolType, ok := t["type"].(string); ok && strings.HasPrefix(toolType, "web_search") {
			return true
		}
	}
	return false
}

// isBackgroundRequest checks if the request is for a Haiku model (background task).
func isBackgroundRequest(body map[string]interface{}) bool {
	model, ok := body["model"].(string)
	if !ok {
		return false
	}
	modelLower := strings.ToLower(model)
	return strings.Contains(modelLower, "claude") && strings.Contains(modelLower, "haiku")
}

// hasThinkingEnabled checks if the request has thinking mode enabled.
func hasThinkingEnabled(body map[string]interface{}) bool {
	thinking, ok := body["thinking"]
	if !ok {
		return false
	}
	// Check if thinking is a boolean true
	if b, ok := thinking.(bool); ok {
		return b
	}
	// Check if thinking is a map with type="enabled"
	if m, ok := thinking.(map[string]interface{}); ok {
		if t, ok := m["type"].(string); ok {
			return t == "enabled"
		}
		return false
	}
	return false
}
