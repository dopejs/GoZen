---
sidebar_position: 2
title: المزوّدون
---

# إدارة المزوّدين

يمثّل المزوّد إعدادات نقطة نهاية لـ API: عنوان الأساس، ورمز المصادقة، واسم النموذج، وغير ذلك.

## مثال على الإعداد

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000"
      },
      "codex_env_vars": {
        "CODEX_SOME_VAR": "value"
      },
      "opencode_env_vars": {
        "OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": "64000"
      }
    }
  }
}
```

## متغيرات البيئة

يمكن لكل مزوّد أن يحدّد متغيرات بيئة خاصة بكل واجهة سطر أوامر:

### متغيرات بيئة شائعة في Claude Code

| المتغير | الوصف |
|---------|-------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | الحد الأقصى لرموز الإخراج |
| `MAX_THINKING_TOKENS` | ميزانية التفكير الموسّع |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | الحد الأقصى لنافذة السياق |
| `BASH_DEFAULT_TIMEOUT_MS` | المهلة الافتراضية لـ Bash |
