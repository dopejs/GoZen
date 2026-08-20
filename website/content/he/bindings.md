---
sidebar_position: 5
title: קישור פרויקטים
---

# קישור פרויקטים

קשרו תיקיות לפרופילים ו/או לכלי CLI מסוימים כדי לקבל הגדרה אוטומטית ברמת הפרויקט.

## שימוש

```bash
cd ~/work/company-project

# קישור פרופיל
zen bind work-profile

# קישור כלי CLI
zen bind --cli codex

# קישור שניהם
zen bind work-profile --cli codex

# בדיקת מצב
zen status

# ביטול הקישור
zen unbind
```

## סדר עדיפויות

ארגומנטים של ה-CLI ← קישורי פרויקט ← ברירות מחדל גלובליות
