---
sidebar_position: 6
title: Prise en charge multi-CLI
---

# Prise en charge multi-CLI

GoZen prend en charge trois CLI d’assistants de programmation :

| CLI | Description | Format d’API |
|-----|-------------|--------------|
| `claude` | Claude Code (par défaut) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

## Définir la CLI par défaut

```bash
zen config default-client

# Via l’interface web
zen web  # page Réglages
```

## CLI par projet

```bash
cd ~/work/project
zen bind --cli codex  # ce répertoire utilise Codex
```

## Remplacement temporaire de la CLI

```bash
zen --cli opencode  # utiliser OpenCode pour cette session
```
