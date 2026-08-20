---
sidebar_position: 6
title: Поддержка нескольких CLI
---

# Поддержка нескольких CLI

GoZen поддерживает три CLI для ИИ-ассистентов программирования:

| CLI | Описание | Формат API |
|-----|----------|------------|
| `claude` | Claude Code (по умолчанию) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

## Выбор CLI по умолчанию

```bash
zen config default-client

# Через веб-интерфейс
zen web  # страница настроек
```

## CLI для конкретного проекта

```bash
cd ~/work/project
zen bind --cli codex  # этот каталог использует Codex
```

## Временная подмена CLI

```bash
zen --cli opencode  # использовать OpenCode в этой сессии
```
