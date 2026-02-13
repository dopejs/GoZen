package proxy

import (
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

var (
	// Global tiktoken encoder for cl100k_base (used by Claude models)
	tokenEncoder     *tiktoken.Tiktoken
	tokenEncoderOnce sync.Once
	tokenEncoderErr  error
)

// getTokenEncoder returns the global tiktoken encoder instance.
func getTokenEncoder() (*tiktoken.Tiktoken, error) {
	tokenEncoderOnce.Do(func() {
		// cl100k_base is the encoding used by Claude models
		tokenEncoder, tokenEncoderErr = tiktoken.GetEncoding("cl100k_base")
	})
	return tokenEncoder, tokenEncoderErr
}

// calculateTokenCount calculates the total token count for a request body.
// It counts tokens in messages, system prompt, and tools.
func calculateTokenCount(body map[string]interface{}) (int, error) {
	enc, err := getTokenEncoder()
	if err != nil {
		// Fallback to character-based estimation if tiktoken fails
		return estimateTokensFromChars(body), nil
	}

	totalTokens := 0

	// Count tokens in messages
	if messages, ok := body["messages"].([]interface{}); ok {
		for _, msg := range messages {
			m, ok := msg.(map[string]interface{})
			if !ok {
				continue
			}

			// Count content tokens
			switch content := m["content"].(type) {
			case string:
				totalTokens += len(enc.Encode(content, nil, nil))
			case []interface{}:
				for _, block := range content {
					b, ok := block.(map[string]interface{})
					if !ok {
						continue
					}
					blockType, _ := b["type"].(string)
					switch blockType {
					case "text":
						if text, ok := b["text"].(string); ok {
							totalTokens += len(enc.Encode(text, nil, nil))
						}
					case "tool_use":
						// Count tool use input as JSON string
						if input, ok := b["input"]; ok {
							totalTokens += estimateJSONTokens(enc, input)
						}
					case "tool_result":
						// Count tool result content
						if resultContent, ok := b["content"].(string); ok {
							totalTokens += len(enc.Encode(resultContent, nil, nil))
						} else if resultContent, ok := b["content"].([]interface{}); ok {
							for _, rc := range resultContent {
								if rcMap, ok := rc.(map[string]interface{}); ok {
									if text, ok := rcMap["text"].(string); ok {
										totalTokens += len(enc.Encode(text, nil, nil))
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Count tokens in system prompt
	switch system := body["system"].(type) {
	case string:
		totalTokens += len(enc.Encode(system, nil, nil))
	case []interface{}:
		for _, item := range system {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemType, ok := itemMap["type"].(string); ok && itemType == "text" {
					if text, ok := itemMap["text"].(string); ok {
						totalTokens += len(enc.Encode(text, nil, nil))
					}
				}
			}
		}
	}

	// Count tokens in tools
	if tools, ok := body["tools"].([]interface{}); ok {
		for _, tool := range tools {
			t, ok := tool.(map[string]interface{})
			if !ok {
				continue
			}
			// Count tool name and description
			if name, ok := t["name"].(string); ok {
				totalTokens += len(enc.Encode(name, nil, nil))
			}
			if desc, ok := t["description"].(string); ok {
				totalTokens += len(enc.Encode(desc, nil, nil))
			}
			// Count input schema (approximate)
			if schema, ok := t["input_schema"]; ok {
				totalTokens += estimateJSONTokens(enc, schema)
			}
		}
	}

	return totalTokens, nil
}

// estimateJSONTokens estimates token count for a JSON object by encoding it as string.
func estimateJSONTokens(enc *tiktoken.Tiktoken, obj interface{}) int {
	// Simple approximation: convert to string and count
	// This is not perfect but good enough for schema/input objects
	switch v := obj.(type) {
	case string:
		return len(enc.Encode(v, nil, nil))
	case map[string]interface{}:
		total := 2 // {} brackets
		for key, val := range v {
			total += len(enc.Encode(key, nil, nil))
			total += estimateJSONTokens(enc, val)
			total += 2 // : and ,
		}
		return total
	case []interface{}:
		total := 2 // [] brackets
		for _, item := range v {
			total += estimateJSONTokens(enc, item)
			total += 1 // ,
		}
		return total
	default:
		// For numbers, booleans, etc., estimate ~5 tokens
		return 5
	}
}

// estimateTokensFromChars provides a fallback character-based estimation.
// This is used when tiktoken is unavailable.
func estimateTokensFromChars(body map[string]interface{}) int {
	totalChars := 0

	// Count characters in messages
	if messages, ok := body["messages"].([]interface{}); ok {
		for _, msg := range messages {
			m, ok := msg.(map[string]interface{})
			if !ok {
				continue
			}
			switch content := m["content"].(type) {
			case string:
				totalChars += len(content)
			case []interface{}:
				for _, block := range content {
					b, ok := block.(map[string]interface{})
					if !ok {
						continue
					}
					if text, ok := b["text"].(string); ok {
						totalChars += len(text)
					}
				}
			}
		}
	}

	// Count characters in system prompt
	if system, ok := body["system"].(string); ok {
		totalChars += len(system)
	} else if systemBlocks, ok := body["system"].([]interface{}); ok {
		for _, block := range systemBlocks {
			b, ok := block.(map[string]interface{})
			if !ok {
				continue
			}
			if text, ok := b["text"].(string); ok {
				totalChars += len(text)
			}
		}
	}

	// Rough estimation: 1 token ≈ 4 characters for English, 1.5 for Chinese
	// Use conservative estimate of 3 characters per token
	return totalChars / 3
}
