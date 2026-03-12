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

	// "help" now falls through to chat intent (no longer a command)
	msg := &Message{Content: "help", IsMention: true}
	result := parser.Parse(msg, false)
	if result == nil {
		t.Error("Parse(help) returned nil")
		return
	}
	// "help" is now treated as chat since the help pattern was removed
	if result.Intent != IntentChat {
		t.Errorf("Parse(help) = %v, want %v", result.Intent, IntentChat)
	}
}

func TestNLUParser_Parse_List(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	// Task/process query words should fall through to IntentChat (LLM handles intent)
	tests := []struct {
		content string
	}{
		{"list"},
		{"任务"},
		{"tasks"},
		{"进程"},
		{"status"},
		{"状态"},
		{"有哪些任务"},
		{"现在有哪些"},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != IntentChat {
			t.Errorf("Parse(%q) = %v, want %v", tt.content, result.Intent, IntentChat)
		}
	}
}

func TestNLUParser_Parse_Status_WithTarget(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	// "status <target>" falls through to chat intent (handled by LLM)
	tests := []struct {
		content string
	}{
		{"status api"},
		{"status myproject"},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != IntentChat {
			t.Errorf("Parse(%q) intent = %v, want %v", tt.content, result.Intent, IntentChat)
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
		// Explicit "send <target> <task>" syntax
		{"send api run tests", IntentSendTask, "api", "run tests"},
		{"send myproject build", IntentSendTask, "myproject", "build"},
		// Colon syntax "<target>: <task>"
		{"api: run tests", IntentSendTask, "api", "run tests"},
		{"myproject: build and deploy", IntentSendTask, "myproject", "build and deploy"},
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
	// "help" now returns IntentChat since help pattern was removed
	msg := &Message{Content: "help", IsMention: false, IsDirectMsg: true}
	result := parser.Parse(msg, true)
	if result == nil {
		t.Fatal("Parse should not return nil for direct messages even when mention required")
	}
	if result.Intent != IntentChat {
		t.Errorf("Parse(help) = %v, want %v", result.Intent, IntentChat)
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
	// Now this falls through to IntentChat
	msg := &Message{Content: "@zen", IsMention: false}
	result := parser.Parse(msg, false)
	if result == nil {
		t.Fatal("Parse should not return nil for mention-only")
	}
	// Current behavior: "@zen" now falls through to chat intent
	if result.Intent != IntentChat {
		t.Errorf("Parse(@zen) intent = %v, want %v", result.Intent, IntentChat)
	}
}

func TestNLUParser_Parse_MentionWithCommand(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	// Mention followed by command - "help" now returns IntentChat
	msg := &Message{Content: "@zen help", IsMention: false}
	result := parser.Parse(msg, false)
	if result == nil {
		t.Fatal("Parse should not return nil")
	}
	if result.Intent != IntentChat {
		t.Errorf("Parse('@zen help') intent = %v, want %v", result.Intent, IntentChat)
	}
}

func TestNLUParser_Parse_Logs(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	// "logs" now falls through to chat intent (handled by LLM)
	tests := []string{"logs", "log", "logs 10"}

	for _, content := range tests {
		msg := &Message{Content: content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", content)
			continue
		}
		if result.Intent != IntentChat {
			t.Errorf("Parse(%q) intent = %v, want %v", content, result.Intent, IntentChat)
		}
	}
}

func TestNLUParser_Parse_Errors(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	// "errors" now falls through to chat intent (handled by LLM)
	msg := &Message{Content: "errors", IsMention: true}
	result := parser.Parse(msg, false)
	if result == nil {
		t.Fatal("Parse(errors) returned nil")
	}
	if result.Intent != IntentChat {
		t.Errorf("Parse(errors) intent = %v, want %v", result.Intent, IntentChat)
	}
}

func TestNLUParser_Parse_Persona(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	tests := []struct {
		content string
		action  string
		task    string
	}{
		{"persona", "show", ""},
		{"人设", "show", ""},
		{"persona 你是一个猫娘", "set", "你是一个猫娘"},
		{"人设 活泼可爱的助手", "set", "活泼可爱的助手"},
		{"persona clear", "clear", ""},
		{"人设 清除", "clear", ""},
	}

	for _, tt := range tests {
		msg := &Message{Content: tt.content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", tt.content)
			continue
		}
		if result.Intent != IntentPersona {
			t.Errorf("Parse(%q) intent = %v, want %v", tt.content, result.Intent, IntentPersona)
		}
		if result.Action != tt.action {
			t.Errorf("Parse(%q) action = %q, want %q", tt.content, result.Action, tt.action)
		}
		if result.Task != tt.task {
			t.Errorf("Parse(%q) task = %q, want %q", tt.content, result.Task, tt.task)
		}
	}
}

func TestNLUParser_Parse_Forget(t *testing.T) {
	parser := NewNLUParser([]string{"@zen"})

	tests := []string{"forget", "忘记", "清除记忆"}

	for _, content := range tests {
		msg := &Message{Content: content, IsMention: true}
		result := parser.Parse(msg, false)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", content)
			continue
		}
		if result.Intent != IntentForget {
			t.Errorf("Parse(%q) intent = %v, want %v", content, result.Intent, IntentForget)
		}
	}
}

func TestNLUParser_Parse_SkillIntegration(t *testing.T) {
	// Create skill registry with builtin skills
	reg := NewSkillRegistry()
	reg.LoadFromConfig(nil) // load builtins

	// Create skill matcher with no LLM fallback for tests
	matcher := NewSkillMatcher(reg, nil, 0.7, 0)

	// Create parser with skill matcher
	parser := NewNLUParserWithSkills([]string{"@zen"}, matcher)

	tests := []struct {
		name       string
		content    string
		wantIntent Intent
	}{
		// Regex still wins for exact commands
		{"regex exact pause", "pause", IntentControl},
		{"regex exact bind", "bind", IntentBind},
		{"regex exact approve", "approve", IntentApprove},
		// Natural language via skill matching
		{"skill natural language zh", "帮我暂停一下", IntentControl},
		{"skill natural language en", "can you pause it", IntentControl},
		{"skill status query", "查看所有进程状态", IntentQueryStatus},
		{"skill list query", "显示所有进程", IntentQueryList},
		{"skill forget natural", "忘记刚才的对话", IntentForget},
		// Fallback to chat
		{"chat fallback", "随便聊聊", IntentChat},
		{"chat generic", "hello how are you", IntentChat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{Content: tt.content, IsMention: true}
			result := parser.Parse(msg, false)
			if result == nil {
				t.Fatalf("Parse(%q) returned nil", tt.content)
			}
			if result.Intent != tt.wantIntent {
				t.Errorf("Parse(%q) intent = %v, want %v", tt.content, result.Intent, tt.wantIntent)
			}
		})
	}
}
