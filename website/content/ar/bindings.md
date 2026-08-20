---
sidebar_position: 5
title: ربط المشاريع
---

# ربط المشاريع

اربط المجلدات بملفات تعريف و/أو واجهات سطر أوامر محددة للحصول على إعداد تلقائي على مستوى المشروع.

## الاستخدام

```bash
cd ~/work/company-project

# ربط ملف تعريف
zen bind work-profile

# ربط واجهة سطر أوامر
zen bind --cli codex

# ربط الاثنين معًا
zen bind work-profile --cli codex

# التحقق من الحالة
zen status

# فك الارتباط
zen unbind
```

## الأولوية

معاملات سطر الأوامر ← ارتباطات المشروع ← الإعدادات الافتراضية العامة
