---
sidebar_position: 2
title: ספקים
---

# ניהול ספקים

ספק מייצג הגדרה של נקודת קצה ל-API: כתובת בסיס, אסימון אימות, שם מודל ועוד.

## דוגמת הגדרה

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000"
      },
      "codex_env_vars": {
        "CODEX_SOME_VAR": "value"
      },
      "opencode_env_vars": {
        "OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": "64000"
      }
    }
  }
}
```

## משתני סביבה

לכל ספק אפשר להגדיר משתני סביבה נפרדים לכל כלי CLI:

### משתני סביבה נפוצים של Claude Code

| משתנה | תיאור |
|-------|-------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | מספר מרבי של אסימוני פלט |
| `MAX_THINKING_TOKENS` | תקציב חשיבה מורחבת |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | חלון ההקשר המרבי |
| `BASH_DEFAULT_TIMEOUT_MS` | זמן קצוב כברירת מחדל ל-Bash |
