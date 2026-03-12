---
title: Agents
---

# Agents

GoZen 可以作为 Claude Code、Codex 以及其他 CLI 助手的编码 Agent 运营层。在不改变现有工作流的前提下，它可以帮助你协调 Agent 工作、监控会话，并施加运行时安全控制。

## GoZen 带来的能力

- **协作协调**：减少多个 Agent 同时处理同一项目时的冲突。
- **可观测性**：集中查看会话、成本、错误和活动。
- **安全护栏**：对支出、请求频率和敏感操作施加限制。
- **任务路由**：把不同类型的工作分配给不同 provider 或 profile。

## 配置示例

```json
{
  "agent": {
    "enabled": true,
    "coordinator": {
      "enabled": true,
      "lock_timeout_sec": 300,
      "inject_warnings": true
    },
    "observatory": {
      "enabled": true,
      "stuck_threshold": 5,
      "idle_timeout_min": 30
    },
    "guardrails": {
      "enabled": true,
      "session_spending_cap": 5.0,
      "request_rate_limit": 30
    }
  }
}
```

## 常见工作流

### 多 Agent 协调

当多个 Agent 在同一仓库中工作时，GoZen 可以跟踪文件活动、提示告警，并帮助避免冲突。

### 会话监控

你可以通过仪表盘和 API 查看活跃会话、Token 使用量、错误次数和运行时长。

### 安全控制

护栏可以暂停失控会话、标记高风险操作，并在重试循环变得昂贵之前进行抑制。

## 相关文档

- [Agent 基础设施](/docs/agent-infrastructure) 更详细介绍新的 runtime、observatory、coordinator 和 guardrails 架构。
- [Bot 网关](/docs/bot) 说明如何从 Telegram、Slack、Discord 等聊天平台控制运行中的会话。
- [用量跟踪](/docs/usage-tracking) 和 [健康监控](/docs/health-monitoring) 介绍支撑 Agent 运维的指标。
