---
sidebar_position: 6
title: マルチ CLI サポート
---

# マルチ CLI サポート

GoZen は 3 種類の AI コーディング支援 CLI をサポートします:

| CLI | 説明 | API フォーマット |
|-----|------|------------------|
| `claude` | Claude Code（デフォルト） | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

## デフォルト CLI の設定

```bash
zen config default-client

# Web UI から設定
zen web  # Settings ページ
```

## プロジェクトごとの CLI

```bash
cd ~/work/project
zen bind --cli codex  # このディレクトリでは Codex を使う
```

## 一時的な CLI 上書き

```bash
zen --cli opencode  # このセッションのみ OpenCode を使う
```
