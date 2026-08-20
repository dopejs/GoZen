---
sidebar_position: 3
title: פרופילים וגיבוי
---

# פרופילים וגיבוי

פרופיל הוא רשימה מסודרת של ספקים לצורך מעבר לגיבוי. כשהספק הראשון אינו זמין, המערכת עוברת אוטומטית לספק הבא.

## דוגמת הגדרה

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-main", "anthropic-backup"]
    },
    "work": {
      "providers": ["company-api"],
      "routing": {
        "think": {"providers": [{"name": "thinking-api"}]}
      }
    }
  }
}
```

## שימוש בפרופילים

```bash
# שימוש בפרופיל ברירת המחדל
zen

# שימוש בפרופיל מסוים
zen -p work

# בחירה אינטראקטיבית
zen -p
```
