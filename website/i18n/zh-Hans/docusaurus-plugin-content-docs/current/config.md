---
sidebar_position: 8
title: 配置参考
---

# 配置参考

## 文件位置

| File | Description |
|------|-------------|
| `~/.zen/zen.json` | 主配置文件 |
| `~/.zen/zend.log` | 守护进程日志 |
| `~/.zen/zend.pid` | 守护进程 PID 文件 |
| `~/.zen/logs.db` | 请求日志数据库（SQLite） |

## 完整配置示例

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

## 字段说明

| Field | Description |
|-------|-------------|
| `version` | 配置文件版本号 |
| `default_profile` | 默认使用的 profile 名称 |
| `default_client` | 默认使用的 CLI 客户端（claude/codex/opencode） |
| `proxy_port` | 代理服务端口（默认：19841） |
| `web_port` | Web 管理界面端口（默认：19840） |
| `providers` | Provider 配置集合 |
| `profiles` | Profile 配置集合 |
| `project_bindings` | 项目绑定配置 |
| `sync` | 配置同步设置（可选） |
