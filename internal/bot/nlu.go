package bot

import (
	"regexp"
	"strings"
)

// NLUParser parses user messages into intents.
type NLUParser struct {
	mentionKeywords []string
	commandPatterns []*commandPattern
}

type commandPattern struct {
	pattern *regexp.Regexp
	intent  Intent
	extract func(matches []string) *ParsedIntent
}

// NewNLUParser creates a new NLU parser.
func NewNLUParser(mentionKeywords []string) *NLUParser {
	if len(mentionKeywords) == 0 {
		mentionKeywords = []string{"@zen", "/zen", "zen"}
	}

	p := &NLUParser{
		mentionKeywords: mentionKeywords,
	}
	p.initPatterns()
	return p
}

func (p *NLUParser) initPatterns() {
	p.commandPatterns = []*commandPattern{
		// list command
		{
			pattern: regexp.MustCompile(`(?i)^list$`),
			intent:  IntentQueryList,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentQueryList}
			},
		},
		// status [target]
		{
			pattern: regexp.MustCompile(`(?i)^status(?:\s+(\S+))?$`),
			intent:  IntentQueryStatus,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentQueryStatus, Target: m[1]}
			},
		},
		// logs [n]
		{
			pattern: regexp.MustCompile(`(?i)^logs?(?:\s+(\d+))?$`),
			intent:  IntentQueryStatus,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentQueryStatus, Action: "logs", Params: map[string]string{"limit": m[1]}}
			},
		},
		// errors
		{
			pattern: regexp.MustCompile(`(?i)^errors?$`),
			intent:  IntentQueryStatus,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentQueryStatus, Action: "errors"}
			},
		},
		// pause/resume/cancel/stop [target]
		{
			pattern: regexp.MustCompile(`(?i)^(pause|resume|cancel|stop)(?:\s+(\S+))?$`),
			intent:  IntentControl,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentControl, Action: strings.ToLower(m[1]), Target: m[2]}
			},
		},
		// bind [target]
		{
			pattern: regexp.MustCompile(`(?i)^bind(?:\s+(\S+))?$`),
			intent:  IntentBind,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentBind, Target: m[1]}
			},
		},
		// help
		{
			pattern: regexp.MustCompile(`(?i)^help$`),
			intent:  IntentHelp,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentHelp}
			},
		},
		// approve/reject (for button clicks or replies)
		{
			pattern: regexp.MustCompile(`(?i)^(approve|yes|ok|批准|同意)$`),
			intent:  IntentApprove,
			extract: func(m []string) *ParsedIntent {
				approved := true
				return &ParsedIntent{Intent: IntentApprove, Approved: &approved}
			},
		},
		{
			pattern: regexp.MustCompile(`(?i)^(reject|no|deny|拒绝|否)$`),
			intent:  IntentApprove,
			extract: func(m []string) *ParsedIntent {
				approved := false
				return &ParsedIntent{Intent: IntentApprove, Approved: &approved}
			},
		},
		// [target] <task> - send task to specific target
		{
			pattern: regexp.MustCompile(`(?i)^(\S+)\s+(.+)$`),
			intent:  IntentSendTask,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentSendTask, Target: m[1], Task: m[2]}
			},
		},
	}
}

// Parse parses a message and returns the intent.
// Returns nil if the message should be ignored.
func (p *NLUParser) Parse(msg *Message, requireMention bool) *ParsedIntent {
	content := strings.TrimSpace(msg.Content)
	if content == "" {
		return nil
	}

	// Check for mention/command prefix
	hasMention := msg.IsMention || msg.IsDirectMsg
	var stripped string

	for _, kw := range p.mentionKeywords {
		lower := strings.ToLower(content)
		kwLower := strings.ToLower(kw)
		if strings.HasPrefix(lower, kwLower) {
			stripped = strings.TrimSpace(content[len(kw):])
			hasMention = true
			break
		}
	}

	// If mention required but not found, ignore
	if requireMention && !hasMention {
		return nil
	}

	// Use stripped content if we found a prefix, otherwise use original
	if stripped != "" {
		content = stripped
	}

	// Empty after stripping prefix means just a mention
	if content == "" {
		return &ParsedIntent{Intent: IntentHelp, Raw: msg.Content}
	}

	// Try command patterns first (no LLM cost)
	for _, cp := range p.commandPatterns {
		if matches := cp.pattern.FindStringSubmatch(content); matches != nil {
			intent := cp.extract(matches)
			intent.Raw = msg.Content
			return intent
		}
	}

	// If no pattern matched, treat as task for bound/default process
	return &ParsedIntent{
		Intent: IntentSendTask,
		Task:   content,
		Raw:    msg.Content,
	}
}

// ParseNaturalLanguage uses LLM to parse natural language.
// This is called when simple pattern matching fails and NLU is enabled.
func (p *NLUParser) ParseNaturalLanguage(content string, processes []string) *ParsedIntent {
	// Natural language patterns (no LLM, just heuristics)
	lower := strings.ToLower(content)

	// Status queries
	statusPatterns := []string{
		"状态", "怎么样", "在干嘛", "在做什么", "status", "what's",
		"看看", "查看", "检查",
	}
	for _, pat := range statusPatterns {
		if strings.Contains(lower, pat) {
			target := p.extractTarget(content, processes)
			return &ParsedIntent{Intent: IntentQueryStatus, Target: target, Raw: content}
		}
	}

	// List queries
	listPatterns := []string{
		"有哪些", "列出", "所有", "list", "哪些项目", "多少个",
	}
	for _, pat := range listPatterns {
		if strings.Contains(lower, pat) {
			return &ParsedIntent{Intent: IntentQueryList, Raw: content}
		}
	}

	// Control commands
	controlMap := map[string]string{
		"暂停": "pause", "停止": "stop", "继续": "resume", "取消": "cancel",
		"pause": "pause", "stop": "stop", "resume": "resume", "cancel": "cancel",
	}
	for cn, action := range controlMap {
		if strings.Contains(lower, cn) {
			target := p.extractTarget(content, processes)
			return &ParsedIntent{Intent: IntentControl, Action: action, Target: target, Raw: content}
		}
	}

	// Default: treat as task
	target := p.extractTarget(content, processes)
	return &ParsedIntent{Intent: IntentSendTask, Target: target, Task: content, Raw: content}
}

// extractTarget tries to find a process name in the content.
func (p *NLUParser) extractTarget(content string, processes []string) string {
	lower := strings.ToLower(content)
	for _, proc := range processes {
		if strings.Contains(lower, strings.ToLower(proc)) {
			return proc
		}
	}
	return ""
}
