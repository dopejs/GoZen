package bot

import (
	"testing"
)

func TestNewNLUParser(t *testing.T) {
	keywords := []string{"@zen", "/zen"}
	parser := NewNLUParser(keywords)

	if parser == nil {
		t.Fatal("NewNLUParser returned nil")
	}
	if len(parser.mentionKeywords) != 2 {
		t.Errorf("expected 2 keywords, got %d", len(parser.mentionKeywords))
	}
}

func TestNLUParser_Parse_Help(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	tests := []struct {
		content string
		intent  Intent
	}{
		{"help", IntentHelp},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != tt.intent {
			t.Errorf("Parse(%q) = %v, want %v", tt.content, result.Intent, tt.intent)
		}
	}
}

func TestNLUParser_Parse_List(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	tests := []struct {
		content string
		intent  Intent
	}{
		{"list", IntentQueryList},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != tt.intent {
			t.Errorf("Parse(%q) = %v, want %v", tt.content, result.Intent, tt.intent)
		}
	}
}

func TestNLUParser_Parse_Status(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	tests := []struct {
		content string
		intent  Intent
		target  string
	}{
		{"status", IntentQueryStatus, ""},
		{"status api", IntentQueryStatus, "api"},
		{"status myproject", IntentQueryStatus, "myproject"},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != tt.intent {
			t.Errorf("Parse(%q) intent = %v, want %v", tt.content, result.Intent, tt.intent)
		}
		if result.Target != tt.target {
			t.Errorf("Parse(%q) target = %q, want %q", tt.content, result.Target, tt.target)
		}
	}
}

func TestNLUParser_Parse_Control(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	tests := []struct {
		content string
		intent  Intent
		action  string
	}{
		{"pause", IntentControl, "pause"},
		{"resume", IntentControl, "resume"},
		{"cancel", IntentControl, "cancel"},
		{"stop", IntentControl, "stop"},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != tt.intent {
			t.Errorf("Parse(%q) intent = %v, want %v", tt.content, result.Intent, tt.intent)
		}
		if result.Action != tt.action {
			t.Errorf("Parse(%q) action = %q, want %q", tt.content, result.Action, tt.action)
		}
	}
}

func TestNLUParser_Parse_Bind(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	tests := []struct {
		content string
		intent  Intent
		target  string
	}{
		{"bind api", IntentBind, "api"},
		{"bind myproject", IntentBind, "myproject"},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != tt.intent {
			t.Errorf("Parse(%q) intent = %v, want %v", tt.content, result.Intent, tt.intent)
		}
		if result.Target != tt.target {
			t.Errorf("Parse(%q) target = %q, want %q", tt.content, result.Target, tt.target)
		}
	}
}

func TestNLUParser_Parse_Approve(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	tests := []struct {
		content  string
		intent   Intent
		approved bool
	}{
		{"yes", IntentApprove, true},
		{"approve", IntentApprove, true},
		{"同意", IntentApprove, true},
		{"no", IntentApprove, false},
		{"reject", IntentApprove, false},
		{"拒绝", IntentApprove, false},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != tt.intent {
			t.Errorf("Parse(%q) intent = %v, want %v", tt.content, result.Intent, tt.intent)
		}
		if result.Approved == nil {
			t.Errorf("Parse(%q) approved is nil", tt.content)
			continue
		}
		if *result.Approved != tt.approved {
			t.Errorf("Parse(%q) approved = %v, want %v", tt.content, *result.Approved, tt.approved)
		}
	}
}

func TestNLUParser_Parse_SendTask(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	tests := []struct {
		content string
		intent  Intent
		target  string
		task    string
	}{
		{"api run tests", IntentSendTask, "api", "run tests"},
		{"myproject build", IntentSendTask, "myproject", "build"},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != tt.intent {
			t.Errorf("Parse(%q) intent = %v, want %v", tt.content, result.Intent, tt.intent)
		}
		if result.Target != tt.target {
			t.Errorf("Parse(%q) target = %q, want %q", tt.content, result.Target, tt.target)
		}
		if result.Task != tt.task {
			t.Errorf("Parse(%q) task = %q, want %q", tt.content, result.Task, tt.task)
		}
	}
}

func TestNLUParser_Parse_RequireMention(t *testing.T) {
	parser := NewNLUParser([]string{"@zen", "/zen"})

	// Without mention, requireMention=true should return nil
	msg := &Message{Content: "help", IsMention: false}
	result := parser.Parse(msg, true)
	if result != nil {
		t.Error("Parse should return nil when mention required but not present")
	}

	// With mention keyword in content
	msg = &Message{Content: "@zen help", IsMention: false}
	result = parser.Parse(msg, true)
	if result == nil {
		t.Error("Parse should not return nil when mention keyword is in content")
	}

	// With IsMention=true
	msg = &Message{Content: "help", IsMention: true}
	result = parser.Parse(msg, true)
	if result == nil {
		t.Error("Parse should not return nil when IsMention is true")
	}
}

func TestNLUParser_Parse_DirectMessage(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	// Direct messages should not require mention
	msg := &Message{Content: "help", IsMention: false, IsDirectMsg: true}
	result := parser.Parse(msg, true)
	if result == nil {
		t.Error("Parse should not return nil for direct messages even when mention required")
	}
}

func TestNLUParser_Parse_EmptyContent(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	msg := &Message{Content: "", IsMention: true}
	result := parser.Parse(msg, false)
	if result != nil {
		t.Error("Parse should return nil for empty content")
	}

	msg = &Message{Content: "   ", IsMention: true}
	result = parser.Parse(msg, false)
	if result != nil {
		t.Error("Parse should return nil for whitespace-only content")
	}
}

func TestNLUParser_Parse_MentionOnly(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	// When content is just "@zen", stripped becomes "", but content stays "@zen"
	// because the code only updates content when stripped != ""
	// This is current behavior - "@zen" gets treated as a task target
	msg := &Message{Content: "@zen", IsMention: false}
	result := parser.Parse(msg, false)
	if result == nil {
		t.Fatal("Parse should not return nil for mention-only")
	}
	// Current behavior: "@zen" matches the send_task pattern with target "@zen"
	if result.Intent != IntentSendTask {
		t.Errorf("Parse(@zen) intent = %v, want %v", result.Intent, IntentSendTask)
	}
}

func TestNLUParser_Parse_MentionWithCommand(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	// Mention followed by command
	msg := &Message{Content: "@zen help", IsMention: false}
	result := parser.Parse(msg, false)
	if result == nil {
		t.Fatal("Parse should not return nil")
	}
	if result.Intent != IntentHelp {
		t.Errorf("Parse('@zen help') intent = %v, want %v", result.Intent, IntentHelp)
	}
}

func TestNLUParser_ParseNaturalLanguage(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})
	processes := []string{"api", "backend", "frontend"}

	tests := []struct {
		content string
		intent  Intent
	}{
		{"api 状态怎么样", IntentQueryStatus},
		{"看看 backend 在干嘛", IntentQueryStatus},
		{"有哪些项目", IntentQueryList},
		{"列出所有进程", IntentQueryList},
		{"暂停 api", IntentControl},
		{"继续 backend", IntentControl},
	}

	for _, tt := range tests {
		result := parser.ParseNaturalLanguage(tt.content, processes)
		if result == nil {
			t.Errorf("ParseNaturalLanguage(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != tt.intent {
			t.Errorf("ParseNaturalLanguage(%q) = %v, want %v", tt.content, result.Intent, tt.intent)
		}
	}
}

func TestNLUParser_extractTarget(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})
	processes := []string{"api", "backend", "frontend"}

	tests := []struct {
		content  string
		expected string
	}{
		{"check api status", "api"},
		{"看看 backend 怎么样", "backend"},
		{"hello world", ""},
	}

	for _, tt := range tests {
		result := parser.extractTarget(tt.content, processes)
		if result != tt.expected {
			t.Errorf("extractTarget(%q) = %q, want %q", tt.content, result, tt.expected)
		}
	}
}

func TestNLUParser_Parse_Logs(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	tests := []struct {
		content string
		action  string
	}{
		{"logs", "logs"},
		{"log", "logs"},
		{"logs 10", "logs"},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Action != tt.action {
			t.Errorf("Parse(%q) action = %q, want %q", tt.content, result.Action, tt.action)
		}
	}
}

func TestNLUParser_Parse_Errors(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	msg := &Message{Content: "errors", IsMention: true}
	result := parser.Parse(msg, false)
	if result == nil {
		t.Fatal("Parse(errors) returned nil")
	}
	if result.Action != "errors" {
		t.Errorf("Parse(errors) action = %q, want 'errors'", result.Action)
	}
}
