---
sidebar_position: 13
title: Webhooks
---

# Webhooks

استقبل إشعارات فورية — تنبيهات الميزانية، وتغيّر حالة المزوّدين، والملخصات اليومية — عبر Slack أو Discord أو webhook خاص بك.

## المزايا

- **صيغ متعددة** — Slack أو Discord أو JSON عام
- **تصفية الأحداث** — اشترك في أنواع أحداث بعينها
- **ترويسات مخصّصة** — أضف مصادقة أو ترويسات خاصة بك
- **إرسال غير متزامن** — التسليم لا يعطّل الطلبات
- **تنسيق تلقائي** — رسائل غنية بالرموز التعبيرية والألوان
- **إمكانية الاختبار** — تحقّق من الإعداد قبل التفعيل

## الإعداد

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

## أنواع الأحداث

| الحدث | الوصف | متى يُطلَق |
|-------|-------------|----------------|
| `budget_warning` | بلوغ حد الميزانية | حين يصل الإنفاق إلى 80% من الحد |
| `budget_exceeded` | تجاوز حد الميزانية | حين يتجاوز الإنفاق الحد المضبوط |
| `provider_down` | صار المزوّد غير سليم | حين تهبط نسبة النجاح دون 70% |
| `provider_up` | تعافى المزوّد | حين يعود مزوّد غير سليم إلى حالته السليمة |
| `failover` | تحويل الطلب | حين ينتقل الطلب إلى مزوّد احتياطي |
| `daily_summary` | ملخص الاستخدام اليومي | مرة يوميًا عند منتصف الليل بتوقيت UTC |

## صيغ الـ Webhook

### Slack

يُكتشف تلقائيًا حين يحتوي العنوان على `slack.com`.

**مثال رسالة:**
```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

**الصيغة:**
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

يُكتشف تلقائيًا حين يحتوي العنوان على `discord.com`.

**مثال embed:**
- **العنوان:** budget_warning
- **الوصف:** ⚠️ تنبيه ميزانية: الميزانية اليومية عند 85.0% ‏($8.50 / $10.00)
- **اللون:** كهرماني (#FBBF24)
- **الطابع الزمني:** 2026-03-05T10:30:00Z

**الصيغة:**
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

### JSON عام

يُستخدم مع بقية العناوين.

**الصيغة:**
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

## بِنى بيانات الأحداث

### تنبيه الميزانية / تجاوزها

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

### سقوط المزوّد / تعافيه

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

### التحويل عند الفشل

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

### الملخص اليومي

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

## الإعداد حسب المنصة

### Slack

1. انتقل إلى [Slack API](https://api.slack.com/apps)
2. أنشئ تطبيقًا جديدًا أو اختر تطبيقًا قائمًا
3. فعّل "Incoming Webhooks"
4. أضف الـ webhook إلى مساحة العمل
5. انسخ عنوان الـ webhook (يبدأ بـ `https://hooks.slack.com/`)

**الإعداد:**
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

1. افتح إعدادات خادم Discord
2. انتقل إلى Integrations ← Webhooks
3. اضغط "New Webhook"
4. اختر القناة وانسخ عنوان الـ webhook

**الإعداد:**
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

### webhook مخصّص

للتكاملات الخاصة، استخدم صيغة JSON العامة:

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

## الإعداد من واجهة الويب

إعدادات الـ webhook متاحة على `http://localhost:19840/settings`:

1. انتقل إلى تبويب "Webhooks"
2. اضغط "Add Webhook"
3. أدخل عنوان الـ webhook
4. اختر الأحداث التي تريد الاشتراك بها
5. (اختياري) أضف ترويسات مخصّصة
6. اضغط "Test" للتحقق من الإعداد
7. اضغط "Save"

## نقاط الـ API

### سرد الـ webhooks

```bash
GET /api/v1/webhooks
```

### إضافة webhook

```bash
POST /api/v1/webhooks
Content-Type: application/json

{
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "events": ["budget_warning", "provider_down"]
}
```

### تحديث webhook

```bash
PUT /api/v1/webhooks/{id}
Content-Type: application/json

{
  "enabled": false
}
```

### حذف webhook

```bash
DELETE /api/v1/webhooks/{id}
```

### اختبار webhook

```bash
POST /api/v1/webhooks/{id}/test
```

يرسل رسالة اختبار للتحقق من الإعداد.

## أمثلة على الرسائل

### تنبيه ميزانية (Slack)

```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

### تجاوز الميزانية (Discord)

```
🚫 Budget Exceeded: monthly limit of $200.00 reached (spent: $205.50). Action: block
```

### سقوط مزوّد (Slack)

```
🔴 Provider Down: anthropic-primary is unhealthy. Error: connection timeout
```

### تعافي مزوّد (Discord)

```
🟢 Provider Up: anthropic-primary is healthy again (latency: 1250ms)
```

### تحويل عند الفشل (Slack)

```
🔄 Failover: Switched from anthropic-primary to anthropic-backup. Reason: rate limit exceeded
```

### الملخص اليومي (Discord)

```
📊 Daily Summary (2026-03-04): 150 requests, $25.50 total cost, 125000 input / 35000 output tokens
```

## أفضل الممارسات

1. **افصل بين الـ webhooks** — أنشئ عناوين مختلفة لأنواع الأحداث المختلفة
2. **اختبر قبل التفعيل** — تحقّق من الإعداد دائمًا قبل الحفظ
3. **أمّن الـ webhooks المخصّصة** — استخدم HTTPS وترويسات مصادقة
4. **راقب حالات الفشل** — إن توقفت الإشعارات فراجع سجلات الخدمة
5. **تجنّب البيانات الحساسة** — لا تضع مفاتيح API أو رموزًا في عنوان الـ webhook
6. **اضبط التنبيهات** — اشترك في الأحداث الحرجة مثل `budget_exceeded` و `provider_down`

## معالجة المشكلات

### الـ webhook لا يستقبل رسائل

1. تأكد أن الـ webhook مفعّل في الإعدادات
2. تأكد من صحة العنوان (اختبره بـ curl)
3. تأكد من ضبط الأحداث بشكل صحيح
4. راجع سجلات الخدمة بحثًا عن أخطاء الـ webhook: `tail -f ~/.zen/zend.log`
5. اختبر الـ webhook عبر الـ API: `POST /api/v1/webhooks/{id}/test`

### فشل webhook الخاص بـ Slack

1. تأكد أن العنوان يبدأ بـ `https://hooks.slack.com/`
2. تأكد أن الـ webhook لم يُلغَ من إعدادات Slack
3. تأكد أن مساحة العمل لم تعطّل الـ webhooks الواردة
4. اختبر بـ curl:
   ```bash
   curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

### فشل webhook الخاص بـ Discord

1. تأكد أن العنوان يبدأ بـ `https://discord.com/api/webhooks/`
2. تأكد أن الـ webhook لم يُحذف من إعدادات Discord
3. تأكد أن للروبوت صلاحية النشر في القناة
4. اختبر بـ curl:
   ```bash
   curl -X POST https://discord.com/api/webhooks/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"content":"test"}'
   ```

### الـ webhook المخصّص لا يعمل

1. تأكد أن نقطة النهاية متاحة (اختبرها بـ curl)
2. تأكد من صحة ترويسات المصادقة
3. تأكد أن نقطة النهاية تقبل طلبات POST
4. تأكد أنها تُعيد رمز حالة من الفئة 2xx
5. راجع سجلات نقطة النهاية بحثًا عن الأخطاء

## اعتبارات أمنية

1. **احمِ عناوين الـ webhook** — تعامل معها كأسرار
2. **استخدم HTTPS** — دائمًا في نقاط نهاية الـ webhook
3. **تحقّق من التواقيع** — طبّق التحقق من التوقيع في الـ webhooks المخصّصة
4. **حدّ من المعدّل** — طبّق تحديد معدّل على نقاط النهاية
5. **لا تسجّل بيانات حساسة** — تجنّب تدوين الحمولة كاملة في السجل

## إعدادات متقدمة

### webhooks مشروطة

أرسل أحداثًا مختلفة إلى عناوين مختلفة:

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

### ترويسات مخصّصة للمصادقة

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
