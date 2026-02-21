package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// Default compression settings
const (
	DefaultThresholdTokens = 50000
	DefaultTargetTokens    = 20000
	DefaultPreserveRecent  = 4
	DefaultSummaryModel    = "claude-3-haiku-20240307"

	// Approximate tokens per character (conservative estimate)
	tokensPerChar = 0.25
)

// Message represents a conversation message for compression.
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// ContextCompressor handles context compression for large conversations.
// [BETA] This feature is experimental and disabled by default.
type ContextCompressor struct {
	config    *config.CompressionConfig
	client    *http.Client
	providers []*Provider
	mu        sync.RWMutex

	// Stats
	requestsCompressed int64
	tokensSaved        int64
}

// CompressionStats holds compression statistics.
type CompressionStats struct {
	RequestsCompressed int64 `json:"requests_compressed"`
	TokensSaved        int64 `json:"tokens_saved"`
}

// Global compressor instance
var (
	globalCompressor     *ContextCompressor
	globalCompressorOnce sync.Once
	globalCompressorMu   sync.RWMutex
)

// InitGlobalCompressor initializes the global context compressor.
func InitGlobalCompressor(providers []*Provider) {
	globalCompressorOnce.Do(func() {
		cfg := config.GetCompression()
		if cfg == nil {
			cfg = &config.CompressionConfig{
				Enabled:         false,
				ThresholdTokens: DefaultThresholdTokens,
				TargetTokens:    DefaultTargetTokens,
				SummaryModel:    DefaultSummaryModel,
				PreserveRecent:  DefaultPreserveRecent,
			}
		}
		globalCompressorMu.Lock()
		globalCompressor = NewContextCompressor(cfg, providers)
		globalCompressorMu.Unlock()
	})
}

// GetGlobalCompressor returns the global context compressor.
func GetGlobalCompressor() *ContextCompressor {
	globalCompressorMu.RLock()
	defer globalCompressorMu.RUnlock()
	return globalCompressor
}

// UpdateGlobalCompressorConfig updates the global compressor configuration.
func UpdateGlobalCompressorConfig(cfg *config.CompressionConfig) {
	globalCompressorMu.Lock()
	defer globalCompressorMu.Unlock()
	if globalCompressor != nil {
		globalCompressor.UpdateConfig(cfg)
	}
}

// UpdateGlobalCompressorProviders updates the providers for the global compressor.
func UpdateGlobalCompressorProviders(providers []*Provider) {
	globalCompressorMu.Lock()
	defer globalCompressorMu.Unlock()
	if globalCompressor != nil {
		globalCompressor.SetProviders(providers)
	}
}

// NewContextCompressor creates a new context compressor.
func NewContextCompressor(cfg *config.CompressionConfig, providers []*Provider) *ContextCompressor {
	if cfg == nil {
		cfg = &config.CompressionConfig{
			Enabled:         false,
			ThresholdTokens: DefaultThresholdTokens,
			TargetTokens:    DefaultTargetTokens,
			SummaryModel:    DefaultSummaryModel,
			PreserveRecent:  DefaultPreserveRecent,
		}
	}
	return &ContextCompressor{
		config:    cfg,
		providers: providers,
		client: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

// UpdateConfig updates the compressor configuration.
func (c *ContextCompressor) UpdateConfig(cfg *config.CompressionConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = cfg
}

// SetProviders updates the available providers.
func (c *ContextCompressor) SetProviders(providers []*Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.providers = providers
}

// IsEnabled returns whether compression is enabled.
func (c *ContextCompressor) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config != nil && c.config.Enabled
}

// GetStats returns compression statistics.
func (c *ContextCompressor) GetStats() CompressionStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return CompressionStats{
		RequestsCompressed: c.requestsCompressed,
		TokensSaved:        c.tokensSaved,
	}
}

// ShouldCompress determines if the messages should be compressed.
func (c *ContextCompressor) ShouldCompress(messages []Message) bool {
	if !c.IsEnabled() {
		return false
	}

	c.mu.RLock()
	threshold := c.config.ThresholdTokens
	c.mu.RUnlock()

	if threshold <= 0 {
		threshold = DefaultThresholdTokens
	}

	tokens := c.EstimateTokens(messages)
	return tokens > threshold
}

// EstimateTokens estimates the token count for messages.
func (c *ContextCompressor) EstimateTokens(messages []Message) int {
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Role)
		totalChars += estimateContentLength(msg.Content)
	}
	return int(float64(totalChars) * tokensPerChar)
}

// estimateContentLength estimates the character length of message content.
func estimateContentLength(content interface{}) int {
	switch v := content.(type) {
	case string:
		return len(v)
	case []interface{}:
		total := 0
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if text, ok := m["text"].(string); ok {
					total += len(text)
				}
			}
		}
		return total
	default:
		// Fallback: marshal to JSON and count
		if data, err := json.Marshal(content); err == nil {
			return len(data)
		}
		return 0
	}
}

// __CONTINUE_HERE__

// Compress compresses the messages by summarizing older messages.
// Returns the compressed messages and any error.
func (c *ContextCompressor) Compress(messages []Message) ([]Message, error) {
	if !c.IsEnabled() || len(messages) == 0 {
		return messages, nil
	}

	c.mu.RLock()
	preserveRecent := c.config.PreserveRecent
	targetTokens := c.config.TargetTokens
	c.mu.RUnlock()

	if preserveRecent <= 0 {
		preserveRecent = DefaultPreserveRecent
	}
	if targetTokens <= 0 {
		targetTokens = DefaultTargetTokens
	}

	// If we don't have enough messages to compress, return as-is
	if len(messages) <= preserveRecent {
		return messages, nil
	}

	// Split messages: older ones to summarize, recent ones to preserve
	toSummarize := messages[:len(messages)-preserveRecent]
	toPreserve := messages[len(messages)-preserveRecent:]

	// Estimate tokens before compression
	tokensBefore := c.EstimateTokens(messages)

	// Summarize older messages
	summary, err := c.Summarize(toSummarize)
	if err != nil {
		// On error, return original messages
		return messages, fmt.Errorf("compression failed: %w", err)
	}

	// Build compressed message list
	compressed := make([]Message, 0, len(toPreserve)+1)

	// Add summary as a system message or user message
	summaryMsg := Message{
		Role:    "user",
		Content: fmt.Sprintf("[Previous conversation summary]\n%s\n[End of summary]", summary),
	}
	compressed = append(compressed, summaryMsg)
	compressed = append(compressed, toPreserve...)

	// Update stats
	tokensAfter := c.EstimateTokens(compressed)
	c.mu.Lock()
	c.requestsCompressed++
	c.tokensSaved += int64(tokensBefore - tokensAfter)
	c.mu.Unlock()

	return compressed, nil
}

// Summarize generates a summary of the given messages using a cheap model.
func (c *ContextCompressor) Summarize(messages []Message) (string, error) {
	c.mu.RLock()
	summaryModel := c.config.SummaryModel
	summaryProvider := c.config.SummaryProvider
	providers := c.providers
	c.mu.RUnlock()

	if summaryModel == "" {
		summaryModel = DefaultSummaryModel
	}

	// Find provider to use for summarization
	var provider *Provider
	if summaryProvider != "" {
		for _, p := range providers {
			if p.Name == summaryProvider && p.IsHealthy() {
				provider = p
				break
			}
		}
	}
	if provider == nil {
		// Use first healthy provider
		for _, p := range providers {
			if p.IsHealthy() {
				provider = p
				break
			}
		}
	}
	if provider == nil && len(providers) > 0 {
		// Fallback to first provider even if unhealthy
		provider = providers[0]
	}
	if provider == nil {
		return "", fmt.Errorf("no provider available for summarization")
	}

	// Build conversation text for summarization
	var convBuilder strings.Builder
	for _, msg := range messages {
		convBuilder.WriteString(fmt.Sprintf("%s: ", msg.Role))
		switch v := msg.Content.(type) {
		case string:
			convBuilder.WriteString(v)
		default:
			if data, err := json.Marshal(v); err == nil {
				convBuilder.WriteString(string(data))
			}
		}
		convBuilder.WriteString("\n\n")
	}

	// Create summarization request
	summaryPrompt := fmt.Sprintf(`Please provide a concise summary of the following conversation.
Focus on:
1. Key topics discussed
2. Important decisions or conclusions
3. Any pending questions or tasks
4. Critical context needed for continuing the conversation

Keep the summary under 500 words.

Conversation:
%s`, convBuilder.String())

	reqBody := map[string]interface{}{
		"model":      summaryModel,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": summaryPrompt},
		},
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to provider
	url := provider.BaseURL.String() + "/v1/messages"
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", provider.Token)
	req.Header.Set("Authorization", "Bearer "+provider.Token)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("summarization request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("summarization failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var respData struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(respData.Content) == 0 {
		return "", fmt.Errorf("empty response from summarization")
	}

	return respData.Content[0].Text, nil
}

// CompressRequestBody compresses the request body if needed.
// Returns the potentially modified body and whether compression was applied.
func (c *ContextCompressor) CompressRequestBody(body []byte) ([]byte, bool, error) {
	if !c.IsEnabled() {
		return body, false, nil
	}

	var reqData map[string]interface{}
	if err := json.Unmarshal(body, &reqData); err != nil {
		return body, false, nil
	}

	messagesRaw, ok := reqData["messages"]
	if !ok {
		return body, false, nil
	}

	messagesData, err := json.Marshal(messagesRaw)
	if err != nil {
		return body, false, nil
	}

	var messages []Message
	if err := json.Unmarshal(messagesData, &messages); err != nil {
		return body, false, nil
	}

	if !c.ShouldCompress(messages) {
		return body, false, nil
	}

	compressed, err := c.Compress(messages)
	if err != nil {
		return body, false, err
	}

	reqData["messages"] = compressed
	newBody, err := json.Marshal(reqData)
	if err != nil {
		return body, false, err
	}

	return newBody, true, nil
}
