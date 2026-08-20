---
sidebar_position: 16
title: תשתית סוכנים (בטא)
---

# תשתית סוכנים (בטא)

:::warning יכולת בבטא
תשתית הסוכנים נמצאת בשלב בטא. היא כבויה כברירת מחדל ודורשת הגדרה מפורשת כדי לפעול.
:::

תמיכה מובנית בתהליכי סוכנים אוטונומיים: ניהול סשנים, תיאום קבצים, ניטור בזמן אמת ובקרות בטיחות.

## יכולות

- **סביבת ריצה לסוכנים** — הרצת משימות אוטונומיות עם ניהול מלא של מחזור החיים
- **מצפה** — ניטור סשנים ופעילות של סוכנים בזמן אמת
- **מעקות בטיחות** — בקרות ומגבלות על התנהגות הסוכנים
- **מתאם** — תיאום מבוסס קבצים לתהליכים עם כמה סוכנים
- **תור משימות** — ניהול משימות עם עדיפויות ותלויות
- **ניהול סשנים** — מעקב אחרי סשני סוכנים בכמה פרויקטים

## ארכיטקטורה

```
Agent Client (Claude Code, Codex, etc.)
    ↓
Agent Runtime
    ↓
┌─────────────┬──────────────┬─────────────┐
│ Observatory │ Guardrails   │ Coordinator │
│ (Monitor)   │ (Safety)     │ (Sync)      │
└─────────────┴──────────────┴─────────────┘
    ↓
Task Queue → Provider API
```

## הגדרה

### הפעלת תשתית הסוכנים

```json
{
  "agent": {
    "enabled": true,
    "runtime": {
      "max_concurrent_tasks": 5,
      "task_timeout": "30m",
      "auto_cleanup": true
    },
    "observatory": {
      "enabled": true,
      "update_interval": "5s",
      "history_retention": "7d"
    },
    "guardrails": {
      "enabled": true,
      "max_file_operations": 100,
      "max_api_calls": 1000,
      "allowed_paths": ["/Users/john/projects"],
      "blocked_commands": ["rm -rf", "sudo"]
    },
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    }
  }
}
```

## רכיבים

### 1. סביבת ריצה לסוכנים

מנהלת את מחזור החיים של הרצת משימות הסוכן.

**יכולות:**
- תזמון והרצה של משימות
- ניהול משימות מקבילות
- טיפול בפסקי זמן
- ניקוי אוטומטי
- התאוששות משגיאות

**הגדרה:**
```json
{
  "runtime": {
    "max_concurrent_tasks": 5,
    "task_timeout": "30m",
    "auto_cleanup": true,
    "retry_failed_tasks": true,
    "max_retries": 3
  }
}
```

**API:**
```bash
# Start agent task
POST /api/v1/agent/tasks
Content-Type: application/json

{
  "name": "code-review",
  "description": "Review pull request #123",
  "priority": 1,
  "config": {
    "model": "claude-opus-4",
    "max_tokens": 100000
  }
}

# Get task status
GET /api/v1/agent/tasks/{task_id}

# Cancel task
DELETE /api/v1/agent/tasks/{task_id}
```

### 2. מצפה

ניטור פעילות הסוכנים בזמן אמת.

**יכולות:**
- מעקב אחרי סשנים
- רישום פעילות
- מדדי ביצועים
- עדכוני מצב
- נתונים היסטוריים

**הגדרה:**
```json
{
  "observatory": {
    "enabled": true,
    "update_interval": "5s",
    "history_retention": "7d",
    "metrics": {
      "track_tokens": true,
      "track_costs": true,
      "track_latency": true
    }
  }
}
```

**מדדים מנוטרים:**
- סשנים פעילים
- משימות בביצוע
- צריכת אסימונים
- קריאות API
- פעולות על קבצים
- שיעור שגיאות
- השהיה ממוצעת

**API:**
```bash
# Get all active sessions
GET /api/v1/agent/sessions

# Get session details
GET /api/v1/agent/sessions/{session_id}

# Get session metrics
GET /api/v1/agent/sessions/{session_id}/metrics
```

### 3. מעקות בטיחות

בקרות ומגבלות על התנהגות הסוכנים.

**יכולות:**
- מגבלות על פעולות
- הגבלת נתיבים
- חסימת פקודות
- מכסות משאבים
- תהליכי אישור

**הגדרה:**
```json
{
  "guardrails": {
    "enabled": true,
    "max_file_operations": 100,
    "max_api_calls": 1000,
    "max_tokens_per_session": 1000000,
    "allowed_paths": [
      "/Users/john/projects",
      "/tmp/agent-workspace"
    ],
    "blocked_paths": [
      "/etc",
      "/System",
      "~/.ssh"
    ],
    "blocked_commands": [
      "rm -rf /",
      "sudo",
      "chmod 777"
    ],
    "require_approval": {
      "file_delete": true,
      "system_commands": true,
      "network_requests": false
    }
  }
}
```

**אכיפה:**
- אימות לפני ההרצה
- ניטור בזמן אמת
- חסימה אוטומטית
- בקשות אישור
- רישום לביקורת

**API:**
```bash
# Get guardrail status
GET /api/v1/agent/guardrails

# Update guardrail rules
PUT /api/v1/agent/guardrails
Content-Type: application/json

{
  "max_file_operations": 200,
  "blocked_commands": ["rm -rf", "sudo", "dd"]
}
```

### 4. מתאם

תיאום מבוסס קבצים לתהליכים עם כמה סוכנים.

**יכולות:**
- נעילת קבצים
- זיהוי שינויים
- יישוב התנגשויות
- סנכרון מצב
- התראות על אירועים

**הגדרה:**
```json
{
  "coordinator": {
    "enabled": true,
    "lock_timeout": "5m",
    "change_detection": true,
    "conflict_resolution": "last-write-wins",
    "notification_webhook": "https://hooks.slack.com/..."
  }
}
```

**שימושים:**
- כמה סוכנים עורכים את אותם קבצים
- מניעת שינויים במקביל
- זיהוי שינויי קבצים חיצוניים
- תיאום תהליכי סוכנים

**API:**
```bash
# Acquire file lock
POST /api/v1/agent/locks
Content-Type: application/json

{
  "path": "/path/to/file.go",
  "session_id": "sess_123",
  "timeout": "5m"
}

# Release file lock
DELETE /api/v1/agent/locks/{lock_id}

# Get file change events
GET /api/v1/agent/changes?since=2026-03-05T10:00:00Z
```

### 5. תור משימות

מנהל משימות סוכן עם עדיפויות ותלויות.

**יכולות:**
- תזמון לפי עדיפות
- תלויות בין משימות
- ניהול התור
- מעקב מצב
- לוגיקת ניסיון חוזר

**הגדרה:**
```json
{
  "task_queue": {
    "enabled": true,
    "max_queue_size": 100,
    "priority_levels": 5,
    "enable_dependencies": true,
    "retry_policy": {
      "max_retries": 3,
      "backoff": "exponential"
    }
  }
}
```

**API:**
```bash
# Add task to queue
POST /api/v1/agent/queue
Content-Type: application/json

{
  "name": "run-tests",
  "priority": 2,
  "depends_on": ["build-project"],
  "config": {}
}

# Get queue status
GET /api/v1/agent/queue

# Remove task from queue
DELETE /api/v1/agent/queue/{task_id}
```

## ממשק ווב

לוח הסוכנים נמצא בכתובת `http://localhost:19840/agent`:

### לשונית Sessions

- **סשנים פעילים** — סשני סוכן שרצים כעת
- **פרטי סשן** — התקדמות משימות, מדדים ויומנים
- **בקרת סשן** — השהיה, המשך, ביטול

### לשונית Tasks

- **תור משימות** — משימות ממתינות ובביצוע
- **היסטוריית משימות** — משימות שהושלמו ושנכשלו
- **פרטי משימה** — הגדרות, יומנים ותוצאות

### לשונית Guardrails

- **מגבלות פעולות** — השימוש הנוכחי מול המגבלות
- **פעולות חסומות** — ניסיונות שנחסמו לאחרונה
- **תור אישורים** — פעולות הממתינות לאישור

### לשונית Metrics

- **צריכת אסימונים** — לכל סשן ובסך הכול
- **קריאות API** — מספר הבקשות והקצב
- **פעולות על קבצים** — קריאה, כתיבה ומחיקה
- **ביצועים** — השהיה ותפוקה

## שילוב עם Claude Code

‏GoZen מזהה אוטומטית סשנים של Claude Code ומספק להם את תשתית הסוכנים:

```bash
# Start Claude Code with agent support
zen --agent

# Agent features are automatically enabled:
# - Session tracking
# - File coordination
# - Guardrails enforcement
# - Real-time monitoring
```

**התועלת:**
- מניעת שינויי קבצים במקביל
- מעקב אחרי צריכת אסימונים ועלויות
- אכיפת מגבלות בטיחות
- ניטור פעילות הסוכנים
- תיאום תהליכים עם כמה סוכנים

## תרחישי שימוש

### פיתוח עם כמה סוכנים

כמה סוכנים עובדים על אותו בסיס קוד:

```json
{
  "agent": {
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    },
    "guardrails": {
      "max_file_operations": 200,
      "allowed_paths": ["/Users/john/project"]
    }
  }
}
```

### משימות ארוכות

ניטור ושליטה במשימות סוכן ארוכות:

```json
{
  "agent": {
    "runtime": {
      "task_timeout": "2h",
      "auto_cleanup": false
    },
    "observatory": {
      "update_interval": "10s",
      "history_retention": "30d"
    }
  }
}
```

### פעולות קריטיות לבטיחות

אכיפת בקרות בטיחות מחמירות:

```json
{
  "agent": {
    "guardrails": {
      "enabled": true,
      "max_file_operations": 50,
      "blocked_commands": ["rm", "sudo", "chmod"],
      "require_approval": {
        "file_delete": true,
        "system_commands": true,
        "network_requests": true
      }
    }
  }
}
```

## המלצות

1. **הפעילו מעקות בטיחות** — תמיד בסביבת ייצור
2. **קבעו מגבלות מתאימות** — התאימו אותן לתרחיש שלכם
3. **נטרו באופן פעיל** — בדקו את לוח המצפה בקביעות
4. **השתמשו בנעילת קבצים** — הפעילו את המתאם בתהליכים עם כמה סוכנים
5. **הגדירו אישורים** — דרשו אישור לפעולות הרסניות
6. **עברו על היומנים** — בצעו ביקורת קבועה על פעילות הסוכנים

## מגבלות

1. **תקורת ביצועים** — ניטור ותיאום מוסיפים השהיה
2. **נעילת קבצים** — עלולה לגרום לעיכובים בתרחישים מרובי סוכנים
3. **צריכת זיכרון** — היסטוריית הסשנים תופסת זיכרון
4. **מורכבות** — דורש הבנה של תהליכי סוכנים
5. **מצב בטא** — היכולות עשויות להשתנות בגרסאות הבאות

## פתרון תקלות

### סשן הסוכן אינו נרשם

1. ודאו ש-`agent.enabled` מוגדר `true`
2. ודאו שהמצפה מופעל
3. ודאו שלקוח הסוכן נתמך (Claude Code, ‏Codex)
4. עברו על יומני הדימון לאיתור שגיאות

### בעיות בנעילת קבצים

1. ודאו שהמתאם מופעל
2. ודאו שזמן הנעילה הקצוב מתאים
3. בדקו את הנעילות הפעילות: `GET /api/v1/agent/locks`
4. שחררו ידנית נעילות תקועות במידת הצורך

### מעקות הבטיחות אינם נאכפים

1. ודאו שהם מופעלים
2. ודאו שהגדרת הכללים נכונה
3. עברו על יומן הפעולות שנחסמו
4. ודאו שלקוח הסוכן מכבד את מעקות הבטיחות

### צריכת זיכרון גבוהה

1. קצרו את תקופת שמירת ההיסטוריה
2. הגדילו את מרווח העדכון
3. הגבילו את מספר המשימות המקבילות
4. הפעילו ניקוי אוטומטי

## שיקולי אבטחה

1. **הגבלת נתיבים** — הגדירו תמיד נתיבים מותרים וחסומים
2. **חסימת פקודות** — חסמו פקודות מסוכנות
3. **תהליכי אישור** — דרשו אישור לפעולות רגישות
4. **רישום לביקורת** — הפעילו רישום מקיף
5. **מגבלות משאבים** — קבעו מגבלות פעולה מתאימות

## שיפורים עתידיים

- פרוטוקולים לשיתוף פעולה בין סוכנים
- אסטרטגיות מתקדמות ליישוב התנגשויות
- למידת מכונה לזיהוי חריגות
- שילוב עם כלי ניטור חיצוניים
- אנליטיקה של התנהגות סוכנים
- יצירה אוטומטית של מדיניות בטיחות
