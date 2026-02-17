---
sidebar_position: 2
title: Provider 管理
---

# Provider 管理

Provider 代表一個 API 端點設定，包含 Base URL、認證 Token、模型名稱等資訊。

## 設定範例

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000"
      },
      "codex_env_vars": {
        "CODEX_SOME_VAR": "value"
      },
      "opencode_env_vars": {
        "OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": "64000"
      }
    }
  }
}
```

## 環境變數

每個 provider 可以為不同 CLI 設定獨立的環境變數：

### Claude Code 常用環境變數

| Variable | Description |
|----------|-------------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | 最大輸出 token |
| `MAX_THINKING_TOKENS` | 擴展思考預算 |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | 最大上下文視窗 |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash 預設逾時 |
