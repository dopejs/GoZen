---
sidebar_position: 6
title: دعم عدة واجهات سطر أوامر
---

# دعم عدة واجهات سطر أوامر

يدعم GoZen ثلاث واجهات سطر أوامر لمساعدي البرمجة بالذكاء الاصطناعي:

| الواجهة | الوصف | صيغة الـ API |
|---------|-------|--------------|
| `claude` | Claude Code (الافتراضي) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

## تعيين الواجهة الافتراضية

```bash
zen config default-client

# عبر واجهة الويب
zen web  # صفحة الإعدادات
```

## واجهة لكل مشروع

```bash
cd ~/work/project
zen bind --cli codex  # هذا المجلد يستخدم Codex
```

## تجاوز مؤقت للواجهة

```bash
zen --cli opencode  # استخدام OpenCode في هذه الجلسة
```
