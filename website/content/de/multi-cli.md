---
sidebar_position: 6
title: Mehrere CLIs
---

# Unterstützung mehrerer CLIs

GoZen unterstützt drei CLIs für KI-gestütztes Programmieren:

| CLI | Beschreibung | API-Format |
|-----|--------------|------------|
| `claude` | Claude Code (Standard) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

## Standard-CLI festlegen

```bash
zen config default-client

# Über die Weboberfläche
zen web  # Seite „Einstellungen“
```

## CLI je Projekt

```bash
cd ~/work/project
zen bind --cli codex  # dieses Verzeichnis nutzt Codex
```

## CLI vorübergehend überschreiben

```bash
zen --cli opencode  # OpenCode für diese Sitzung verwenden
```
