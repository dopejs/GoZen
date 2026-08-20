---
title: איזון עומסים
---

# איזון עומסים

מעבר למעבר פשוט לגיבוי, GoZen תומך בכמה אסטרטגיות לבחירת ספק. אפשר לבחור אסטרטגיה לכל פרופיל, לשלב אותה עם בדיקות תקינות ולנתב תעבורה לפי זמינות, השהיה או עלות.

## אסטרטגיות זמינות

### מעבר לגיבוי

מנסה את הספקים לפי הסדר עד שאחד מצליח. זו ברירת המחדל, ומתאימה למבנה של ספק ראשי וגיבוי.

```json
{
  "profiles": {
    "default": {
      "providers": ["primary", "backup"],
      "strategy": "failover"
    }
  }
}
```

### סבב מחזורי

מחלק את הבקשות באופן שווה בין כמה ספקים שקולים.

```json
{
  "profiles": {
    "balanced": {
      "providers": ["provider-a", "provider-b", "provider-c"],
      "strategy": "round-robin"
    }
  }
}
```

### השהיה מזערית

מעדיף את הספק עם זמן התגובה הנמוך ביותר לאחרונה.

```json
{
  "profiles": {
    "fast": {
      "providers": ["us-east", "us-west", "eu"],
      "strategy": "least-latency"
    }
  }
}
```

### עלות מזערית

מעדיף את הספק הזול ביותר עבור המודל המבוקש.

```json
{
  "profiles": {
    "budget": {
      "providers": ["cheap-provider", "premium-provider"],
      "strategy": "least-cost"
    }
  }
}
```

## ניתוב מודע לתקינות

כל האסטרטגיות עובדות יחד עם ניטור התקינות. כאשר `health_aware` מופעל, ספקים לא תקינים מדולגים אוטומטית עד שהם מתאוששים.

```json
{
  "profiles": {
    "production": {
      "providers": ["primary", "secondary", "tertiary"],
      "strategy": "least-latency",
      "health_aware": true
    }
  }
}
```

## איך בוחרים אסטרטגיה

- השתמשו ב-`failover` כשאמינות היא השיקול הראשון.
- השתמשו ב-`round-robin` כשהספקים שווי ערך.
- השתמשו ב-`least-latency` לעומסים אינטראקטיביים או רגישים לזמן.
- השתמשו ב-`least-cost` כשהתקציב חשוב יותר ממהירות גולמית.

## מסמכים קשורים

- [פרופילים](/docs/profiles) מסביר כיצד מגדירים קבוצות ספקים.
- [ניתוב](/docs/routing) עוסק בבחירת ספק לפי תרחיש.
- [ניטור תקינות](/docs/health-monitoring) מסביר כיצד בדיקות התקינות משפיעות על הניתוב.
