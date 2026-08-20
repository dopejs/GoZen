---
sidebar_position: 8
title: 設定リファレンス
---

# 設定リファレンス

## ファイルの場所

| ファイル | 説明 |
|----------|------|
| `~/.zen/zen.json` | メイン設定ファイル |
| `~/.zen/zend.log` | デーモンログ |
| `~/.zen/zend.pid` | デーモン PID ファイル |
| `~/.zen/logs.db` | リクエストログデータベース（SQLite） |

## 完全な設定例

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

## フィールド一覧

| フィールド | 説明 |
|-----------|------|
| `version` | 設定ファイルのバージョン番号 |
| `default_profile` | デフォルトプロファイル名 |
| `default_client` | デフォルト CLI クライアント（claude / codex / opencode） |
| `proxy_port` | プロキシサーバーのポート（デフォルト: 19841） |
| `web_port` | Web 管理インターフェースのポート（デフォルト: 19840） |
| `providers` | プロバイダー設定の集合 |
| `profiles` | プロファイル設定の集合 |
| `project_bindings` | プロジェクトバインディング設定 |
| `sync` | 設定同期の設定（任意） |
