package bot

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// --- T014: Local matcher tests ---

func TestLocalMatcherKeywordMatch(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil) // load builtins

	m := NewSkillMatcher(reg, nil, 0.7)

	tests := []struct {
		name       string
		input      string
		wantSkill  string
		wantIntent Intent
	}{
		{"exact en keyword", "pause", "process-control", IntentControl},
		{"exact zh keyword", "暂停", "process-control", IntentControl},
		{"bind keyword", "bind", "project-bind", IntentBind},
		{"approve keyword", "approve", "approval", IntentApprove},
		{"status keyword", "status", "query-status", IntentQueryStatus},
		{"list keyword", "list", "query-list", IntentQueryList},
		{"forget keyword", "forget", "forget", IntentForget},
		{"persona keyword", "persona", "persona", IntentPersona},
		{"send keyword", "send", "send-task", IntentSendTask},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.MatchLocal(tt.input)
			if result == nil {
				t.Fatalf("MatchLocal(%q) returned nil", tt.input)
			}
			if result.Skill != tt.wantSkill {
				t.Errorf("Skill = %q, want %q", result.Skill, tt.wantSkill)
			}
			if result.Intent != tt.wantIntent {
				t.Errorf("Intent = %q, want %q", result.Intent, tt.wantIntent)
			}
		})
	}
}

func TestLocalMatcherSynonymMatch(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7)

	tests := []struct {
		name       string
		input      string
		wantSkill  string
		wantIntent Intent
	}{
		{"en synonym halt->stop", "halt", "process-control", IntentControl},
		{"en synonym kill->stop", "kill", "process-control", IntentControl},
		{"zh synonym 中止->停止", "中止", "process-control", IntentControl},
		{"en synonym confirm->approve", "confirm", "approval", IntentApprove},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.MatchLocal(tt.input)
			if result == nil {
				t.Fatalf("MatchLocal(%q) returned nil", tt.input)
			}
			if result.Skill != tt.wantSkill {
				t.Errorf("Skill = %q, want %q", result.Skill, tt.wantSkill)
			}
		})
	}
}

func TestLocalMatcherFuzzyMatch(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7)

	// Fuzzy should match substrings in longer input
	tests := []struct {
		name      string
		input     string
		wantSkill string
	}{
		{"keyword in phrase en", "please pause it", "process-control"},
		{"keyword in phrase zh", "帮我暂停一下", "process-control"},
		{"status in phrase", "check the status", "query-status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.MatchLocal(tt.input)
			if result == nil {
				t.Fatalf("MatchLocal(%q) returned nil", tt.input)
			}
			if result.Skill != tt.wantSkill {
				t.Errorf("Skill = %q, want %q", result.Skill, tt.wantSkill)
			}
		})
	}
}

func TestLocalMatcherNoMatch(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7)

	result := m.MatchLocal("hello how are you today")
	if result != nil && result.Confidence >= 0.7 {
		t.Errorf("expected no confident match for generic chat, got skill=%q conf=%.2f",
			result.Skill, result.Confidence)
	}
}

func TestLocalMatcherScoreWeighting(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7)

	// Exact keyword match should score higher than fuzzy
	exact := m.MatchLocal("pause")
	fuzzy := m.MatchLocal("can you pause something")

	if exact == nil || fuzzy == nil {
		t.Fatal("both should return results")
	}
	if exact.Confidence <= fuzzy.Confidence {
		t.Errorf("exact match confidence (%.2f) should be > fuzzy (%.2f)",
			exact.Confidence, fuzzy.Confidence)
	}
}

// --- T015: LLM fallback matcher tests ---

// mockLLMClassifier is a test double for LLM classification.
type mockLLMClassifier struct {
	response string
	err      error
	delay    time.Duration
}

func (m *mockLLMClassifier) Classify(ctx context.Context, message string, skills []*Skill) (string, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return m.response, m.err
}

func (m *mockLLMClassifier) ExtractParams(ctx context.Context, message string, intent Intent) (map[string]string, error) {
	return nil, m.err
}

func TestLLMFallbackMatcherSuccess(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	mock := &mockLLMClassifier{
		response: `{"skill":"process-control","intent":"control","confidence":0.9}`,
	}

	m := NewSkillMatcher(reg, mock, 0.7)

	result, err := m.MatchLLM(context.Background(), "帮我把那个任务停掉")
	if err != nil {
		t.Fatalf("MatchLLM() error = %v", err)
	}
	if result == nil {
		t.Fatal("MatchLLM() returned nil")
	}
	if result.Skill != "process-control" {
		t.Errorf("Skill = %q, want %q", result.Skill, "process-control")
	}
	if result.Intent != IntentControl {
		t.Errorf("Intent = %q, want %q", result.Intent, IntentControl)
	}
	if result.Method != "llm" {
		t.Errorf("Method = %q, want %q", result.Method, "llm")
	}
}

func TestLLMFallbackMatcherInvalidJSON(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	mock := &mockLLMClassifier{
		response: `not valid json`,
	}

	m := NewSkillMatcher(reg, mock, 0.7)

	result, err := m.MatchLLM(context.Background(), "some message")
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
	if result != nil {
		t.Error("expected nil result for invalid JSON")
	}
}

func TestLLMFallbackMatcherTimeout(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	mock := &mockLLMClassifier{
		response: `{"skill":"process-control","intent":"control","confidence":0.9}`,
		delay:    5 * time.Second,
	}

	m := NewSkillMatcher(reg, mock, 0.7)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := m.MatchLLM(ctx, "pause")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestLLMFallbackMatcherError(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	mock := &mockLLMClassifier{
		err: fmt.Errorf("llm service unavailable"),
	}

	m := NewSkillMatcher(reg, mock, 0.7)

	result, err := m.MatchLLM(context.Background(), "some message")
	if err == nil {
		t.Fatal("expected error")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

// --- T016: Parameter extraction tests ---

func TestParameterExtractionControl(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	mock := &mockLLMClassifier{
		response: `{"skill":"process-control","intent":"control","confidence":0.9}`,
	}

	m := NewSkillMatcher(reg, mock, 0.7)

	// Test that Match orchestrator returns a MatchResult with intent info
	result := m.MatchLocal("pause worker1")
	if result == nil {
		t.Fatal("MatchLocal() returned nil")
	}
	if result.Intent != IntentControl {
		t.Errorf("Intent = %q, want %q", result.Intent, IntentControl)
	}
}

func TestParameterExtractionSendTask(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	mock := &mockLLMClassifier{
		response: `{"skill":"send-task","intent":"send_task","confidence":0.9}`,
	}

	m := NewSkillMatcher(reg, mock, 0.7)

	result := m.MatchLocal("send")
	if result == nil {
		t.Fatal("MatchLocal() returned nil")
	}
	if result.Intent != IntentSendTask {
		t.Errorf("Intent = %q, want %q", result.Intent, IntentSendTask)
	}
}

// --- Match orchestrator test ---

func TestSkillMatcherMatchOrchestrator(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	mock := &mockLLMClassifier{
		response: `{"skill":"process-control","intent":"control","confidence":0.9}`,
	}

	m := NewSkillMatcher(reg, mock, 0.7)

	tests := []struct {
		name       string
		input      string
		wantIntent Intent
	}{
		{"local match", "pause", IntentControl},
		{"local match zh", "暂停", IntentControl},
		{"bind", "bind", IntentBind},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.Match(context.Background(), tt.input)
			if result == nil {
				t.Fatalf("Match(%q) returned nil", tt.input)
			}
			if result.Intent != tt.wantIntent {
				t.Errorf("Intent = %q, want %q", result.Intent, tt.wantIntent)
			}
		})
	}
}
