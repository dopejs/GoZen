---
sidebar_position: 6
title: תמיכה בכמה כלי CLI
---

# תמיכה בכמה כלי CLI

GoZen תומך בשלושה כלי CLI לעוזרי תכנות מבוססי בינה מלאכותית:

| CLI | תיאור | תבנית API |
|-----|-------|-----------|
| `claude` | Claude Code (ברירת מחדל) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

## קביעת כלי ה-CLI שברירת המחדל

```bash
zen config default-client

# דרך ממשק הווב
zen web  # עמוד ההגדרות
```

## כלי CLI לכל פרויקט

```bash
cd ~/work/project
zen bind --cli codex  # התיקייה הזו משתמשת ב-Codex
```

## עקיפה זמנית של כלי ה-CLI

```bash
zen --cli opencode  # שימוש ב-OpenCode לסשן הנוכחי
```
