---
sidebar_position: 8
title: مرجع الإعدادات
---

# مرجع الإعدادات

## مواقع الملفات

| الملف | الوصف |
|------|-------|
| `~/.zen/zen.json` | ملف الإعدادات الرئيسي |
| `~/.zen/zend.log` | سجل الخدمة |
| `~/.zen/zend.pid` | ملف معرّف عملية الخدمة |
| `~/.zen/logs.db` | قاعدة بيانات سجل الطلبات (SQLite) |

## مثال إعداد كامل

```json
{
  "version": 7,
  "default_profile": "default",
  "default_client": "claude",
  "proxy_port": 19841,
  "web_port": 19840,
  "providers": {
    "anthropic": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000"
      }
    }
  },
  "profiles": {
    "default": {
      "providers": ["anthropic"]
    }
  },
  "project_bindings": {
    "/path/to/project": {
      "profile": "work",
      "client": "codex"
    }
  }
}
```

## مرجع الحقول

| الحقل | الوصف |
|-------|-------|
| `version` | رقم إصدار ملف الإعدادات |
| `default_profile` | اسم ملف التعريف الافتراضي |
| `default_client` | عميل سطر الأوامر الافتراضي (claude/codex/opencode) |
| `proxy_port` | منفذ خادم الوسيط (الافتراضي: 19841) |
| `web_port` | منفذ واجهة الإدارة عبر الويب (الافتراضي: 19840) |
| `providers` | مجموعة إعدادات المزوّدين |
| `profiles` | مجموعة إعدادات ملفات التعريف |
| `project_bindings` | إعدادات ارتباطات المشاريع |
| `sync` | إعدادات مزامنة الإعدادات (اختياري) |
