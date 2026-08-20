---
sidebar_position: 15
title: צינור Middleware (בטא)
---

# צינור Middleware (בטא)

:::warning יכולת בבטא
צינור ה-middleware נמצא בשלב בטא. הוא כבוי כברירת מחדל ודורש הגדרה מפורשת כדי לפעול.
:::

הרחיבו את GoZen עם middleware מתחלף: שינוי בקשות ותגובות, רישום ליומן, הגבלת קצב ועיבוד משלכם.

## יכולות

- **ארכיטקטורה מתחלפת** — הוסיפו לוגיקה משלכם בלי לשנות את הליבה
- **הרצה לפי עדיפות** — שלטו בסדר ההרצה של ה-middleware
- **נקודות אחיזה לבקשה ולתגובה** — התערבו לפני השליחה ואחרי הקבלה
- **middleware מובנים** — הזרקת הקשר, רישום ליומן, הגבלת קצב ודחיסה
- **טוען תוספים** — טענו middleware מקבצים מקומיים או מכתובות מרוחקות
- **טיפול בשגיאות** — טיפול מסודר עם התנהגות נסיגה

## ארכיטקטורה

```
Client Request
    ↓
[Middleware 1: Priority 100]
    ↓
[Middleware 2: Priority 200]
    ↓
[Middleware 3: Priority 300]
    ↓
Provider API
    ↓
[Middleware 3: Response]
    ↓
[Middleware 2: Response]
    ↓
[Middleware 1: Response]
    ↓
Client Response
```

## הגדרה

### הפעלת צינור ה-middleware

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "context-injection",
        "enabled": true,
        "priority": 100,
        "config": {}
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info"
        }
      }
    ]
  }
}
```

**אפשרויות:**

| אפשרות | תיאור |
|--------|-------------|
| `enabled` | מפעיל את צינור ה-middleware |
| `pipeline` | מערך הגדרות ה-middleware |
| `name` | מזהה ה-middleware |
| `priority` | סדר ההרצה (נמוך יותר = מוקדם יותר) |
| `config` | הגדרות ייחודיות ל-middleware |

## middleware מובנים

### 1. הזרקת הקשר

מזריק הקשר משלכם לתוך הבקשות.

```json
{
  "name": "context-injection",
  "enabled": true,
  "priority": 100,
  "config": {
    "system_prompt": "You are a helpful coding assistant.",
    "metadata": {
      "session_id": "sess_123",
      "user_id": "user_456"
    }
  }
}
```

**שימושים:**
- הוספת הנחיות מערכת
- הזרקת מטא-נתוני סשן
- הוספת הקשר משתמש

### 2. רישום בקשות

רושם ליומן את כל הבקשות והתגובות.

```json
{
  "name": "request-logger",
  "enabled": true,
  "priority": 200,
  "config": {
    "log_level": "info",
    "log_body": false,
    "log_headers": true
  }
}
```

**שימושים:**
- ניפוי שגיאות
- מסלולי ביקורת
- ניטור ביצועים

### 3. מגביל קצב

מגביל את קצב הבקשות לכל ספק או באופן גלובלי.

```json
{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60,
    "burst": 10,
    "per_provider": true
  }
}
```

**שימושים:**
- מניעת שגיאות מגבלת קצב
- שליטה בצריכת ה-API
- הגנה מפני שימוש לרעה

### 4. דחיסה (בטא)

דוחס את ההקשר כשמספר האסימונים חורג מהסף.

```json
{
  "name": "compression",
  "enabled": true,
  "priority": 400,
  "config": {
    "threshold_tokens": 50000,
    "target_tokens": 20000
  }
}
```

לפרטים ראו [דחיסת הקשר](./compression.md).

### 5. זיכרון סשן (בטא)

שומר על זיכרון השיחה בין סשנים.

```json
{
  "name": "session-memory",
  "enabled": true,
  "priority": 150,
  "config": {
    "max_memories": 100,
    "ttl_hours": 24,
    "storage": "sqlite"
  }
}
```

**שימושים:**
- לזכור העדפות משתמש
- לעקוב אחרי היסטוריית השיחה
- לשמר הקשר בין סשנים

### 6. תזמור (בטא)

מנתב בקשות לכמה ספקים ומאחד את התגובות.

```json
{
  "name": "orchestration",
  "enabled": true,
  "priority": 500,
  "config": {
    "strategy": "parallel",
    "providers": ["anthropic", "openai"],
    "consensus": "longest"
  }
}
```

**שימושים:**
- השוואת פלטים בין מודלים
- יתירות לבקשות קריטיות
- שיפור איכות דרך הסכמה בין מודלים

## middleware מותאם

### ממשק ה-middleware

```go
type Middleware interface {
    Name() string
    Priority() int
    ProcessRequest(ctx *RequestContext) error
    ProcessResponse(ctx *ResponseContext) error
}

type RequestContext struct {
    Provider  string
    Model     string
    Messages  []Message
    Metadata  map[string]interface{}
}

type ResponseContext struct {
    Provider  string
    Model     string
    Response  *APIResponse
    Latency   time.Duration
    Metadata  map[string]interface{}
}
```

### דוגמה: הזרקת כותרת מותאמת

```go
package main

import (
    "github.com/dopejs/gozen/internal/middleware"
)

type CustomHeaderMiddleware struct {
    headers map[string]string
}

func (m *CustomHeaderMiddleware) Name() string {
    return "custom-headers"
}

func (m *CustomHeaderMiddleware) Priority() int {
    return 250
}

func (m *CustomHeaderMiddleware) ProcessRequest(ctx *middleware.RequestContext) error {
    for k, v := range m.headers {
        ctx.Metadata[k] = v
    }
    return nil
}

func (m *CustomHeaderMiddleware) ProcessResponse(ctx *middleware.ResponseContext) error {
    // No response processing needed
    return nil
}

func init() {
    middleware.Register("custom-headers", func(config map[string]interface{}) middleware.Middleware {
        return &CustomHeaderMiddleware{
            headers: config["headers"].(map[string]string),
        }
    })
}
```

### טעינת middleware מותאם

#### תוסף מקומי

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "local",
        "path": "/path/to/custom-middleware.so",
        "config": {
          "headers": {
            "X-Custom-Header": "value"
          }
        }
      }
    ]
  }
}
```

#### תוסף מרוחק

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "remote",
        "url": "https://example.com/middleware/custom-headers.so",
        "checksum": "sha256:abc123...",
        "config": {}
      }
    ]
  }
}
```

## ממשק ווב

הגדרות ה-middleware נמצאות בכתובת `http://localhost:19840/settings`:

1. עברו ללשונית "Middleware" (מסומנת בתג BETA)
2. הפעילו את "Enable Middleware Pipeline"
3. הוסיפו או הסירו middleware מהצינור
4. כווננו עדיפות והגדרות
5. הפעילו או כבו כל middleware בנפרד
6. לחצו על "Save"

## נקודות קצה ב-API

### רשימת middleware

```bash
GET /api/v1/middleware
```

תגובה:
```json
{
  "enabled": true,
  "pipeline": [
    {
      "name": "context-injection",
      "enabled": true,
      "priority": 100,
      "type": "builtin"
    },
    {
      "name": "request-logger",
      "enabled": true,
      "priority": 200,
      "type": "builtin"
    }
  ]
}
```

### הוספת middleware

```bash
POST /api/v1/middleware
Content-Type: application/json

{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60
  }
}
```

### עדכון middleware

```bash
PUT /api/v1/middleware/{name}
Content-Type: application/json

{
  "enabled": false
}
```

### הסרת middleware

```bash
DELETE /api/v1/middleware/{name}
```

### טעינה מחדש של הצינור

```bash
POST /api/v1/middleware/reload
```

## תרחישי שימוש

### סביבת פיתוח

הוסיפו רישום לניפוי שגיאות ובחינת בקשות:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 100,
        "config": {
          "log_level": "debug",
          "log_body": true
        }
      }
    ]
  }
}
```

### סביבת ייצור

הוסיפו הגבלת קצב וניטור:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "rate-limiter",
        "enabled": true,
        "priority": 100,
        "config": {
          "requests_per_minute": 100,
          "burst": 20
        }
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info",
          "log_body": false
        }
      }
    ]
  }
}
```

### השוואה בין ספקים

השתמשו בתזמור כדי להשוות פלטים:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "orchestration",
        "enabled": true,
        "priority": 500,
        "config": {
          "strategy": "parallel",
          "providers": ["anthropic", "openai", "google"],
          "consensus": "longest"
        }
      }
    ]
  }
}
```

## המלצות

1. **בחרו עדיפויות מתאימות** — מספרים נמוכים רצים קודם
2. **שמרו על middleware ממוקד** — כל אחד עושה דבר אחד היטב
3. **טפלו בשגיאות בזהירות** — שגיאה לא אמורה לשבור את הצינור
4. **בדקו ביסודיות** — ודאו את ההתנהגות לפני ייצור
5. **עקבו אחרי ביצועים** — מדדו את התקורה
6. **תעדו את ההגדרות** — תארו בבירור את האפשרויות

## מגבלות

1. **תקורת ביצועים** — כל middleware מוסיף השהיה
2. **מורכבות** — יותר מדי middleware מקשים על ניפוי שגיאות
3. **אבטחת תוספים** — תוספים מרוחקים דורשים אמון ואימות
4. **התפשטות שגיאות** — שגיאה ב-middleware עלולה להשפיע על כל הבקשות
5. **מורכבות הגדרה** — צינורות מורכבים קשים יותר לתחזוקה

## פתרון תקלות

### ה-middleware אינו רץ

1. ודאו ש-`middleware.enabled` מוגדר `true`
2. ודאו שה-middleware מופעל בצינור
3. ודאו שהעדיפות נקבעה נכון
4. עברו על יומני הדימון לאיתור שגיאות middleware

### התנהגות לא צפויה

1. בדקו את סדר ההרצה (עדיפות)
2. ודאו שההגדרות נכונות
3. בדקו את ה-middleware בבידוד
4. עברו על יומני ה-middleware

### בעיות ביצועים

1. אתרו את ה-middleware האיטי (לפי היומנים)
2. הפחיתו את מספר ה-middleware
3. שפרו את המימוש
4. שקלו לכבות middleware שאינו חיוני

### כשל בטעינת תוסף

1. ודאו שנתיב התוסף נכון
2. ודאו שהתוסף הודר לארכיטקטורה הנכונה
3. ודאו שסכום הביקורת תואם (בתוספים מרוחקים)
4. עברו על יומני התוסף לאיתור שגיאות

## שיקולי אבטחה

1. **אמתו תוספים** — טענו רק תוספים מהימנים
2. **בדקו סכומי ביקורת** — תמיד בתוספים מרוחקים
3. **בודדו תוספים** — שקלו להריץ אותם בסביבה מבודדת
4. **בצעו ביקורת קוד** — קראו את קוד ה-middleware לפני פריסה
5. **עקבו אחרי ההתנהגות** — שימו לב להתנהגות חריגה

## שיפורים עתידיים

- תמיכה בתוספי WebAssembly לתאימות חוצת פלטפורמות
- זירה לשיתוף תוספים מהקהילה
- עורך צינור חזותי בממשק הווב
- פרופיילינג ביצועים ל-middleware
- טעינה חמה בעדכוני תוספים
- מסגרת בדיקות ל-middleware
