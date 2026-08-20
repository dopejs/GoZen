---
sidebar_position: 3
title: ملفات التعريف والتحويل عند الفشل
---

# ملفات التعريف والتحويل عند الفشل

ملف التعريف قائمة مرتّبة من المزوّدين للتحويل عند الفشل. فإذا تعذّر المزوّد الأول، ينتقل النظام تلقائيًا إلى الذي يليه.

## مثال على الإعداد

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

## استخدام ملفات التعريف

```bash
# استخدام ملف التعريف الافتراضي
zen

# استخدام ملف تعريف محدد
zen -p work

# الاختيار التفاعلي
zen -p
```
