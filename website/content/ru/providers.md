---
sidebar_position: 2
title: Провайдеры
---

# Управление провайдерами

Провайдер описывает конфигурацию API-эндпоинта: базовый URL, токен авторизации, имя модели и прочее.

## Пример конфигурации

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

## Переменные окружения

У каждого провайдера могут быть свои переменные окружения для каждой CLI:

### Частые переменные окружения Claude Code

| Переменная | Описание |
|------------|----------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | Максимум токенов в ответе |
| `MAX_THINKING_TOKENS` | Бюджет расширенного размышления |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | Максимальное окно контекста |
| `BASH_DEFAULT_TIMEOUT_MS` | Тайм-аут Bash по умолчанию |
