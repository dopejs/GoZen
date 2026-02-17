---
sidebar_position: 8
title: 設定參考
---

# 設定參考

## 檔案位置

| File | Description |
|------|-------------|
| `~/.zen/zen.json` | 主設定檔 |
| `~/.zen/zend.log` | 守護程序日誌 |
| `~/.zen/zend.pid` | 守護程序 PID 檔案 |
| `~/.zen/logs.db` | 請求日誌資料庫（SQLite） |

## 完整設定範例

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

## 欄位說明

| Field | Description |
|-------|-------------|
| `version` | 設定檔版本號 |
| `default_profile` | 預設使用的 profile 名稱 |
| `default_client` | 預設使用的 CLI 用戶端（claude/codex/opencode） |
| `proxy_port` | 代理伺服器連接埠（預設：19841） |
| `web_port` | Web 管理介面連接埠（預設：19840） |
| `providers` | Provider 設定集合 |
| `profiles` | Profile 設定集合 |
| `project_bindings` | 專案綁定設定 |
| `sync` | 組態同步設定（可選） |
