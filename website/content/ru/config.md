---
sidebar_position: 8
title: Справочник конфигурации
---

# Справочник конфигурации

## Расположение файлов

| Файл | Описание |
|------|-------|
| `~/.zen/zen.json` | Основной файл конфигурации |
| `~/.zen/zend.log` | Журнал демона |
| `~/.zen/zend.pid` | PID-файл демона |
| `~/.zen/logs.db` | База журналов запросов (SQLite) |

## Полный пример конфигурации

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

## Справочник полей

| Поле | Описание |
|-------|-------|
| `version` | Номер версии файла конфигурации |
| `default_profile` | Имя профиля по умолчанию |
| `default_client` | CLI-клиент по умолчанию (claude/codex/opencode) |
| `proxy_port` | Порт прокси-сервера (по умолчанию: 19841) |
| `web_port` | Порт веб-интерфейса управления (по умолчанию: 19840) |
| `providers` | Набор конфигураций провайдеров |
| `profiles` | Набор конфигураций профилей |
| `project_bindings` | Конфигурация привязок проектов |
| `sync` | Настройки синхронизации конфигурации (необязательно) |
