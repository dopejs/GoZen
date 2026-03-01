# API Contract: Bot Skills

**Feature**: 003-skill-intent-recognition
**Base Path**: `/api/v1/bot/skills`

## Endpoints

### GET /api/v1/bot/skills

获取所有已注册 Skill（内置 + 自定义）。

**Response 200**:
```json
{
  "skills": [
    {
      "name": "process-control",
      "description": "控制进程的启停（pause/resume/cancel/stop）",
      "intent": "control",
      "priority": 10,
      "enabled": true,
      "builtin": true,
      "keywords": {
        "en": ["pause", "resume", "stop", "cancel"],
        "zh": ["暂停", "恢复", "停止", "取消"]
      },
      "stats": {
        "total_matches": 42,
        "local_matches": 38,
        "llm_matches": 4
      }
    }
  ],
  "config": {
    "enabled": true,
    "confidence_threshold": 0.7,
    "llm_fallback": true,
    "log_buffer_size": 200
  }
}
```

### PUT /api/v1/bot/skills/config

更新 Skill 系统全局配置。

**Request**:
```json
{
  "enabled": true,
  "confidence_threshold": 0.75,
  "llm_fallback": true,
  "log_buffer_size": 500
}
```

**Response 200**: 返回更新后的 config 对象

### POST /api/v1/bot/skills

创建自定义 Skill。

**Request**:
```json
{
  "name": "code-review",
  "description": "请求代码审查",
  "intent": "send_task",
  "priority": 50,
  "keywords": {
    "en": ["review", "code review", "check code"],
    "zh": ["审查", "代码审查", "看看代码"]
  },
  "synonyms": {"审核": "审查", "检查": "审查"},
  "examples": ["帮我审查一下这段代码", "review my latest commit"]
}
```

**Response 201**: 返回创建的 Skill 对象
**Response 400**: `{"error": "skill name already exists"}` 或验证错误
**Response 422**: `{"error": "invalid intent type: xxx"}`

### PUT /api/v1/bot/skills/{name}

更新自定义 Skill（内置 Skill 仅可修改 enabled 字段）。

**Response 200**: 返回更新后的 Skill 对象
**Response 403**: `{"error": "cannot modify builtin skill fields"}` （修改内置 Skill 非 enabled 字段时）
**Response 404**: `{"error": "skill not found"}`

### DELETE /api/v1/bot/skills/{name}

删除自定义 Skill。

**Response 204**: 删除成功
**Response 403**: `{"error": "cannot delete builtin skill"}`
**Response 404**: `{"error": "skill not found"}`

### POST /api/v1/bot/skills/test

测试意图匹配（调试用）。

**Request**:
```json
{
  "message": "帮我把那个任务停掉"
}
```

**Response 200**:
```json
{
  "result": {
    "skill": "process-control",
    "intent": "control",
    "confidence": 0.92,
    "method": "local"
  },
  "scores": [
    {
      "skill_name": "process-control",
      "keyword_score": 0.8,
      "synonym_score": 0.95,
      "fuzzy_score": 0.7,
      "local_score": 0.87,
      "llm_score": 0,
      "final_score": 0.92
    }
  ],
  "duration_ms": 12,
  "llm_used": false
}
```

### GET /api/v1/bot/skills/logs

获取最近的匹配日志。

**Query Parameters**:
- `limit` (int, optional, default 50, max 200)

**Response 200**:
```json
{
  "logs": [
    {
      "timestamp": "2026-02-28T10:30:00Z",
      "input": "帮我把任务停掉",
      "platform": "telegram",
      "result": {
        "skill": "process-control",
        "intent": "control",
        "confidence": 0.92,
        "method": "local"
      },
      "duration_ms": 12,
      "llm_used": false
    }
  ],
  "total": 42
}
```
