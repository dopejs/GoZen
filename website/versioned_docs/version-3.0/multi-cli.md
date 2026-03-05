---
sidebar_position: 6
title: Multi-CLI Support
---

# Multi-CLI Support

GoZen supports three AI coding assistant CLIs:

| CLI | Description | API Format |
|-----|-------------|------------|
| `claude` | Claude Code (default) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

## Set Default CLI

```bash
zen config default-client

# Via Web UI
zen web  # Settings page
```

## Per-Project CLI

```bash
cd ~/work/project
zen bind --cli codex  # This directory uses Codex
```

## Temporary CLI Override

```bash
zen --cli opencode  # Use OpenCode for this session
```
