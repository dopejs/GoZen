---
sidebar_position: 4
title: ניתוב לפי תרחיש
---

# ניתוב לפי תרחיש

ניתוב אוטומטי של בקשות לספקים שונים לפי מאפייני הבקשה.

## תרחישים נתמכים

| תרחיש | תיאור |
|-------|-------|
| `think` | מצב חשיבה מופעל |
| `image` | מכיל תוכן תמונה |
| `longContext` | התוכן חורג מהסף |
| `webSearch` | משתמש בכלי web_search |
| `background` | משתמש במודל Haiku |

## מנגנון נפילה לאחור

אם כל הספקים של תרחיש נכשלים, המערכת חוזרת אוטומטית לספקי ברירת המחדל של הפרופיל.

## דוגמת הגדרה

```json
{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}
```
