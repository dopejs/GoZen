---
sidebar_position: 13
title: Webhooks
---

# Webhooks

קבלו התראות בזמן אמת — אזהרות תקציב, שינויי מצב של ספקים וסיכומים יומיים — דרך Slack, ‏Discord או webhook משלכם.

## יכולות

- **כמה תבניות** — ‏Slack, ‏Discord או JSON גנרי
- **סינון אירועים** — הירשמו רק לסוגי האירועים שמעניינים אתכם
- **כותרות מותאמות** — הוסיפו אימות או כותרות משלכם
- **שליחה אסינכרונית** — המסירה אינה חוסמת בקשות
- **עיצוב אוטומטי** — הודעות עשירות עם אימוג׳י וצבעים
- **בדיקה** — ודאו את ההגדרה לפני ההפעלה

## הגדרה

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": [
        "budget_warning",
        "budget_exceeded",
        "provider_down",
        "provider_up",
        "failover",
        "daily_summary"
      ],
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  ]
}
```

## סוגי אירועים

| אירוע | תיאור | מתי נשלח |
|-------|-------------|----------------|
| `budget_warning` | הגעה לסף התקציב | כשההוצאה מגיעה ל-80% מהמגבלה |
| `budget_exceeded` | חריגה מהתקציב | כשההוצאה עוברת את המגבלה שהוגדרה |
| `provider_down` | ספק הפך ללא תקין | כשאחוז ההצלחה יורד מתחת ל-70% |
| `provider_up` | ספק התאושש | כשספק לא תקין חוזר להיות תקין |
| `failover` | הבקשה עברה לגיבוי | כשבקשה עוברת לספק גיבוי |
| `daily_summary` | סיכום שימוש יומי | פעם ביום בחצות UTC |

## תבניות webhook

### Slack

מזוהה אוטומטית כשהכתובת מכילה `slack.com`.

**דוגמת הודעה:**
```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

**תבנית:**
```json
{
  "text": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)"
      }
    }
  ]
}
```

### Discord

מזוהה אוטומטית כשהכתובת מכילה `discord.com`.

**דוגמת embed:**
- **כותרת:** budget_warning
- **תיאור:** ⚠️ אזהרת תקציב: התקציב היומי ב-85.0% ‏($8.50 / $10.00)
- **צבע:** ענבר (#FBBF24)
- **חותמת זמן:** 2026-03-05T10:30:00Z

**תבנית:**
```json
{
  "content": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
  "embeds": [
    {
      "title": "budget_warning",
      "description": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
      "timestamp": "2026-03-05T10:30:00Z",
      "color": 16432932
    }
  ]
}
```

### JSON גנרי

משמש לכל שאר הכתובות.

**תבנית:**
```json
{
  "event": "budget_warning",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "period": "daily",
    "spent": 8.5,
    "limit": 10.0,
    "percent": 85.0,
    "project": ""
  }
}
```

## מבני נתוני האירועים

### אזהרת תקציב / חריגה

```json
{
  "event": "budget_warning",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "period": "daily",
    "spent": 8.5,
    "limit": 10.0,
    "percent": 85.0,
    "action": "warn",
    "project": "my-project"
  }
}
```

### ספק ירד / חזר

```json
{
  "event": "provider_down",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "provider": "anthropic-primary",
    "status": "unhealthy",
    "error": "connection timeout",
    "latency_ms": 0
  }
}
```

### מעבר לגיבוי

```json
{
  "event": "failover",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "from_provider": "anthropic-primary",
    "to_provider": "anthropic-backup",
    "reason": "rate limit exceeded",
    "session_id": "sess_abc123"
  }
}
```

### סיכום יומי

```json
{
  "event": "daily_summary",
  "timestamp": "2026-03-05T00:00:00Z",
  "data": {
    "date": "2026-03-04",
    "total_cost": 25.50,
    "total_requests": 150,
    "total_input_tokens": 125000,
    "total_output_tokens": 35000,
    "by_provider": {
      "anthropic": 18.20,
      "openai": 7.30
    }
  }
}
```

## הגדרה לפי פלטפורמה

### Slack

1. היכנסו ל-[Slack API](https://api.slack.com/apps)
2. צרו אפליקציה חדשה או בחרו קיימת
3. הפעילו "Incoming Webhooks"
4. הוסיפו את ה-webhook למרחב העבודה
5. העתיקו את כתובת ה-webhook (מתחילה ב-`https://hooks.slack.com/`)

**הגדרה:**
```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
      "events": ["budget_warning", "provider_down"]
    }
  ]
}
```

### Discord

1. פתחו את הגדרות שרת ה-Discord
2. עברו ל-Integrations ← Webhooks
3. לחצו על "New Webhook"
4. בחרו ערוץ והעתיקו את כתובת ה-webhook

**הגדרה:**
```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://discord.com/api/webhooks/123456789/XXXXXXXXXXXXXXXXXXXX",
      "events": ["budget_exceeded", "failover"]
    }
  ]
}
```

### webhook מותאם

לאינטגרציות משלכם השתמשו בתבנית ה-JSON הגנרית:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://your-server.com/webhook",
      "events": ["budget_warning", "daily_summary"],
      "headers": {
        "Authorization": "Bearer YOUR_SECRET_TOKEN",
        "X-Custom-Header": "value"
      }
    }
  ]
}
```

## הגדרה בממשק הווב

הגדרות ה-webhook נמצאות בכתובת `http://localhost:19840/settings`:

1. עברו ללשונית "Webhooks"
2. לחצו על "Add Webhook"
3. הזינו את כתובת ה-webhook
4. בחרו לאילו אירועים להירשם
5. (אופציונלי) הוסיפו כותרות מותאמות
6. לחצו על "Test" כדי לוודא את ההגדרה
7. לחצו על "Save"

## נקודות קצה ב-API

### רשימת webhooks

```bash
GET /api/v1/webhooks
```

### הוספת webhook

```bash
POST /api/v1/webhooks
Content-Type: application/json

{
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "events": ["budget_warning", "provider_down"]
}
```

### עדכון webhook

```bash
PUT /api/v1/webhooks/{id}
Content-Type: application/json

{
  "enabled": false
}
```

### מחיקת webhook

```bash
DELETE /api/v1/webhooks/{id}
```

### בדיקת webhook

```bash
POST /api/v1/webhooks/{id}/test
```

שולח הודעת בדיקה כדי לוודא את ההגדרה.

## דוגמאות הודעות

### אזהרת תקציב (Slack)

```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

### חריגה מהתקציב (Discord)

```
🚫 Budget Exceeded: monthly limit of $200.00 reached (spent: $205.50). Action: block
```

### ספק ירד (Slack)

```
🔴 Provider Down: anthropic-primary is unhealthy. Error: connection timeout
```

### ספק חזר (Discord)

```
🟢 Provider Up: anthropic-primary is healthy again (latency: 1250ms)
```

### מעבר לגיבוי (Slack)

```
🔄 Failover: Switched from anthropic-primary to anthropic-backup. Reason: rate limit exceeded
```

### סיכום יומי (Discord)

```
📊 Daily Summary (2026-03-04): 150 requests, $25.50 total cost, 125000 input / 35000 output tokens
```

## המלצות

1. **הפרידו בין webhooks** — צרו כתובות שונות לסוגי אירועים שונים
2. **בדקו לפני הפעלה** — ודאו את ההגדרה לפני השמירה
3. **אבטחו webhooks מותאמים** — השתמשו ב-HTTPS ובכותרות אימות
4. **עקבו אחרי כשלים** — אם ההתראות נפסקו, בדקו את יומני הדימון
5. **הימנעו ממידע רגיש** — אל תכניסו מפתחות API או אסימונים לכתובת ה-webhook
6. **הגדירו התראות** — הירשמו לאירועים קריטיים כמו `budget_exceeded` ו-`provider_down`

## פתרון תקלות

### ה-webhook אינו מקבל הודעות

1. ודאו שה-webhook מופעל בתצורה
2. ודאו שהכתובת נכונה (בדקו עם curl)
3. ודאו שהאירועים הוגדרו נכון
4. חפשו שגיאות webhook ביומני הדימון: `tail -f ~/.zen/zend.log`
5. בדקו את ה-webhook דרך ה-API: `POST /api/v1/webhooks/{id}/test`

### webhook של Slack נכשל

1. ודאו שהכתובת מתחילה ב-`https://hooks.slack.com/`
2. ודאו שה-webhook לא בוטל בהגדרות Slack
3. ודאו שמרחב העבודה לא חסם webhooks נכנסים
4. בדקו עם curl:
   ```bash
   curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

### webhook של Discord נכשל

1. ודאו שהכתובת מתחילה ב-`https://discord.com/api/webhooks/`
2. ודאו שה-webhook לא נמחק בהגדרות Discord
3. ודאו שלבוט יש הרשאה לפרסם בערוץ
4. בדקו עם curl:
   ```bash
   curl -X POST https://discord.com/api/webhooks/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"content":"test"}'
   ```

### webhook מותאם אינו עובד

1. ודאו שנקודת הקצה נגישה (בדקו עם curl)
2. בדקו שכותרות האימות נכונות
3. ודאו שנקודת הקצה מקבלת בקשות POST
4. ודאו שהיא מחזירה קוד סטטוס 2xx
5. בדקו את יומני נקודת הקצה לאיתור שגיאות

## שיקולי אבטחה

1. **הגנו על כתובות ה-webhook** — התייחסו אליהן כאל סודות
2. **השתמשו ב-HTTPS** — תמיד, בנקודות הקצה של ה-webhook
3. **אמתו חתימות** — מימשו בדיקת חתימה ב-webhooks מותאמים
4. **הגבילו קצב** — הפעילו הגבלת קצב בנקודות הקצה
5. **אל תרשמו מידע רגיש** — הימנעו מתיעוד המטען המלא ביומן

## הגדרות מתקדמות

### webhooks מותנים

שלחו אירועים שונים ל-webhooks שונים:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/CRITICAL/ALERTS",
      "events": ["budget_exceeded", "provider_down"]
    },
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/DAILY/REPORTS",
      "events": ["daily_summary"]
    },
    {
      "enabled": true,
      "url": "https://discord.com/api/webhooks/MONITORING",
      "events": ["failover", "provider_up"]
    }
  ]
}
```

### כותרות מותאמות לאימות

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://your-server.com/webhook",
      "events": ["budget_warning"],
      "headers": {
        "Authorization": "Bearer YOUR_SECRET_TOKEN",
        "X-API-Key": "your-api-key",
        "X-Webhook-Source": "gozen"
      }
    }
  ]
}
```
