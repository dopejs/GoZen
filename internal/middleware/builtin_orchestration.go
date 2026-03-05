package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// OrchestrationConfig holds configuration for the multi-model orchestration middleware.
type OrchestrationConfig struct {
	Enabled       bool                    `json:"enabled"`
	DefaultMode   string                  `json:"default_mode,omitempty"`   // "single", "voting", "chain", "review"
	VotingModels  []string                `json:"voting_models,omitempty"`  // models for voting mode
	ChainConfig   *ChainOrchestrationConfig `json:"chain,omitempty"`        // chain mode config
	ReviewConfig  *ReviewOrchestrationConfig `json:"review,omitempty"`      // review mode config
	Timeout       int                     `json:"timeout,omitempty"`        // timeout in seconds (default: 120)
}

// ChainOrchestrationConfig configures chain processing mode.
type ChainOrchestrationConfig struct {
	DraftModel  string `json:"draft_model"`  // cheap model for initial draft
	RefineModel string `json:"refine_model"` // expensive model for refinement
}

// ReviewOrchestrationConfig configures adversarial review mode.
type ReviewOrchestrationConfig struct {
	CodeModel   string `json:"code_model"`   // model that writes code
	ReviewModel string `json:"review_model"` // model that reviews code
	MaxRounds   int    `json:"max_rounds"`   // max review rounds (default: 2)
}

// OrchestrationMiddleware provides multi-model orchestration capabilities.
// [BETA] This feature is experimental.
type OrchestrationMiddleware struct {
	config OrchestrationConfig
	client *http.Client
}

// NewOrchestration creates a new orchestration middleware.
func NewOrchestration() Middleware {
	return &OrchestrationMiddleware{
		config: OrchestrationConfig{
			DefaultMode: "single",
			Timeout:     120,
		},
		client: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

func (m *OrchestrationMiddleware) Name() string {
	return "orchestration"
}

func (m *OrchestrationMiddleware) Version() string {
	return "1.0.0"
}

func (m *OrchestrationMiddleware) Description() string {
	return "Multi-model orchestration: voting, chain processing, and adversarial review"
}

func (m *OrchestrationMiddleware) Priority() int {
	return 50 // Middle of the chain
}

func (m *OrchestrationMiddleware) Init(config json.RawMessage) error {
	if len(config) > 0 {
		if err := json.Unmarshal(config, &m.config); err != nil {
			return err
		}
	}

	// Set defaults
	if m.config.DefaultMode == "" {
		m.config.DefaultMode = "single"
	}
	if m.config.Timeout == 0 {
		m.config.Timeout = 120
	}
	if m.config.ReviewConfig != nil && m.config.ReviewConfig.MaxRounds == 0 {
		m.config.ReviewConfig.MaxRounds = 2
	}

	m.client.Timeout = time.Duration(m.config.Timeout) * time.Second

	return nil
}

func (m *OrchestrationMiddleware) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
	// Check for orchestration mode in metadata or use default
	mode := m.config.DefaultMode
	if modeOverride, ok := ctx.Metadata["orchestration_mode"].(string); ok {
		mode = modeOverride
	}

	// Store mode for response processing
	ctx.Metadata["orchestration_mode"] = mode
	ctx.Metadata["orchestration_original_body"] = ctx.Body

	return ctx, nil
}

func (m *OrchestrationMiddleware) ProcessResponse(ctx *ResponseContext) (*ResponseContext, error) {
	mode, _ := ctx.Request.Metadata["orchestration_mode"].(string)

	switch mode {
	case "voting":
		return m.processVoting(ctx)
	case "chain":
		return m.processChain(ctx)
	case "review":
		return m.processReview(ctx)
	default:
		// Single mode - pass through
		return ctx, nil
	}
}

func (m *OrchestrationMiddleware) Close() error {
	return nil
}

// processVoting implements voting mode - ask multiple models and take consensus.
func (m *OrchestrationMiddleware) processVoting(ctx *ResponseContext) (*ResponseContext, error) {
	if len(m.config.VotingModels) < 2 {
		return ctx, nil // Not enough models for voting
	}

	originalBody, _ := ctx.Request.Metadata["orchestration_original_body"].([]byte)
	if originalBody == nil {
		return ctx, nil
	}

	// Parse original request
	var reqData map[string]interface{}
	if err := json.Unmarshal(originalBody, &reqData); err != nil {
		return ctx, nil
	}

	// Get responses from other models in parallel
	var wg sync.WaitGroup
	responses := make([]string, len(m.config.VotingModels))
	responses[0] = m.extractResponseText(ctx.Body) // First response already received

	for i := 1; i < len(m.config.VotingModels); i++ {
		wg.Add(1)
		go func(idx int, model string) {
			defer wg.Done()
			reqData["model"] = model
			body, _ := json.Marshal(reqData)
			resp, err := m.sendRequest(ctx.Request, body)
			if err == nil {
				responses[idx] = m.extractResponseText(resp)
			}
		}(i, m.config.VotingModels[i])
	}
	wg.Wait()

	// Simple consensus: find most common response or combine
	consensus := m.findConsensus(responses)

	// Build new response with consensus
	newBody := m.buildResponse(consensus, ctx.InputTokens, ctx.OutputTokens)
	ctx.Body = newBody

	return ctx, nil
}

// processChain implements chain mode - draft with cheap model, refine with expensive.
func (m *OrchestrationMiddleware) processChain(ctx *ResponseContext) (*ResponseContext, error) {
	if m.config.ChainConfig == nil {
		return ctx, nil
	}

	// The initial response is from the draft model
	draftResponse := m.extractResponseText(ctx.Body)

	originalBody, _ := ctx.Request.Metadata["orchestration_original_body"].([]byte)
	if originalBody == nil {
		return ctx, nil
	}

	// Parse original request
	var reqData map[string]interface{}
	if err := json.Unmarshal(originalBody, &reqData); err != nil {
		return ctx, nil
	}

	// Build refinement request
	messages, _ := reqData["messages"].([]interface{})
	refinementPrompt := fmt.Sprintf(
		"Please review and improve the following draft response. Fix any errors, improve clarity, and ensure completeness:\n\n---\nDraft:\n%s\n---\n\nProvide the improved version:",
		draftResponse,
	)

	messages = append(messages, map[string]interface{}{
		"role":    "assistant",
		"content": draftResponse,
	})
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": refinementPrompt,
	})

	reqData["messages"] = messages
	reqData["model"] = m.config.ChainConfig.RefineModel

	body, _ := json.Marshal(reqData)
	refinedResp, err := m.sendRequest(ctx.Request, body)
	if err != nil {
		return ctx, nil // Fall back to draft response
	}

	ctx.Body = refinedResp
	return ctx, nil
}

// processReview implements adversarial review mode.
func (m *OrchestrationMiddleware) processReview(ctx *ResponseContext) (*ResponseContext, error) {
	if m.config.ReviewConfig == nil {
		return ctx, nil
	}

	codeResponse := m.extractResponseText(ctx.Body)

	originalBody, _ := ctx.Request.Metadata["orchestration_original_body"].([]byte)
	if originalBody == nil {
		return ctx, nil
	}

	var reqData map[string]interface{}
	if err := json.Unmarshal(originalBody, &reqData); err != nil {
		return ctx, nil
	}

	// Iterative review process
	for round := 0; round < m.config.ReviewConfig.MaxRounds; round++ {
		// Get review
		reviewPrompt := fmt.Sprintf(
			"Please review the following code for bugs, security issues, and improvements. Be specific and critical:\n\n```\n%s\n```\n\nList any issues found:",
			codeResponse,
		)

		reviewReqData := map[string]interface{}{
			"model": m.config.ReviewConfig.ReviewModel,
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": reviewPrompt,
				},
			},
		}

		reviewBody, _ := json.Marshal(reviewReqData)
		reviewResp, err := m.sendRequest(ctx.Request, reviewBody)
		if err != nil {
			break
		}

		review := m.extractResponseText(reviewResp)

		// Check if review found issues
		if !m.hasIssues(review) {
			break // No issues found, we're done
		}

		// Get fix from code model
		fixPrompt := fmt.Sprintf(
			"The following code was reviewed and issues were found. Please fix the issues:\n\nOriginal code:\n```\n%s\n```\n\nReview feedback:\n%s\n\nProvide the fixed code:",
			codeResponse,
			review,
		)

		messages, _ := reqData["messages"].([]interface{})
		messages = append(messages, map[string]interface{}{
			"role":    "assistant",
			"content": codeResponse,
		})
		messages = append(messages, map[string]interface{}{
			"role":    "user",
			"content": fixPrompt,
		})

		reqData["messages"] = messages
		reqData["model"] = m.config.ReviewConfig.CodeModel

		fixBody, _ := json.Marshal(reqData)
		fixResp, err := m.sendRequest(ctx.Request, fixBody)
		if err != nil {
			break
		}

		codeResponse = m.extractResponseText(fixResp)
		ctx.Body = fixResp
	}

	return ctx, nil
}

// sendRequest sends a request to the upstream provider.
func (m *OrchestrationMiddleware) sendRequest(reqCtx *RequestContext, body []byte) ([]byte, error) {
	// Build URL from provider info in context
	// Note: This is a simplified version - in production, you'd use the actual provider URL
	url := "http://127.0.0.1:19841" + reqCtx.Path

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range reqCtx.Headers {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// extractResponseText extracts the text content from a response body.
func (m *OrchestrationMiddleware) extractResponseText(body []byte) string {
	var respData struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &respData); err != nil {
		return ""
	}
	if len(respData.Content) == 0 {
		return ""
	}
	return respData.Content[0].Text
}

// findConsensus finds consensus among multiple responses.
func (m *OrchestrationMiddleware) findConsensus(responses []string) string {
	// Simple implementation: return the longest non-empty response
	// TODO: Implement proper consensus algorithm (e.g., semantic similarity)
	var best string
	for _, r := range responses {
		if len(r) > len(best) {
			best = r
		}
	}
	return best
}

// hasIssues checks if a review found any issues.
func (m *OrchestrationMiddleware) hasIssues(review string) bool {
	lower := strings.ToLower(review)
	noIssuePatterns := []string{
		"no issues", "looks good", "no problems", "no bugs",
		"code is correct", "no errors", "well written",
	}
	for _, pattern := range noIssuePatterns {
		if strings.Contains(lower, pattern) {
			return false
		}
	}
	return true
}

// buildResponse builds a response body with the given text.
func (m *OrchestrationMiddleware) buildResponse(text string, inputTokens, outputTokens int) []byte {
	resp := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": text,
			},
		},
		"usage": map[string]interface{}{
			"input_tokens":  inputTokens,
			"output_tokens": outputTokens,
		},
	}
	body, _ := json.Marshal(resp)
	return body
}
