---
sidebar_position: 2
title: Provider 管理
---

# Provider 管理

Provider 代表一个 API 端点配置，包含 Base URL、认证 Token、模型名称等信息。

## 配置示例

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

## 环境变量

每个 provider 可以为不同 CLI 配置独立的环境变量：

### Claude Code 常用环境变量

| Variable | Description |
|----------|-------------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | 最大输出 token |
| `MAX_THINKING_TOKENS` | 扩展思考预算 |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | 最大上下文窗口 |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash 默认超时 |
