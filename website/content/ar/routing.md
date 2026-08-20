---
sidebar_position: 4
title: التوجيه حسب السيناريو
---

# التوجيه حسب السيناريو

توجيه الطلبات تلقائيًا إلى مزوّدين مختلفين بحسب خصائص كل طلب.

## السيناريوهات المدعومة

| السيناريو | الوصف |
|-----------|-------|
| `think` | وضع التفكير مفعّل |
| `image` | يحتوي على محتوى صوري |
| `longContext` | يتجاوز المحتوى الحد المحدد |
| `webSearch` | يستخدم أداة web_search |
| `background` | يستخدم نموذج Haiku |

## آلية الرجوع

إذا فشل جميع مزوّدي سيناريو ما، يعود التوجيه تلقائيًا إلى مزوّدي ملف التعريف الافتراضيين.

## مثال على الإعداد

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
