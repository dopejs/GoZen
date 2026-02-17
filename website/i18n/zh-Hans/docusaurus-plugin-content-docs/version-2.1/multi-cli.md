---
sidebar_position: 6
title: 多 CLI 支持
---

# 多 CLI 支持

GoZen 支持三种 AI 编程助手 CLI：

| CLI | Description | API Format |
|-----|-------------|------------|
| `claude` | Claude Code（默认） | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

## 设置默认 CLI

```bash
zen config default-client

# Via Web UI
zen web  # Settings page
```

## 按项目配置 CLI

```bash
cd ~/work/project
zen bind --cli codex  # This directory uses Codex
```

## 临时使用其他 CLI

```bash
zen --cli opencode  # Use OpenCode for this session
```
