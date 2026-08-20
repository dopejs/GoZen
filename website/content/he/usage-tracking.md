---
sidebar_position: 11
title: מעקב שימוש ובקרת תקציב
---

# מעקב שימוש ובקרת תקציב

עקבו אחרי צריכת אסימונים ועלויות לפי ספקים, מודלים ופרויקטים. הגדירו תקרות הוצאה שנאכפות אוטומטית.

## יכולות

- **מעקב בזמן אמת** — צריכת אסימונים ועלות לכל בקשה
- **צבירה רב-ממדית** — לפי ספק, מודל, פרויקט ותקופה
- **תקרות תקציב** — מגבלות הוצאה יומיות, שבועיות וחודשיות
- **פעולות אוטומטיות** — אזהרה, הורדת דרג או חסימה בעת חריגה
- **הערכת עלות** — תמחור מדויק לכל המודלים המרכזיים
- **נתונים היסטוריים** — אחסון SQLite עם צבירה שעתית לשיפור הביצועים

## הגדרה

### הפעלת מעקב השימוש

```json
{
  "usage_tracking": {
    "enabled": true,
    "db_path": "~/.zen/usage.db"
  }
}
```

### הגדרת תמחור המודלים

```json
{
  "pricing": {
    "models": {
      "claude-opus-4": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet-4": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4o": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    },
    "model_families": {
      "claude-opus": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    }
  }
}
```

**התאמת מודלים**: תחילה מושווים שמות מודל מדויקים, ולאחר מכן קידומות של משפחת המודל.

### קביעת תקרות תקציב

```json
{
  "budget": {
    "daily": {
      "enabled": true,
      "limit": 10.0,
      "action": "warn"
    },
    "weekly": {
      "enabled": true,
      "limit": 50.0,
      "action": "downgrade"
    },
    "monthly": {
      "enabled": true,
      "limit": 200.0,
      "action": "block"
    }
  }
}
```

## פעולות תקציב

| פעולה | התנהגות |
|-------|---------|
| `warn` | רושם אזהרה ושולח התראת webhook, אך מאפשר את הבקשה |
| `downgrade` | מעבר למודל זול יותר (למשל opus ← sonnet ← haiku) |
| `block` | דוחה את הבקשה עם קוד 429 |

## ממשק ווב

לוח המחוונים לשימוש נמצא בכתובת `http://localhost:19840/usage`:

- **סקירה** — עלות כוללת, בקשות ואסימונים בתקופה הנוכחית
- **לפי ספק** — פילוח עלויות לכל ספק
- **לפי מודל** — נתוני שימוש לכל מודל
- **לפי פרויקט** — עלויות לכל פרויקט (דרך קישורי פרויקטים)
- **ציר זמן** — מגמות עלות שעתיות ויומיות
- **מצב התקציב** — חיווי חזותי למגבלות יומיות, שבועיות וחודשיות

## נקודות קצה ב-API

### קבלת סיכום שימוש

```bash
GET /api/v1/usage/summary?period=daily
```

תגובה:
```json
{
  "period": "daily",
  "start": "2026-03-05T00:00:00Z",
  "end": "2026-03-05T23:59:59Z",
  "total_cost": 8.45,
  "total_requests": 42,
  "total_input_tokens": 125000,
  "total_output_tokens": 35000,
  "by_provider": {
    "anthropic": 6.20,
    "openai": 2.25
  },
  "by_model": {
    "claude-sonnet-4": 5.10,
    "claude-opus-4": 1.10,
    "gpt-4o": 2.25
  }
}
```

### קבלת מצב התקציב

```bash
GET /api/v1/budget/status
```

תגובה:
```json
{
  "daily": {
    "enabled": true,
    "limit": 10.0,
    "spent": 8.45,
    "percent": 84.5,
    "action": "warn",
    "exceeded": false
  },
  "weekly": {
    "enabled": true,
    "limit": 50.0,
    "spent": 32.10,
    "percent": 64.2,
    "action": "downgrade",
    "exceeded": false
  },
  "monthly": {
    "enabled": true,
    "limit": 200.0,
    "spent": 145.80,
    "percent": 72.9,
    "action": "block",
    "exceeded": false
  }
}
```

### עדכון תקרות התקציב

```bash
PUT /api/v1/budget/limits
Content-Type: application/json

{
  "daily": {
    "enabled": true,
    "limit": 15.0,
    "action": "warn"
  }
}
```

## מעקב ברמת הפרויקט

עקבו אחרי עלויות לכל פרויקט בעזרת קישורי תיקיות:

```bash
# Bind current directory to a profile
zen bind work-profile

# All requests from this directory are tagged with the project path
# View costs in Web UI under "By Project"
```

## התראות webhook

קבלו התראה כשחורגים מהתקציב:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": ["budget_warning", "budget_exceeded"]
    }
  ]
}
```

ראו [Webhooks](./webhooks.md) להגדרה המלאה.

## המלצות

1. **התחילו באזהרות** — השתמשו תחילה ב-`warn` כדי להבין את דפוסי השימוש
2. **הגדירו מגבלות מציאותיות** — התבססו על נתוני שימוש היסטוריים
3. **הורדת דרג בפיתוח** — עברו אוטומטית למודלים זולים יותר בזמן בדיקות
4. **שמרו את החסימה לייצור** — השתמשו ב-`block` רק לתקרות הוצאה קשיחות
5. **בדקו מדי יום** — היכנסו ללוח המחוונים בקביעות כדי להימנע מהפתעות
6. **הפעילו webhooks** — קבלו התראות בזמן אמת עם ההתקרבות למגבלה

## פתרון תקלות

### השימוש אינו נרשם

1. ודאו ש-`usage_tracking.enabled` מוגדר `true` בתצורה
2. ודאו שנתיב מסד הנתונים ניתן לכתיבה: `~/.zen/usage.db`
3. הפעילו מחדש את הדימון: `zen daemon restart`

### עלויות שגויות

1. ודאו שתמחור המודלים בתצורה תואם למחירים הנוכחיים
2. בדקו את התאמת שם המודל (התאמה מדויקת מול קידומת משפחה)
3. עדכנו את התמחור אם הספקים שינו מחירים

### התקציב אינו נאכף

1. ודאו שהגדרת התקציב מופעלת
2. ודאו שנקבעה פעולה (`warn`, ‏`downgrade` או `block`)
3. בדקו ביומני הדימון אם יש שגיאות בבודק התקציב

## ביצועים

- **צבירה שעתית** — נתונים גולמיים נצברים כל שעה כדי להקל על השאילתות
- **שאילתות מאונדקסות** — אינדקסים לפי ספק, מודל, פרויקט וחותמת זמן
- **אחסון חסכוני** — כ-1KB לכל בקשה, כ-30MB ל-30,000 בקשות
- **לוח מחוונים מהיר** — זמני שאילתה מתחת לשנייה בשימוש טיפוסי
