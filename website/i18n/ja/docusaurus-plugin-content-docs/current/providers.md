---
sidebar_position: 2
title: プロバイダー管理
---

# プロバイダー管理

プロバイダーは、base URL、認証トークン、モデル名などを含む API エンドポイント設定です。

## 設定例

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

## 環境変数

各プロバイダーには CLI ごとの環境変数を設定できます。

### Claude Code でよく使う環境変数

| 変数 | 説明 |
|------|------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | 最大出力トークン数 |
| `MAX_THINKING_TOKENS` | 拡張思考用の予算 |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | 最大コンテキストウィンドウ |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash のデフォルトタイムアウト |
