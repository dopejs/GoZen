package proxy

import (
	"strings"
	"unicode"
)

// NormalizeScenarioKey converts scenario keys to canonical camelCase format.
// Supports kebab-case, snake_case, and camelCase inputs.
// Examples:
//   - "web-search" → "webSearch"
//   - "long_context" → "longContext"
//   - "webSearch" → "webSearch" (unchanged)
//   - "think" → "think" (unchanged)
func NormalizeScenarioKey(key string) string {
	if key == "" {
		return ""
	}

	// Split on hyphens and underscores
	parts := splitOnDelimiters(key)
	if len(parts) == 0 {
		return key
	}

	// First part stays lowercase, rest are title-cased
	result := strings.ToLower(parts[0])
	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			result += titleCase(parts[i])
		}
	}

	return result
}

// splitOnDelimiters splits a string on hyphens and underscores
func splitOnDelimiters(s string) []string {
	var parts []string
	var current strings.Builder

	for _, r := range s {
		if r == '-' || r == '_' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// titleCase converts the first character to uppercase, rest to lowercase
func titleCase(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}
	return string(runes)
}
