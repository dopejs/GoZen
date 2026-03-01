package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MatchResult represents a single intent match result.
type MatchResult struct {
	Skill       string      `json:"skill"`        // matched skill name
	Intent      Intent      `json:"intent"`       // recognized intent
	Confidence  float64     `json:"confidence"`   // score (0.0 - 1.0)
	Method      string      `json:"method"`       // "local" or "llm"
	ParsedIntent *ParsedIntent `json:"parsed_intent,omitempty"`
}

// SkillScore contains detailed scoring for a skill.
type SkillScore struct {
	SkillName    string  `json:"skill_name"`
	KeywordScore float64 `json:"keyword_score"`
	SynonymScore float64 `json:"synonym_score"`
	FuzzyScore   float64 `json:"fuzzy_score"`
	LocalScore   float64 `json:"local_score"`
	LLMScore     float64 `json:"llm_score"`
	FinalScore   float64 `json:"final_score"`
}

// MatchLog records a match attempt for debugging.
type MatchLog struct {
	Timestamp time.Time `json:"timestamp"`
	Input     string    `json:"input"`
	Platform  string    `json:"platform"`
	UserID    string    `json:"user_id"`
	Scores    []SkillScore `json:"scores,omitempty"`
	Result    *MatchResult `json:"result"`
	Duration  time.Duration `json:"duration_ms"`
	LLMUsed   bool        `json:"llm_used"`
}

// LLMClassifier interface for LLM-based classification and parameter extraction.
type LLMClassifier interface {
	// Classify returns the matched skill and intent via LLM.
	Classify(ctx context.Context, message string, skills []*Skill) (string, error) // JSON response
	// ExtractParams extracts intent-specific parameters via LLM.
	ExtractParams(ctx context.Context, message string, intent Intent) (map[string]string, error)
}

// SkillMatcher implements hybrid intent matching (local + LLM fallback).
type SkillMatcher struct {
	reg        *SkillRegistry
	classifier LLMClassifier
	threshold  float64
	logBuffer  *MatchLogBuffer
}

// NewSkillMatcher creates a new skill matcher.
func NewSkillMatcher(reg *SkillRegistry, classifier LLMClassifier, threshold float64, bufferSize int) *SkillMatcher {
	sm := &SkillMatcher{
		reg:        reg,
		classifier: classifier,
		threshold:  threshold,
	}
	if bufferSize > 0 {
		sm.logBuffer = NewMatchLogBuffer(bufferSize)
	}
	return sm
}

// Classifier returns the LLM classifier used by the matcher (may be nil).
func (m *SkillMatcher) Classifier() LLMClassifier {
	return m.classifier
}

// GetMatchLogs returns the most recent match logs, up to limit.
func (m *SkillMatcher) GetMatchLogs(limit int) []*MatchLog {
	if m.logBuffer == nil {
		return nil
	}
	return m.logBuffer.List(limit)
}

// ClearMatchLogs clears all match logs.
func (m *SkillMatcher) ClearMatchLogs() {
	if m.logBuffer != nil {
		m.logBuffer.Clear()
	}
}

// MatchLocal attempts to match using local keyword/synonym/fuzzy matching.
func (m *SkillMatcher) MatchLocal(message string) *MatchResult {
	skills := m.reg.ListEnabled()
	if len(skills) == 0 {
		return nil
	}

	message = strings.ToLower(strings.TrimSpace(message))
	bestScore := 0.0
	var bestSkill *Skill

	for _, s := range skills {
		if !s.Enabled {
			continue
		}
		score := m.computeLocalScore(s, message)
		if score > bestScore {
			bestScore = score
			bestSkill = s
		}
	}

	if bestSkill == nil || bestScore < m.threshold {
		return nil
	}

	return &MatchResult{
		Skill:      bestSkill.Name,
		Intent:     bestSkill.Intent,
		Confidence: bestScore,
		Method:     "local",
	}
}

func (m *SkillMatcher) computeLocalScore(s *Skill, message string) float64 {
	// Exact keyword match (highest confidence)
	for _, keywords := range s.Keywords {
		for _, kw := range keywords {
			if strings.ToLower(kw) == message {
				return 0.95
			}
		}
	}

	// Exact synonym match (high confidence)
	for variant := range s.Synonyms {
		if variant == message {
			return 0.9
		}
	}

	// Keyword substring match
	keywordScore := 0.0
	for _, keywords := range s.Keywords {
		for _, kw := range keywords {
			kwLower := strings.ToLower(kw)
			if strings.Contains(message, kwLower) {
				// Substring match: ensure confidence above threshold
				subScore := 0.5 + 0.5*float64(len(kwLower))/float64(len(message))
				if subScore < 0.8 {
					subScore = 0.8 // minimum for substring keyword match
				}
				if subScore > keywordScore {
					keywordScore = subScore
				}
			}
		}
		if keywordScore >= 0.8 {
			break
		}
	}

	// Synonym substring match
	synonymScore := 0.0
	for variant := range s.Synonyms {
		if strings.Contains(message, variant) {
			synScore := 0.4 + 0.4*float64(len(variant))/float64(len(message))
			if synScore < 0.75 {
				synScore = 0.75 // minimum for synonym substring match
			}
			if synScore > synonymScore {
				synonymScore = synScore
			}
		}
	}

	// Fuzzy word match
	fuzzyScore := 0.0
	words := strings.Fields(message)
	for _, w := range words {
		for _, keywords := range s.Keywords {
			for _, kw := range keywords {
				if strings.Contains(kw, w) || strings.Contains(w, kw) {
					fuzzyScore = 0.6
					break
				}
			}
		}
		for variant := range s.Synonyms {
			if strings.Contains(variant, w) || strings.Contains(w, variant) {
				fuzzyScore = 0.6
				break
			}
		}
	}

	// Weighted combination favoring keyword matches
	total := keywordScore*0.9 + synonymScore*0.08 + fuzzyScore*0.02
	if total > 1.0 {
		total = 1.0
	}
	return total
}

// MatchLLM attempts to match using LLM fallback.
func (m *SkillMatcher) MatchLLM(ctx context.Context, message string) (*MatchResult, error) {
	if m.classifier == nil {
		return nil, fmt.Errorf("no LLM classifier available")
	}

	skills := m.reg.ListEnabled()
	if len(skills) == 0 {
		return nil, fmt.Errorf("no skills available")
	}

	respJSON, err := m.classifier.Classify(ctx, message, skills)
	if err != nil {
		return nil, fmt.Errorf("LLM classification: %w", err)
	}

	var resp struct {
		Skill      string  `json:"skill"`
		Intent     string  `json:"intent"`
		Confidence float64 `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		return nil, fmt.Errorf("parse LLM response: %w", err)
	}

	// Validate skill exists and is enabled
	skill := m.reg.Get(resp.Skill)
	if skill == nil || !skill.Enabled {
		return nil, fmt.Errorf("LLM returned invalid skill: %s", resp.Skill)
	}

	return &MatchResult{
		Skill:      resp.Skill,
		Intent:     Intent(resp.Intent),
		Confidence: resp.Confidence,
		Method:     "llm",
	}, nil
}

// recordMatchLog records a match attempt to the log buffer if available.
func (m *SkillMatcher) recordMatchLog(message string, result *MatchResult, duration time.Duration, llmUsed bool) {
	if m.logBuffer == nil {
		return
	}
	log := &MatchLog{
		Timestamp: time.Now(),
		Input:     message,
		Result:    result,
		Duration:  duration,
		LLMUsed:   llmUsed,
	}
	m.logBuffer.Add(log)
}

// Match orchestrates local → LLM fallback → chat fallback.
func (m *SkillMatcher) Match(ctx context.Context, message string) *MatchResult {
	// First try local matching
	localStart := time.Now()
	localResult := m.MatchLocal(message)
	localDuration := time.Since(localStart)

	// Record local match attempt
	m.recordMatchLog(message, localResult, localDuration, false)

	if localResult != nil && localResult.Confidence >= m.threshold {
		return localResult
	}

	// Fall back to LLM if available
	if m.classifier != nil {
		llmStart := time.Now()
		llmResult, err := m.MatchLLM(ctx, message)
		llmDuration := time.Since(llmStart)

		// Record LLM match attempt
		m.recordMatchLog(message, llmResult, llmDuration, true)

		if err == nil && llmResult != nil && llmResult.Confidence >= m.threshold {
			return llmResult
		}
	}

	// No confident match
	// Record final no-match result (nil result, duration 0, llmUsed false)
	m.recordMatchLog(message, nil, 0, false)
	return nil
}

// LLMClassifierAdapter adapts LLMClient to the LLMClassifier interface.
type LLMClassifierAdapter struct {
	client *LLMClient
}

// NewLLMClassifierAdapter creates a new adapter.
func NewLLMClassifierAdapter(client *LLMClient) *LLMClassifierAdapter {
	return &LLMClassifierAdapter{client: client}
}

// Classify uses the LLM to classify the message into a skill.
func (a *LLMClassifierAdapter) Classify(ctx context.Context, message string, skills []*Skill) (string, error) {
	// Build a prompt that lists available skills
	var skillDescs []string
	for _, s := range skills {
		if !s.Enabled {
			continue
		}
		skillDescs = append(skillDescs, fmt.Sprintf("- %s: %s (intent: %s, keywords: %v)",
			s.Name, s.Description, s.Intent, s.Keywords))
	}

	prompt := fmt.Sprintf(`You are an intent classification assistant. Given a user message, select the most appropriate skill from the list below.

Available skills:
%s

User message: "%s"

Return a JSON object with exactly these fields:
{
  "skill": "<skill name>",
  "intent": "<intent string>",
  "confidence": <float between 0.0 and 1.0>
}

If no skill matches, return skill: "" and intent: "chat".`, strings.Join(skillDescs, "\n"), message)

	history := []ChatMessage{
		{Role: "user", Content: message},
	}

	response, err := a.client.Chat(ctx, prompt, history)
	if err != nil {
		return "", err
	}
	return response, nil
}

// ExtractParams extracts intent-specific parameters via LLM.
func (a *LLMClassifierAdapter) ExtractParams(ctx context.Context, message string, intent Intent) (map[string]string, error) {
	prompt := fmt.Sprintf(`Extract structured parameters from the user message for intent "%s".

User message: "%s"

Return a JSON object with the appropriate parameters for this intent type.`, intent, message)

	history := []ChatMessage{
		{Role: "user", Content: message},
	}

	response, err := a.client.Chat(ctx, prompt, history)
	if err != nil {
		return nil, err
	}

	var params map[string]string
	if err := json.Unmarshal([]byte(response), &params); err != nil {
		return nil, fmt.Errorf("parse LLM params response: %w", err)
	}
	return params, nil
}

// MatchLogBuffer is a ring buffer for storing match logs.
type MatchLogBuffer struct {
	mu    sync.RWMutex
	logs  []*MatchLog
	start int
	size  int
	cap   int
}

// NewMatchLogBuffer creates a new buffer with the given capacity.
func NewMatchLogBuffer(capacity int) *MatchLogBuffer {
	return &MatchLogBuffer{
		logs: make([]*MatchLog, capacity),
		cap:  capacity,
	}
}

// Add adds a log entry to the buffer.
func (b *MatchLogBuffer) Add(log *MatchLog) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cap == 0 {
		return
	}

	idx := (b.start + b.size) % b.cap
	b.logs[idx] = log

	if b.size < b.cap {
		b.size++
	} else {
		b.start = (b.start + 1) % b.cap
	}
}

// List returns the most recent logs, up to limit.
func (b *MatchLogBuffer) List(limit int) []*MatchLog {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if limit <= 0 || limit > b.size {
		limit = b.size
	}

	result := make([]*MatchLog, limit)
	for i := 0; i < limit; i++ {
		idx := (b.start + b.size - limit + i) % b.cap
		result[i] = b.logs[idx]
	}
	return result
}

// Clear removes all logs from the buffer.
func (b *MatchLogBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.start = 0
	b.size = 0
	// Keep slice allocated
}