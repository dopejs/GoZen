---
title: Agents
---

# Agents

GoZen can act as an operations layer for coding agents such as Claude Code, Codex, and other CLI-driven assistants. It helps you coordinate agent work, monitor sessions, and apply runtime safety controls without changing your existing workflow.

## What GoZen adds

- **Coordination**: Reduce conflicts when multiple agents touch the same project.
- **Observability**: Track sessions, costs, errors, and activity in one place.
- **Guardrails**: Apply limits around spend, request rate, and sensitive actions.
- **Task routing**: Send different work to different providers or profiles.

## Example configuration

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

## Common workflows

### Multi-agent coordination

When several agents work in the same repository, GoZen can track file activity, surface warnings, and make collisions easier to avoid.

### Session monitoring

Use the dashboard and APIs to inspect active sessions, token usage, error counts, and runtime duration.

### Safety enforcement

Guardrails can pause runaway sessions, flag risky operations, and slow down retry loops before they become expensive.

## Related docs

- [Agent Infrastructure](/docs/agent-infrastructure) covers the newer runtime, observatory, coordinator, and guardrail architecture in more detail.
- [Bot Gateway](/docs/bot) explains how to control running sessions from Telegram, Slack, Discord, and other chat platforms.
- [Usage Tracking](/docs/usage-tracking) and [Health Monitoring](/docs/health-monitoring) cover the metrics that power agent operations.
