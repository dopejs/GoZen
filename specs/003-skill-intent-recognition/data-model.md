# Data Model: Skill-Based Intent Recognition

**Feature**: 003-skill-intent-recognition
**Date**: 2026-02-28

## Entities

### Skill

意图识别的基本单元，定义触发条件和关联意图。

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | 唯一标识符，如 "process-control" |
| description | string | yes | 人类可读描述，如 "控制进程的启停" |
| intent | Intent (string) | yes | 关联的意图类型，如 "control" |
| priority | int | yes | 匹配优先级，数字越小优先级越高（1-100） |
| enabled | bool | yes | 是否启用，默认 true |
| builtin | bool | yes | 是否为内置 Skill（内置不可删除） |
| keywords | map[string][]string | yes | 多语言关键词映射，key 为语言代码（"en"、"zh"），value 为关键词列表 |
| synonyms | map[string]string | no | 同义词映射，key 为变体词，value 为标准词 |
| examples | []string | no | 示例句子，用于 LLM 回退时的上下文 |

**Uniqueness**: name 字段全局唯一
**Lifecycle**: created → enabled/disabled → deleted（内置 Skill 不可删除，仅可禁用）

### MatchResult

单次意图匹配的结果。

| Field | Type | Description |
|-------|------|-------------|
| skill | string | 匹配的 Skill 名称（空字符串表示无匹配） |
| intent | Intent | 识别出的意图类型 |
| confidence | float64 | 置信度分数 (0.0 - 1.0) |
| method | string | 匹配方法："local" 或 "llm" |
| parsedIntent | *ParsedIntent | 最终构造的意图对象（含参数） |

### MatchLog

意图匹配的调试记录。

| Field | Type | Description |
|-------|------|-------------|
| timestamp | time.Time | 匹配时间 |
| input | string | 用户输入消息 |
| platform | string | 来源平台 |
| userID | string | 用户标识 |
| scores | []SkillScore | 各 Skill 的匹配得分 |
| result | MatchResult | 最终匹配结果 |
| duration | time.Duration | 匹配总耗时 |
| llmUsed | bool | 是否触发了 LLM 回退 |

### SkillScore

单个 Skill 的匹配得分明细。

| Field | Type | Description |
|-------|------|-------------|
| skillName | string | Skill 名称 |
| keywordScore | float64 | 关键词匹配得分 |
| synonymScore | float64 | 同义词匹配得分 |
| fuzzyScore | float64 | 模糊匹配得分 |
| localScore | float64 | 本地综合得分（加权） |
| llmScore | float64 | LLM 回退得分（未触发则为 0） |
| finalScore | float64 | 最终得分 |

## Config Changes

### SkillConfig (新增，嵌入 BotConfig)

```
BotConfig
└── Skills *SkillsConfig `json:"skills,omitempty"`
    ├── Enabled bool `json:"enabled"` (默认 true)
    ├── ConfidenceThreshold float64 `json:"confidence_threshold"` (默认 0.7)
    ├── LLMFallback bool `json:"llm_fallback"` (默认 true)
    ├── LogBufferSize int `json:"log_buffer_size"` (默认 200)
    └── Custom []SkillDefinition `json:"custom,omitempty"`
        ├── Name string
        ├── Description string
        ├── Intent string
        ├── Priority int
        ├── Keywords map[string][]string
        ├── Synonyms map[string]string
        └── Examples []string
```

**Config Version**: v9 → v10
**Migration**: 旧配置无 `skills` 字段时，自动初始化为默认值（enabled=true, 仅内置 Skill）

## State Transitions

```
Skill Lifecycle:
  [created] → [enabled] ⇄ [disabled] → [deleted]*
  * 仅用户自定义 Skill 可删除，内置 Skill 仅可禁用

Match Flow:
  [message received]
    → [regex match?] → YES → [use regex result]
    → NO → [local skill match]
      → [confidence ≥ threshold?] → YES → [use skill result]
      → NO → [LLM fallback enabled?]
        → YES → [LLM classify] → [confidence ≥ threshold?]
          → YES → [LLM parameter extract] → [use result]
          → NO → [fallback to IntentChat]
        → NO → [fallback to IntentChat]
```
