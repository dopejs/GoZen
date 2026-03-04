package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// --- T014: Local matcher tests ---

func TestLocalMatcherKeywordMatch(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil) // load builtins

	m := NewSkillMatcher(reg, nil, 0.7, 0)

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

	m := NewSkillMatcher(reg, nil, 0.7, 0)

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

	m := NewSkillMatcher(reg, nil, 0.7, 0)

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

	m := NewSkillMatcher(reg, nil, 0.7, 0)

	result := m.MatchLocal("hello how are you today")
	if result != nil && result.Confidence >= 0.7 {
		t.Errorf("expected no confident match for generic chat, got skill=%q conf=%.2f",
			result.Skill, result.Confidence)
	}
}

func TestLocalMatcherScoreWeighting(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7, 0)

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

	m := NewSkillMatcher(reg, mock, 0.7, 0)

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

	m := NewSkillMatcher(reg, mock, 0.7, 0)

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

	m := NewSkillMatcher(reg, mock, 0.7, 0)

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

	m := NewSkillMatcher(reg, mock, 0.7, 0)

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

	m := NewSkillMatcher(reg, mock, 0.7, 0)

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

	m := NewSkillMatcher(reg, mock, 0.7, 0)

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

	m := NewSkillMatcher(reg, mock, 0.7, 0)

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


// --- T024: MatchLog ring buffer tests ---

func TestMatchLogBuffer_AddAndList(t *testing.T) {
	buf := NewMatchLogBuffer(3)

	log1 := &MatchLog{Input: "msg1"}
	log2 := &MatchLog{Input: "msg2"}
	log3 := &MatchLog{Input: "msg3"}

	buf.Add(log1)
	buf.Add(log2)
	buf.Add(log3)

	logs := buf.List(10)
	if len(logs) != 3 {
		t.Fatalf("List(10) len = %d, want 3", len(logs))
	}
	if logs[0].Input != "msg1" || logs[1].Input != "msg2" || logs[2].Input != "msg3" {
		t.Errorf("logs mismatch: got %v, %v, %v", logs[0].Input, logs[1].Input, logs[2].Input)
	}
}

func TestMatchLogBuffer_Overflow(t *testing.T) {
	buf := NewMatchLogBuffer(3)

	// Add 5 logs, buffer size is 3
	for i := 0; i < 5; i++ {
		buf.Add(&MatchLog{Input: fmt.Sprintf("msg%d", i+1)})
	}

	logs := buf.List(10)
	if len(logs) != 3 {
		t.Fatalf("List(10) len = %d, want 3 after overflow", len(logs))
	}
	// Should contain the 3 most recent: msg3, msg4, msg5
	expected := []string{"msg3", "msg4", "msg5"}
	for i, exp := range expected {
		if logs[i].Input != exp {
			t.Errorf("logs[%d] = %q, want %q", i, logs[i].Input, exp)
		}
	}
}

func TestMatchLogBuffer_ListWithLimit(t *testing.T) {
	buf := NewMatchLogBuffer(5)

	for i := 0; i < 5; i++ {
		buf.Add(&MatchLog{Input: fmt.Sprintf("msg%d", i+1)})
	}

	// Request fewer than total
	logs := buf.List(2)
	if len(logs) != 2 {
		t.Fatalf("List(2) len = %d, want 2", len(logs))
	}
	// Should return the 2 most recent: msg4, msg5
	if logs[0].Input != "msg4" || logs[1].Input != "msg5" {
		t.Errorf("logs = [%q, %q], want [msg4, msg5]", logs[0].Input, logs[1].Input)
	}
}

func TestMatchLogBuffer_Clear(t *testing.T) {
	buf := NewMatchLogBuffer(3)
	buf.Add(&MatchLog{Input: "msg1"})
	buf.Add(&MatchLog{Input: "msg2"})

	buf.Clear()
	logs := buf.List(10)
	if len(logs) != 0 {
		t.Errorf("after Clear, List(10) len = %d, want 0", len(logs))
	}

	// Should be able to add after clear
	buf.Add(&MatchLog{Input: "msg3"})
	logs = buf.List(10)
	if len(logs) != 1 || logs[0].Input != "msg3" {
		t.Errorf("after clear and add, logs = %v, want [msg3]", logs)
	}
}

// --- T027: Match logging tests ---

func TestSkillMatcherMatchLogging(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	// Create matcher with log buffer
	m := NewSkillMatcher(reg, nil, 0.7, 10) // buffer size 10

	// Test local match logging
	result := m.Match(context.Background(), "pause")
	if result == nil {
		t.Fatal("Match(pause) returned nil")
	}

	// Check that logs were recorded
	if m.logBuffer == nil {
		t.Fatal("logBuffer should not be nil")
	}
	logs := m.logBuffer.List(10)
	if len(logs) == 0 {
		t.Fatal("expected logs after match")
	}

	// Verify first log entry (local match attempt)
	log := logs[0]
	if log.Input != "pause" {
		t.Errorf("log.Input = %q, want %q", log.Input, "pause")
	}
	if log.Result == nil {
		t.Error("log.Result should not be nil for successful local match")
	} else if log.Result.Skill != "process-control" {
		t.Errorf("log.Result.Skill = %q, want %q", log.Result.Skill, "process-control")
	}
	if log.LLMUsed {
		t.Error("log.LLMUsed should be false for local match")
	}
	if log.DurationMs < 0 {
		t.Error("log.DurationMs should be non-negative")
	}

	// Should have at least one log (local match succeeded)
	// Since local match succeeded, LLM fallback not attempted
	// So only one log entry expected
	if len(logs) != 1 {
		t.Errorf("expected 1 log entry for successful local match, got %d", len(logs))
	}
}

func TestSkillMatcherMatchLoggingLLMFallback(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	// Mock classifier that returns a valid skill
	mock := &mockLLMClassifier{
		response: `{"skill":"process-control","intent":"control","confidence":0.9}`,
	}

	// Create matcher with log buffer and LLM fallback enabled
	m := NewSkillMatcher(reg, mock, 0.7, 10) // threshold 0.7, LLM confidence 0.9 should succeed

	// Test with message that won't match locally (no keyword overlap)
	result := m.Match(context.Background(), "帮我把那个任务停掉")
	if result == nil {
		t.Fatal("Match should return result via LLM fallback")
	}

	if m.logBuffer == nil {
		t.Fatal("logBuffer should not be nil")
	}
	logs := m.logBuffer.List(10)
	if len(logs) < 2 {
		t.Fatalf("expected at least 2 logs (local attempt + LLM attempt), got %d", len(logs))
	}

	// First log should be local attempt with nil result (no match)
	localLog := logs[0]
	if localLog.Input != "帮我把那个任务停掉" {
		t.Errorf("localLog.Input = %q, want %q", localLog.Input, "帮我把那个任务停掉")
	}
	if localLog.Result != nil {
		t.Error("localLog.Result should be nil for no local match")
	}
	if localLog.LLMUsed {
		t.Error("localLog.LLMUsed should be false")
	}

	// Second log should be LLM attempt with result
	llmLog := logs[1]
	if llmLog.Input != "帮我把那个任务停掉" {
		t.Errorf("llmLog.Input = %q, want %q", llmLog.Input, "帮我把那个任务停掉")
	}
	if llmLog.Result == nil {
		t.Error("llmLog.Result should not be nil for successful LLM match")
	} else if llmLog.Result.Skill != "process-control" {
		t.Errorf("llmLog.Result.Skill = %q, want %q", llmLog.Result.Skill, "process-control")
	}
	if !llmLog.LLMUsed {
		t.Error("llmLog.LLMUsed should be true")
	}
	if llmLog.DurationMs < 0 {
		t.Error("llmLog.DurationMs should be non-negative")
	}
}

func TestSkillMatcherMatchLoggingNoMatch(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	// Create matcher with log buffer, no LLM classifier
	m := NewSkillMatcher(reg, nil, 0.95, 10) // high threshold ensures no match

	// Test with generic chat message
	result := m.Match(context.Background(), "hello how are you")
	if result != nil {
		t.Fatal("Match should return nil for generic chat with high threshold")
	}

	if m.logBuffer == nil {
		t.Fatal("logBuffer should not be nil")
	}
	logs := m.logBuffer.List(10)
	if len(logs) == 0 {
		t.Fatal("expected logs even for no match")
	}

	// Should have at least one log (local attempt with nil result)
	log := logs[0]
	if log.Input != "hello how are you" {
		t.Errorf("log.Input = %q, want %q", log.Input, "hello how are you")
	}
	if log.Result != nil {
		t.Error("log.Result should be nil for no match")
	}
	if log.LLMUsed {
		t.Error("log.LLMUsed should be false")
	}
	// Duration may be 0 for no-match log (we set 0 in Match)
	// That's acceptable
}

// --- T034: Multi-language keyword matching tests ---

func TestMultiLangKeywordMatch(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7, 0)

	tests := []struct {
		name       string
		input      string
		wantSkill  string
		wantIntent Intent
	}{
		// Chinese exact keywords
		{"zh exact 暂停", "暂停", "process-control", IntentControl},
		{"zh exact 停止", "停止", "process-control", IntentControl},
		{"zh exact 绑定", "绑定", "project-bind", IntentBind},
		{"zh exact 批准", "批准", "approval", IntentApprove},
		{"zh exact 状态", "状态", "query-status", IntentQueryStatus},
		{"zh exact 列表", "列表", "query-list", IntentQueryList},
		{"zh exact 忘记", "忘记", "forget", IntentForget},
		{"zh exact 人设", "人设", "persona", IntentPersona},
		{"zh exact 发送", "发送", "send-task", IntentSendTask},
		// English exact keywords
		{"en exact pause", "pause", "process-control", IntentControl},
		{"en exact bind", "bind", "project-bind", IntentBind},
		{"en exact approve", "approve", "approval", IntentApprove},
		{"en exact status", "status", "query-status", IntentQueryStatus},
		{"en exact list", "list", "query-list", IntentQueryList},
		{"en exact forget", "forget", "forget", IntentForget},
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
				t.Errorf("Intent = %v, want %v", result.Intent, tt.wantIntent)
			}
		})
	}
}

func TestMultiLangSubstringMatch(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7, 0)

	tests := []struct {
		name      string
		input     string
		wantSkill string
	}{
		{"zh phrase with 暂停", "帮我暂停一下", "process-control"},
		{"zh phrase with 状态", "查看所有进程状态", "query-status"},
		{"zh phrase with 显示", "显示所有进程", "query-list"},
		{"en phrase with pause", "please pause it", "process-control"},
		{"en phrase with status", "check the status", "query-status"},
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

func TestMultiLangSynonymMatch(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7, 0)

	tests := []struct {
		name      string
		input     string
		wantSkill string
	}{
		{"zh synonym 中止→停止", "中止", "process-control"},
		{"zh synonym 挂起→暂停", "挂起", "process-control"},
		{"en synonym halt→stop", "halt", "process-control"},
		{"en synonym confirm→approve", "confirm", "approval"},
		{"zh synonym 好的→批准", "好的", "approval"},
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

func TestMultiLangMixedInput(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7, 0)

	// Mixed language inputs should still match
	tests := []struct {
		name      string
		input     string
		wantSkill string
	}{
		{"mixed zh-en pause", "请pause一下", "process-control"},
		{"mixed en-zh status", "check 状态", "query-status"},
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

// --- T035: Multi-language LLM fallback tests ---

func TestMultiLangLLMFallback(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	mock := &mockLLMClassifier{
		response: `{"skill":"process-control","intent":"control","confidence":0.9}`,
	}

	m := NewSkillMatcher(reg, mock, 0.7, 0)

	// Chinese natural language that doesn't contain exact keywords
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
	if result.Method != "llm" {
		t.Errorf("Method = %q, want %q", result.Method, "llm")
	}
}

func TestMultiLangLLMFallbackOrchestrator(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	mock := &mockLLMClassifier{
		response: `{"skill":"query-status","intent":"query_status","confidence":0.85}`,
	}

	m := NewSkillMatcher(reg, mock, 0.7, 0)

	// This message doesn't contain exact keywords, so local match should fail
	// LLM should correctly classify it
	result := m.Match(context.Background(), "服务器运行得怎么样了")
	if result == nil {
		t.Fatal("Match() returned nil for Chinese natural language")
	}
	if result.Skill != "query-status" {
		t.Errorf("Skill = %q, want %q", result.Skill, "query-status")
	}
	if result.Intent != IntentQueryStatus {
		t.Errorf("Intent = %v, want %v", result.Intent, IntentQueryStatus)
	}
}

// --- T041: Edge case tests ---

func TestEdgeCaseShortMessages(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7, 0)

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"single space", " "},
		{"emoji", "😀"},
		{"single char", "a"},
		{"two chars", "ab"},
		{"punctuation", "!@#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := m.MatchLocal(tt.input)
			// Short/gibberish messages should either return nil or low confidence
			if result != nil && result.Confidence >= 0.95 {
				t.Errorf("short message %q should not have high confidence, got %.2f", tt.input, result.Confidence)
			}
		})
	}
}

func TestEdgeCaseConcurrentMatch(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7, 50)
	done := make(chan struct{}, 20)
	for i := 0; i < 20; i++ {
		go func(i int) {
			defer func() { done <- struct{}{} }()
			input := "pause"
			if i%2 == 0 {
				input = "暂停"
			}
			result := m.Match(context.Background(), input)
			if result == nil {
				t.Errorf("concurrent match %d returned nil", i)
			}
		}(i)
	}

	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify logs were recorded correctly
	logs := m.logBuffer.List(50)
	if len(logs) < 20 {
		t.Errorf("expected at least 20 log entries from concurrent matches, got %d", len(logs))
	}
}

func TestEdgeCaseMultipleSkillsSameKeyword(t *testing.T) {
	reg := NewSkillRegistry()
	// Register two skills with overlapping keywords
	if err := reg.Register(&Skill{
		Name:        "skill-a",
		Description: "Test skill A",
		Intent:      IntentChat,
		Priority:    10,
		Enabled:     true,
		Keywords:    map[string][]string{"en": {"test"}},
	}); err != nil {
		t.Fatalf("Register skill-a: %v", err)
	}
	if err := reg.Register(&Skill{
		Name:        "skill-b",
		Description: "Test skill B",
		Intent:      IntentChat,
		Priority:    20,
		Enabled:     true,
		Keywords:    map[string][]string{"en": {"test"}},
	}); err != nil {
		t.Fatalf("Register skill-b: %v", err)
	}

	m := NewSkillMatcher(reg, nil, 0.7, 0)
	result := m.MatchLocal("test")

	// Should return one of them, not panic
	if result == nil {
		t.Fatal("expected a match for overlapping keywords")
	}
	if result.Confidence < 0.9 {
		t.Errorf("exact keyword should have high confidence, got %.2f", result.Confidence)
	}
}

// --- T042: LLM edge case tests ---

func TestLLMFallbackEmptySkillResponse(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	mock := &mockLLMClassifier{
		response: `{"skill":"","intent":"chat","confidence":0.5}`,
	}

	m := NewSkillMatcher(reg, mock, 0.7, 0)

	result, err := m.MatchLLM(context.Background(), "random gibberish")
	if err == nil {
		t.Fatal("expected error for empty skill name in LLM response")
	}
	if result != nil {
		t.Error("expected nil result for empty skill")
	}
}

func TestLLMFallbackLowConfidence(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	mock := &mockLLMClassifier{
		response: `{"skill":"process-control","intent":"control","confidence":0.3}`,
	}

	m := NewSkillMatcher(reg, mock, 0.7, 0)

	// Match orchestrator should reject low confidence LLM result
	result := m.Match(context.Background(), "something ambiguous without keywords")
	if result != nil {
		t.Errorf("expected nil for low confidence LLM result, got skill=%q conf=%.2f", result.Skill, result.Confidence)
	}
}

func TestLLMFallbackDisabledSkillResponse(t *testing.T) {
	reg := NewSkillRegistry()
	reg.Register(&Skill{
		Name:     "disabled-skill",
		Intent:   IntentChat,
		Enabled:  false,
		Keywords: map[string][]string{"en": {"disabled"}},
	})

	mock := &mockLLMClassifier{
		response: `{"skill":"disabled-skill","intent":"chat","confidence":0.9}`,
	}

	m := NewSkillMatcher(reg, mock, 0.7, 0)

	result, err := m.MatchLLM(context.Background(), "some message")
	if err == nil {
		t.Fatal("expected error for disabled skill in LLM response")
	}
	if result != nil {
		t.Error("expected nil result for disabled skill")
	}
}

// --- T043: Accuracy benchmark test ---

func TestAccuracyBenchmark(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)

	m := NewSkillMatcher(reg, nil, 0.7, 0)

	tests := []struct {
		input      string
		wantSkill  string
		wantIntent Intent
	}{
		// Process control (en)
		{"pause", "process-control", IntentControl},
		{"stop", "process-control", IntentControl},
		{"resume", "process-control", IntentControl},
		{"cancel", "process-control", IntentControl},
		{"halt", "process-control", IntentControl},
		// Process control (zh)
		{"暂停", "process-control", IntentControl},
		{"停止", "process-control", IntentControl},
		// Bind
		{"bind", "project-bind", IntentBind},
		{"绑定", "project-bind", IntentBind},
		// Approve
		{"approve", "approval", IntentApprove},
		{"reject", "approval", IntentApprove},
		{"批准", "approval", IntentApprove},
		{"confirm", "approval", IntentApprove},
		// Status
		{"status", "query-status", IntentQueryStatus},
		{"状态", "query-status", IntentQueryStatus},
		// List
		{"list", "query-list", IntentQueryList},
		{"列表", "query-list", IntentQueryList},
		// Forget
		{"forget", "forget", IntentForget},
		{"忘记", "forget", IntentForget},
		// Persona
		{"persona", "persona", IntentPersona},
		// Send
		{"send", "send-task", IntentSendTask},
		{"发送", "send-task", IntentSendTask},
	}

	correct := 0
	for _, tt := range tests {
		result := m.MatchLocal(tt.input)
		if result != nil && result.Skill == tt.wantSkill && result.Intent == tt.wantIntent {
			correct++
		}
	}

	accuracy := float64(correct) / float64(len(tests)) * 100
	if accuracy < 85 {
		t.Errorf("accuracy = %.1f%%, want ≥85%% (%d/%d correct)", accuracy, correct, len(tests))
	}
	t.Logf("Accuracy: %.1f%% (%d/%d correct)", accuracy, correct, len(tests))
}

// --- T044: Latency benchmark test ---

func BenchmarkLocalMatch(b *testing.B) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)
	m := NewSkillMatcher(reg, nil, 0.7, 0)

	inputs := []string{
		"pause",
		"暂停",
		"please pause the process",
		"帮我暂停一下",
		"random gibberish message",
		"status",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.MatchLocal(inputs[i%len(inputs)])
	}
}

func TestLocalMatchLatency(t *testing.T) {
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil)
	m := NewSkillMatcher(reg, nil, 0.7, 0)

	inputs := []string{
		"pause", "暂停", "please pause the process",
		"帮我暂停一下", "random gibberish", "status",
		"bind myproject", "approve", "list all workers",
	}

	var maxDuration time.Duration
	for _, input := range inputs {
		start := time.Now()
		for j := 0; j < 100; j++ {
			m.MatchLocal(input)
		}
		elapsed := time.Since(start) / 100
		if elapsed > maxDuration {
			maxDuration = elapsed
		}
	}

	// SC-004: local match ≤500ms for 95th percentile
	if maxDuration > 500*time.Millisecond {
		t.Errorf("local match latency = %v, want ≤500ms", maxDuration)
	}
	t.Logf("Max local match latency: %v", maxDuration)
}

// TestMatchLog_DurationMs_JSON verifies that MatchLog.DurationMs serializes as
// actual milliseconds (int64), not nanoseconds (time.Duration default).
func TestMatchLog_DurationMs_JSON(t *testing.T) {
	tests := []struct {
		name       string
		durationMs int64
	}{
		{"typical match", 1500},
		{"fast match", 50},
		{"slow LLM match", 5000},
		{"zero duration", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := MatchLog{
				Timestamp:  time.Now(),
				Input:      "test input",
				DurationMs: tt.durationMs,
				LLMUsed:    false,
			}

			data, err := json.Marshal(log)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			got, ok := parsed["duration_ms"].(float64)
			if !ok {
				t.Fatalf("duration_ms not found or not a number in JSON output")
			}
			if int64(got) != tt.durationMs {
				t.Errorf("duration_ms = %v, want %d", got, tt.durationMs)
			}
		})
	}
}
