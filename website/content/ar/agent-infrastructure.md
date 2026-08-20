---
sidebar_position: 16
title: بنية الوكلاء (تجريبي)
---

# بنية الوكلاء (تجريبي)

:::warning ميزة تجريبية
بنية الوكلاء ما زالت في مرحلة تجريبية. وهي معطّلة افتراضيًا وتحتاج إلى إعداد صريح لتفعيلها.
:::

دعم مدمج لسير عمل الوكلاء المستقلين: إدارة الجلسات، وتنسيق الملفات، والمراقبة اللحظية، وضوابط الأمان.

## المزايا

- **بيئة تشغيل الوكلاء** — تنفيذ مهام مستقلة بإدارة كاملة لدورة الحياة
- **المرصد** — مراقبة لحظية لجلسات الوكلاء وأنشطتهم
- **حواجز الأمان** — ضوابط وقيود على سلوك الوكلاء
- **المنسّق** — تنسيق عبر الملفات لسير العمل متعدد الوكلاء
- **طابور المهام** — إدارة المهام بالأولوية والاعتماديات
- **إدارة الجلسات** — تتبّع جلسات الوكلاء عبر عدة مشاريع

## البنية

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

## الإعداد

### تفعيل بنية الوكلاء

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

## المكوّنات

### 1. بيئة تشغيل الوكلاء

تدير دورة حياة تنفيذ مهام الوكيل.

**المزايا:**
- جدولة المهام وتنفيذها
- إدارة المهام المتزامنة
- التعامل مع المهلات
- تنظيف تلقائي
- التعافي من الأخطاء

**الإعداد:**
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

**الـ API:**
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

### 2. المرصد

مراقبة لحظية لأنشطة الوكلاء.

**المزايا:**
- تتبّع الجلسات
- تسجيل النشاط
- مقاييس الأداء
- تحديثات الحالة
- بيانات تاريخية

**الإعداد:**
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

**المقاييس المراقَبة:**
- الجلسات النشطة
- المهام الجارية
- استهلاك الرموز
- استدعاءات الـ API
- عمليات الملفات
- نسبة الأخطاء
- متوسط الكمون

**الـ API:**
```bash
# Get all active sessions
GET /api/v1/agent/sessions

# Get session details
GET /api/v1/agent/sessions/{session_id}

# Get session metrics
GET /api/v1/agent/sessions/{session_id}/metrics
```

### 3. حواجز الأمان

ضوابط وقيود على سلوك الوكلاء.

**المزايا:**
- حدود على العمليات
- تقييد المسارات
- حجب الأوامر
- حصص للموارد
- مسارات موافقة

**الإعداد:**
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

**آلية التطبيق:**
- تحقّق قبل التنفيذ
- مراقبة لحظية
- حجب تلقائي
- طلبات موافقة
- تسجيل للتدقيق

**الـ API:**
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

### 4. المنسّق

تنسيق عبر الملفات لسير العمل متعدد الوكلاء.

**المزايا:**
- قفل الملفات
- اكتشاف التغييرات
- حل التعارضات
- مزامنة الحالة
- إشعارات الأحداث

**الإعداد:**
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

**حالات الاستخدام:**
- عدة وكلاء يحرّرون الملفات نفسها
- منع التعديلات المتزامنة
- اكتشاف تغييرات خارجية على الملفات
- تنسيق سير عمل الوكلاء

**الـ API:**
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

### 5. طابور المهام

يدير مهام الوكلاء بالأولوية والاعتماديات.

**المزايا:**
- جدولة حسب الأولوية
- اعتماديات بين المهام
- إدارة الطابور
- تتبّع الحالة
- منطق إعادة المحاولة

**الإعداد:**
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

**الـ API:**
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

## واجهة الويب

لوحة الوكلاء متاحة على `http://localhost:19840/agent`:

### تبويب الجلسات

- **الجلسات النشطة** — جلسات الوكلاء العاملة حاليًا
- **تفاصيل الجلسة** — تقدّم المهام والمقاييس والسجلات
- **التحكم بالجلسة** — إيقاف مؤقت، استئناف، إلغاء

### تبويب المهام

- **طابور المهام** — المهام المنتظرة والجارية
- **سجل المهام** — المهام المكتملة والفاشلة
- **تفاصيل المهمة** — الإعدادات والسجلات والنتائج

### تبويب حواجز الأمان

- **حدود العمليات** — الاستخدام الحالي مقابل الحدود
- **العمليات المحجوبة** — المحاولات المحجوبة مؤخرًا
- **طابور الموافقات** — العمليات بانتظار الموافقة

### تبويب المقاييس

- **استهلاك الرموز** — لكل جلسة وبالإجمالي
- **استدعاءات الـ API** — عدد الطلبات ومعدّلها
- **عمليات الملفات** — القراءة والكتابة والحذف
- **الأداء** — الكمون والإنتاجية

## التكامل مع Claude Code

يكتشف GoZen جلسات Claude Code تلقائيًا ويوفّر لها بنية الوكلاء:

```bash
# Start Claude Code with agent support
zen --agent

# Agent features are automatically enabled:
# - Session tracking
# - File coordination
# - Guardrails enforcement
# - Real-time monitoring
```

**الفوائد:**
- منع تعديل الملفات في وقت واحد
- تتبّع استهلاك الرموز والتكاليف
- فرض قيود الأمان
- مراقبة أنشطة الوكلاء
- تنسيق سير العمل متعدد الوكلاء

## حالات الاستخدام

### تطوير بعدة وكلاء

عدة وكلاء يعملون على القاعدة البرمجية نفسها:

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

### مهام طويلة الأمد

مراقبة المهام الطويلة والتحكم بها:

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

### عمليات حرجة أمنيًا

فرض ضوابط أمان صارمة:

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

## أفضل الممارسات

1. **فعّل حواجز الأمان** — دائمًا في بيئة الإنتاج
2. **اضبط حدودًا مناسبة** — وفق حالة استخدامك
3. **راقب باستمرار** — تفقّد لوحة المرصد بانتظام
4. **استخدم قفل الملفات** — فعّل المنسّق في سير العمل متعدد الوكلاء
5. **اضبط الموافقات** — اطلب موافقة على العمليات المدمِّرة
6. **راجع السجلات** — دقّق أنشطة الوكلاء بانتظام

## القيود

1. **عبء على الأداء** — المراقبة والتنسيق يضيفان كمونًا
2. **قفل الملفات** — قد يسبّب تأخيرًا في سيناريوهات متعددة الوكلاء
3. **استهلاك الذاكرة** — سجل الجلسات يستهلك ذاكرة
4. **التعقيد** — يتطلب فهمًا لسير عمل الوكلاء
5. **حالة تجريبية** — قد تتغيّر المزايا في الإصدارات القادمة

## معالجة المشكلات

### جلسة الوكيل لا تُتتبَّع

1. تأكد أن `agent.enabled` مضبوط على `true`
2. تأكد أن المرصد مفعّل
3. تأكد أن عميل الوكيل مدعوم (Claude Code، ‏Codex)
4. راجع سجلات الخدمة بحثًا عن الأخطاء

### مشكلات في قفل الملفات

1. تأكد أن المنسّق مفعّل
2. تأكد أن مهلة القفل مناسبة
3. راجع الأقفال النشطة: `GET /api/v1/agent/locks`
4. حرّر الأقفال العالقة يدويًا عند الحاجة

### حواجز الأمان لا تُطبَّق

1. تأكد أنها مفعّلة
2. تأكد من صحة إعداد القواعد
3. راجع سجل العمليات المحجوبة
4. تأكد أن عميل الوكيل يحترم حواجز الأمان

### استهلاك ذاكرة مرتفع

1. قلّل مدة الاحتفاظ بالسجل
2. زد الفاصل الزمني للتحديث
3. حدّ من عدد المهام المتزامنة
4. فعّل التنظيف التلقائي

## اعتبارات أمنية

1. **تقييد المسارات** — اضبط دائمًا المسارات المسموحة والمحجوبة
2. **حجب الأوامر** — احجب الأوامر الخطرة
3. **مسارات الموافقة** — اطلب موافقة على العمليات الحساسة
4. **تسجيل التدقيق** — فعّل تسجيلًا شاملًا
5. **حدود الموارد** — اضبط حدودًا مناسبة للعمليات

## تحسينات مستقبلية

- بروتوكولات للتعاون بين عدة وكلاء
- استراتيجيات متقدمة لحل التعارضات
- تعلّم آلي لاكتشاف الشذوذ
- تكامل مع أدوات مراقبة خارجية
- تحليلات لسلوك الوكلاء
- توليد آلي لسياسات الأمان
