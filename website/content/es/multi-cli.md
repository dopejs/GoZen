---
sidebar_position: 6
title: Soporte Multi-CLI
---

# Soporte Multi-CLI

GoZen soporta tres CLIs de asistentes de programación con IA:

| CLI | Descripción | Formato API |
|-----|-------------|-------------|
| `claude` | Claude Code (predeterminado) | API de Mensajes de Anthropic |
| `codex` | OpenAI Codex CLI | API de Completaciones de Chat de OpenAI |
| `opencode` | OpenCode | Anthropic / OpenAI |

## Establecer CLI Predeterminado

```bash
zen config default-client

# Via Web UI
zen web  # Settings page
```

## CLI por Proyecto

```bash
cd ~/work/project
zen bind --cli codex  # This directory uses Codex
```

## Anulación Temporal de CLI

```bash
zen --cli opencode  # Use OpenCode for this session
```
