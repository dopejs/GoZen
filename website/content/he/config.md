---
sidebar_position: 8
title: מדריך התצורה
---

# מדריך התצורה

## מיקומי קבצים

| קובץ | תיאור |
|------|-------|
| `~/.zen/zen.json` | קובץ התצורה הראשי |
| `~/.zen/zend.log` | יומן הדימון |
| `~/.zen/zend.pid` | קובץ ה-PID של הדימון |
| `~/.zen/logs.db` | מסד יומני הבקשות (SQLite) |

## דוגמת תצורה מלאה

```json
{
  "version": 7,
  "default_profile": "default",
  "default_client": "claude",
  "proxy_port": 19841,
  "web_port": 19840,
  "providers": {
    "anthropic": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000"
      }
    }
  },
  "profiles": {
    "default": {
      "providers": ["anthropic"]
    }
  },
  "project_bindings": {
    "/path/to/project": {
      "profile": "work",
      "client": "codex"
    }
  }
}
```

## מדריך השדות

| שדה | תיאור |
|-------|-------|
| `version` | מספר גרסת קובץ התצורה |
| `default_profile` | שם פרופיל ברירת המחדל |
| `default_client` | לקוח ה-CLI שברירת המחדל (claude/codex/opencode) |
| `proxy_port` | הפורט של שרת הפרוקסי (ברירת מחדל: 19841) |
| `web_port` | הפורט של ממשק הניהול (ברירת מחדל: 19840) |
| `providers` | אוסף הגדרות הספקים |
| `profiles` | אוסף הגדרות הפרופילים |
| `project_bindings` | הגדרות קישורי הפרויקטים |
| `sync` | הגדרות סנכרון התצורה (אופציונלי) |
