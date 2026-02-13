package proxy

import (
	"encoding/json"
	"strings"

	"github.com/dopejs/opencc/internal/config"
)

const (
	defaultLongContextThreshold = 32000
	// Minimum token count for current request when using session history
	minCurrentTokensForSessionCheck = 20000
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
// Session history logic (inspired by claude-code-router):
// - lastUsage.InputTokens represents the ACTUAL tokens sent to API (after any compaction)
// - If the last request used > threshold tokens AND current request is substantial (>20k),
//   continue using longContext model for consistency in long conversations
// - This accounts for Claude Code's automatic context compaction
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

	// Check session history for long context continuation
	// This helps maintain model consistency in long conversations
	if sessionID != "" {
		lastUsage := GetSessionUsage(sessionID)
		if lastUsage != nil {
			// If last request's input (including context) exceeded threshold
			// and current request is substantial, continue using longContext
			// Note: lastUsage.InputTokens includes the full context sent to API,
			// which may have been compacted by Claude Code's autocompact feature
			if lastUsage.InputTokens > threshold && tokenCount > minCurrentTokensForSessionCheck {
				return true
			}
		}
	}

	// Check current request token count
	return tokenCount >= threshold
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
		// Non-empty map without explicit "disabled" is considered enabled
		return len(m) > 0
	}
	return false
}
