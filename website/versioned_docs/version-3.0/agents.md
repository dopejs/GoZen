---
sidebar_position: 16
title: Agent Infrastructure
---

# Agent Infrastructure

GoZen v3.0 provides infrastructure for managing AI coding agents like Claude Code and Codex.

## Overview

GoZen doesn't replace agents — it becomes the operations platform for them:

- **Agent Coordinator**: Prevent file conflicts between agents
- **Agent Observatory**: Real-time monitoring of all sessions
- **Agent Guardrails**: Safety controls and spending limits
- **Agent Task Queue**: Batch task management

## Configuration

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

## Agent Coordinator

Prevents conflicts when multiple agents work on the same codebase.

### File Locking

When an agent modifies a file, GoZen can:
1. Track which files are being edited
2. Inject warnings to other agents about locked files
3. Prevent simultaneous edits to the same file

### Change Awareness

GoZen tracks file changes and can inject context about recent modifications into agent requests.

## Agent Observatory

Real-time monitoring dashboard for all agent sessions.

### Metrics Tracked

- Current task description
- Token consumption
- Cost spent
- Request count
- Error count
- Session duration

### Stuck Detection

GoZen detects when an agent is stuck:
- Consecutive errors on the same issue
- Repeated retry patterns
- No progress indicators

### Session Control

- View all active sessions
- Kill runaway sessions
- Pause sessions at spending cap

## Agent Guardrails

Safety controls at the proxy layer.

### Spending Cap

```json
{
  "guardrails": {
    "session_spending_cap": 5.0,
    "auto_pause_on_cap": true
  }
}
```

When a session reaches the spending cap, GoZen can:
- Log a warning
- Pause the session
- Require user confirmation to continue

### Rate Limiting

Prevent agents from entering infinite retry loops:

```json
{
  "guardrails": {
    "request_rate_limit": 30
  }
}
```

### Sensitive Operation Detection

GoZen can detect and flag sensitive operations:
- File deletions
- Config file modifications
- Database operations

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/agent/config` | GET/PUT | Agent configuration |
| `/api/v1/agent/sessions` | GET | List all sessions |
| `/api/v1/agent/sessions/{id}` | GET | Session details |
| `/api/v1/agent/sessions/{id}/kill` | POST | Kill a session |
| `/api/v1/agent/locks` | GET | List file locks |
| `/api/v1/agent/changes` | GET | Recent file changes |

## Web UI

The Agent dashboard provides:

- Active sessions list with real-time metrics
- Session status indicators (active/idle/stuck)
- Kill button for each session
- Total spend across all sessions
- File locks panel
- Recent changes log
