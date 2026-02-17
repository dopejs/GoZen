---
sidebar_position: 6
title: 多 CLI 支援
---

# 多 CLI 支援

GoZen 支援三種 AI 程式設計助手 CLI：

| CLI | Description | API Format |
|-----|-------------|------------|
| `claude` | Claude Code（預設） | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

## 設定預設 CLI

```bash
zen config default-client

# Via Web UI
zen web  # Settings page
```

## 按專案設定 CLI

```bash
cd ~/work/project
zen bind --cli codex  # This directory uses Codex
```

## 臨時使用其他 CLI

```bash
zen --cli opencode  # Use OpenCode for this session
```
