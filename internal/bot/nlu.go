package bot

import (
	"context"
	"regexp"
	"strings"
)

// NLUParser parses user messages into intents.
type NLUParser struct {
	mentionKeywords []string
	commandPatterns []*commandPattern
	skillMatcher    *SkillMatcher // optional skill-based matcher
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

// NewNLUParserWithSkills creates a new NLU parser with skill-based matching.
func NewNLUParserWithSkills(mentionKeywords []string, skillMatcher *SkillMatcher) *NLUParser {
	if len(mentionKeywords) == 0 {
		mentionKeywords = []string{"@zen", "/zen", "zen"}
	}

	p := &NLUParser{
		mentionKeywords: mentionKeywords,
		skillMatcher:    skillMatcher,
	}
	p.initPatterns()
	return p
}

func (p *NLUParser) initPatterns() {
	p.commandPatterns = []*commandPattern{
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
		// persona set/show/clear
		{
			pattern: regexp.MustCompile(`(?i)^(?:persona|人设)\s+(?:clear|清除)$`),
			intent:  IntentPersona,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentPersona, Action: "clear"}
			},
		},
		{
			pattern: regexp.MustCompile(`(?i)^(?:persona|人设)\s+(.+)$`),
			intent:  IntentPersona,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentPersona, Action: "set", Task: strings.TrimSpace(m[1])}
			},
		},
		{
			pattern: regexp.MustCompile(`(?i)^(?:persona|人设)$`),
			intent:  IntentPersona,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentPersona, Action: "show"}
			},
		},
		// forget / clear history
		{
			pattern: regexp.MustCompile(`(?i)^(?:forget|忘记|清除记忆)$`),
			intent:  IntentForget,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentForget}
			},
		},
		// send <target> <task> - explicit send task syntax
		{
			pattern: regexp.MustCompile(`(?i)^send\s+(\S+)\s+(.+)$`),
			intent:  IntentSendTask,
			extract: func(m []string) *ParsedIntent {
				return &ParsedIntent{Intent: IntentSendTask, Target: m[1], Task: m[2]}
			},
		},
		// <target>: <task> - colon syntax for send task
		{
			pattern: regexp.MustCompile(`(?i)^(\S+):\s+(.+)$`),
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
		return &ParsedIntent{Intent: IntentChat, Raw: msg.Content}
	}

	// Try command patterns first (no LLM cost)
	for _, cp := range p.commandPatterns {
		if matches := cp.pattern.FindStringSubmatch(content); matches != nil {
			intent := cp.extract(matches)
			intent.Raw = msg.Content
			return intent
		}
	}

	// Try skill-based matching if available
	if p.skillMatcher != nil {
		result := p.skillMatcher.Match(context.Background(), content)
		if result != nil && result.Confidence >= 0.7 { // use threshold from matcher
			// Convert MatchResult to ParsedIntent
			parsed := &ParsedIntent{
				Intent: result.Intent,
				Raw:    msg.Content,
				Task:   content, // preserve original message for parameter extraction
			}
			// Note: parameter extraction will be handled by handlers.go (T023)
			return parsed
		}
	}

	// If no pattern matched, treat as chat (conversational fallback)
	return &ParsedIntent{
		Intent: IntentChat,
		Task:   content,
		Raw:    msg.Content,
	}
}
