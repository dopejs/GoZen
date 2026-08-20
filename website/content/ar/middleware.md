---
sidebar_position: 15
title: خط المعالجة الوسيطة (تجريبي)
---

# خط المعالجة الوسيطة (تجريبي)

:::warning ميزة تجريبية
خط المعالجة الوسيطة ما زال في مرحلة تجريبية. وهو معطّل افتراضيًا ويحتاج إلى إعداد صريح لتفعيله.
:::

وسّع GoZen بوحدات وسيطة قابلة للتركيب: تحويل الطلبات والاستجابات، والتسجيل، وتحديد المعدّل، والمعالجة المخصّصة.

## المزايا

- **بنية قابلة للتركيب** — أضف منطقك الخاص دون تعديل النواة
- **تنفيذ حسب الأولوية** — تحكّم في ترتيب تنفيذ الوحدات
- **خطّافات للطلب والاستجابة** — تدخّل قبل الإرسال وبعد الاستلام
- **وحدات مدمجة** — حقن السياق، والتسجيل، وتحديد المعدّل، والضغط
- **محمّل إضافات** — حمّل الوحدات من ملفات محلية أو عناوين بعيدة
- **معالجة الأخطاء** — معالجة سلسة مع سلوك احتياطي

## البنية

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

## الإعداد

### تفعيل خط المعالجة

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

**الخيارات:**

| الخيار | الوصف |
|--------|-------------|
| `enabled` | يفعّل خط المعالجة الوسيطة |
| `pipeline` | مصفوفة إعدادات الوحدات الوسيطة |
| `name` | معرّف الوحدة |
| `priority` | ترتيب التنفيذ (الأقل = أبكر) |
| `config` | إعدادات خاصة بالوحدة |

## الوحدات المدمجة

### 1. حقن السياق

يحقن سياقًا مخصّصًا داخل الطلبات.

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

**حالات الاستخدام:**
- إضافة موجّهات نظام
- حقن بيانات وصفية للجلسة
- إضافة سياق المستخدم

### 2. مسجّل الطلبات

يسجّل كل الطلبات والاستجابات.

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

**حالات الاستخدام:**
- تتبّع الأخطاء
- مسارات التدقيق
- مراقبة الأداء

### 3. محدّد المعدّل

يحدّ من معدّل الطلبات لكل مزوّد أو على مستوى النظام.

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

**حالات الاستخدام:**
- تفادي أخطاء حدود المعدّل
- ضبط استهلاك الـ API
- الحماية من الإساءة

### 4. الضغط (تجريبي)

يضغط السياق حين يتجاوز عدد الرموز الحد المحدد.

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

راجع [ضغط السياق](./compression.md) للتفاصيل.

### 5. ذاكرة الجلسة (تجريبي)

يحافظ على ذاكرة المحادثة عبر الجلسات.

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

**حالات الاستخدام:**
- تذكّر تفضيلات المستخدم
- تتبّع سجل المحادثة
- الحفاظ على السياق بين الجلسات

### 6. التنسيق (تجريبي)

يوجّه الطلبات إلى عدة مزوّدين ويجمع الاستجابات.

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

**حالات الاستخدام:**
- مقارنة مخرجات النماذج
- تكرار احتياطي للطلبات الحرجة
- تحسين الجودة عبر التوافق

## وحدات وسيطة مخصّصة

### واجهة الوحدة الوسيطة

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

### مثال: حقن ترويسة مخصّصة

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

### تحميل وحدة مخصّصة

#### إضافة محلية

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

#### إضافة بعيدة

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

## واجهة الويب

إعدادات الوحدات الوسيطة متاحة على `http://localhost:19840/settings`:

1. انتقل إلى تبويب "Middleware" (المؤشَّر بشارة BETA)
2. فعّل "Enable Middleware Pipeline"
3. أضف وحدات إلى الخط أو أزلها منه
4. اضبط الأولوية والإعدادات
5. فعّل أو عطّل كل وحدة على حدة
6. اضغط "Save"

## نقاط الـ API

### سرد الوحدات

```bash
GET /api/v1/middleware
```

الاستجابة:
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

### إضافة وحدة

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

### تحديث وحدة

```bash
PUT /api/v1/middleware/{name}
Content-Type: application/json

{
  "enabled": false
}
```

### إزالة وحدة

```bash
DELETE /api/v1/middleware/{name}
```

### إعادة تحميل الخط

```bash
POST /api/v1/middleware/reload
```

## حالات الاستخدام

### بيئة التطوير

أضف تسجيلًا للتنقيح وفحصًا للطلبات:

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

### بيئة الإنتاج

أضف تحديدًا للمعدّل ومراقبة:

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

### مقارنة عدة مزوّدين

استخدم التنسيق لمقارنة المخرجات:

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

## أفضل الممارسات

1. **اختر أولويات مناسبة** — الأرقام الأقل تُنفَّذ أولًا
2. **أبقِ كل وحدة مركّزة** — تؤدي مهمة واحدة وتؤديها جيدًا
3. **عالج الأخطاء بلطف** — لا تدع خطأً يكسر الخط بأكمله
4. **اختبر جيدًا** — تحقّق من السلوك قبل الإنتاج
5. **راقب الأداء** — تتبّع العبء الذي تضيفه الوحدات
6. **وثّق الإعدادات** — اشرح الخيارات بوضوح

## القيود

1. **عبء على الأداء** — كل وحدة تضيف كمونًا
2. **التعقيد** — كثرة الوحدات تُصعّب التنقيح
3. **أمان الإضافات** — الإضافات البعيدة تتطلب ثقة وتحققًا
4. **انتشار الأخطاء** — خطأ في وحدة قد يؤثر على كل الطلبات
5. **تعقيد الإعداد** — الخطوط المعقّدة أصعب في الصيانة

## معالجة المشكلات

### الوحدة لا تُنفَّذ

1. تأكد أن `middleware.enabled` مضبوط على `true`
2. تأكد أن الوحدة مفعّلة داخل الخط
3. تأكد من ضبط الأولوية بشكل صحيح
4. راجع سجلات الخدمة بحثًا عن أخطاء الوحدات

### سلوك غير متوقع

1. تحقّق من ترتيب التنفيذ (الأولوية)
2. تأكد من صحة الإعدادات
3. اختبر الوحدة بمعزل عن غيرها
4. راجع سجلات الوحدة

### مشكلات في الأداء

1. حدّد الوحدة البطيئة (من السجلات)
2. قلّل عدد الوحدات
3. حسّن تنفيذ الوحدة
4. فكّر في تعطيل الوحدات غير الضرورية

### فشل تحميل الإضافة

1. تأكد من صحة مسار الإضافة
2. تأكد أنها مُصرَّفة للمعمارية الصحيحة
3. تأكد من تطابق مجموع التحقق (للإضافات البعيدة)
4. راجع سجلات الإضافة بحثًا عن الأخطاء

## اعتبارات أمنية

1. **تحقّق من الإضافات** — لا تحمّل إلا الموثوقة منها
2. **تأكد من مجاميع التحقق** — دائمًا مع الإضافات البعيدة
3. **اعزل الإضافات** — فكّر في تشغيلها في بيئة معزولة
4. **دقّق الوحدات** — راجع الشيفرة قبل النشر
5. **راقب السلوك** — انتبه لأي سلوك غير متوقع

## تحسينات مستقبلية

- دعم إضافات WebAssembly للتوافق عبر المنصات
- سوق لمشاركة إضافات المجتمع
- محرّر مرئي للخط داخل واجهة الويب
- قياس أداء الوحدات الوسيطة
- إعادة تحميل فورية عند تحديث الإضافات
- إطار لاختبار الوحدات الوسيطة
