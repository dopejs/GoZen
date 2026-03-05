---
sidebar_position: 16
title: Agent Infrastructure (BETA)
---

# Agent Infrastructure (BETA)

:::warning BETA Feature
Agent infrastructure is currently in beta. It is disabled by default and requires explicit configuration to enable.
:::

Built-in support for autonomous agent workflows with session management, file coordination, real-time monitoring, and safety controls.

## Features

- **Agent Runtime** — Execute autonomous agent tasks with full lifecycle management
- **Observatory** — Real-time monitoring of agent sessions and activities
- **Guardrails** — Safety controls and constraints for agent behavior
- **Coordinator** — File-based coordination for multi-agent workflows
- **Task Queue** — Manage agent tasks with priority and dependencies
- **Session Management** — Track agent sessions across multiple projects

## Architecture

```
Agent Client (Claude Code, Codex, etc.)
    ↓
Agent Runtime
    ↓
┌─────────────┬──────────────┬─────────────┐
│ Observatory │ Guardrails   │ Coordinator │
│ (Monitor)   │ (Safety)     │ (Sync)      │
└─────────────┴──────────────┴─────────────┘
    ↓
Task Queue → Provider API
```

## Configuration

### Enable Agent Infrastructure

```json
{
  "agent": {
    "enabled": true,
    "runtime": {
      "max_concurrent_tasks": 5,
      "task_timeout": "30m",
      "auto_cleanup": true
    },
    "observatory": {
      "enabled": true,
      "update_interval": "5s",
      "history_retention": "7d"
    },
    "guardrails": {
      "enabled": true,
      "max_file_operations": 100,
      "max_api_calls": 1000,
      "allowed_paths": ["/Users/john/projects"],
      "blocked_commands": ["rm -rf", "sudo"]
    },
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    }
  }
}
```

## Components

### 1. Agent Runtime

Manages agent task execution lifecycle.

**Features:**
- Task scheduling and execution
- Concurrent task management
- Timeout handling
- Automatic cleanup
- Error recovery

**Configuration:**
```json
{
  "runtime": {
    "max_concurrent_tasks": 5,
    "task_timeout": "30m",
    "auto_cleanup": true,
    "retry_failed_tasks": true,
    "max_retries": 3
  }
}
```

**API:**
```bash
# Start agent task
POST /api/v1/agent/tasks
Content-Type: application/json

{
  "name": "code-review",
  "description": "Review pull request #123",
  "priority": 1,
  "config": {
    "model": "claude-opus-4",
    "max_tokens": 100000
  }
}

# Get task status
GET /api/v1/agent/tasks/{task_id}

# Cancel task
DELETE /api/v1/agent/tasks/{task_id}
```

### 2. Observatory

Real-time monitoring of agent activities.

**Features:**
- Session tracking
- Activity logging
- Performance metrics
- Status updates
- Historical data

**Configuration:**
```json
{
  "observatory": {
    "enabled": true,
    "update_interval": "5s",
    "history_retention": "7d",
    "metrics": {
      "track_tokens": true,
      "track_costs": true,
      "track_latency": true
    }
  }
}
```

**Monitored Metrics:**
- Active sessions
- Tasks in progress
- Token usage
- API calls
- File operations
- Error rate
- Average latency

**API:**
```bash
# Get all active sessions
GET /api/v1/agent/sessions

# Get session details
GET /api/v1/agent/sessions/{session_id}

# Get session metrics
GET /api/v1/agent/sessions/{session_id}/metrics
```

### 3. Guardrails

Safety controls and constraints for agent behavior.

**Features:**
- Operation limits
- Path restrictions
- Command blocking
- Resource quotas
- Approval workflows

**Configuration:**
```json
{
  "guardrails": {
    "enabled": true,
    "max_file_operations": 100,
    "max_api_calls": 1000,
    "max_tokens_per_session": 1000000,
    "allowed_paths": [
      "/Users/john/projects",
      "/tmp/agent-workspace"
    ],
    "blocked_paths": [
      "/etc",
      "/System",
      "~/.ssh"
    ],
    "blocked_commands": [
      "rm -rf /",
      "sudo",
      "chmod 777"
    ],
    "require_approval": {
      "file_delete": true,
      "system_commands": true,
      "network_requests": false
    }
  }
}
```

**Enforcement:**
- Pre-execution validation
- Real-time monitoring
- Automatic blocking
- Approval prompts
- Audit logging

**API:**
```bash
# Get guardrail status
GET /api/v1/agent/guardrails

# Update guardrail rules
PUT /api/v1/agent/guardrails
Content-Type: application/json

{
  "max_file_operations": 200,
  "blocked_commands": ["rm -rf", "sudo", "dd"]
}
```

### 4. Coordinator

File-based coordination for multi-agent workflows.

**Features:**
- File locking
- Change detection
- Conflict resolution
- State synchronization
- Event notifications

**Configuration:**
```json
{
  "coordinator": {
    "enabled": true,
    "lock_timeout": "5m",
    "change_detection": true,
    "conflict_resolution": "last-write-wins",
    "notification_webhook": "https://hooks.slack.com/..."
  }
}
```

**Use Cases:**
- Multiple agents editing same files
- Preventing concurrent modifications
- Detecting external file changes
- Coordinating agent workflows

**API:**
```bash
# Acquire file lock
POST /api/v1/agent/locks
Content-Type: application/json

{
  "path": "/path/to/file.go",
  "session_id": "sess_123",
  "timeout": "5m"
}

# Release file lock
DELETE /api/v1/agent/locks/{lock_id}

# Get file change events
GET /api/v1/agent/changes?since=2026-03-05T10:00:00Z
```

### 5. Task Queue

Manage agent tasks with priority and dependencies.

**Features:**
- Priority scheduling
- Task dependencies
- Queue management
- Status tracking
- Retry logic

**Configuration:**
```json
{
  "task_queue": {
    "enabled": true,
    "max_queue_size": 100,
    "priority_levels": 5,
    "enable_dependencies": true,
    "retry_policy": {
      "max_retries": 3,
      "backoff": "exponential"
    }
  }
}
```

**API:**
```bash
# Add task to queue
POST /api/v1/agent/queue
Content-Type: application/json

{
  "name": "run-tests",
  "priority": 2,
  "depends_on": ["build-project"],
  "config": {}
}

# Get queue status
GET /api/v1/agent/queue

# Remove task from queue
DELETE /api/v1/agent/queue/{task_id}
```

## Web UI

Access agent dashboard at `http://localhost:19840/agent`:

### Sessions Tab

- **Active sessions** — Currently running agent sessions
- **Session details** — Task progress, metrics, logs
- **Session controls** — Pause, resume, cancel

### Tasks Tab

- **Task queue** — Pending and in-progress tasks
- **Task history** — Completed and failed tasks
- **Task details** — Configuration, logs, results

### Guardrails Tab

- **Operation limits** — Current usage vs. limits
- **Blocked operations** — Recent blocked attempts
- **Approval queue** — Operations awaiting approval

### Metrics Tab

- **Token usage** — Per session and total
- **API calls** — Request count and rate
- **File operations** — Read/write/delete counts
- **Performance** — Latency and throughput

## Integration with Claude Code

GoZen automatically detects Claude Code sessions and provides agent infrastructure:

```bash
# Start Claude Code with agent support
zen --agent

# Agent features are automatically enabled:
# - Session tracking
# - File coordination
# - Guardrails enforcement
# - Real-time monitoring
```

**Benefits:**
- Prevent concurrent file modifications
- Track token usage and costs
- Enforce safety constraints
- Monitor agent activities
- Coordinate multi-agent workflows

## Use Cases

### Multi-Agent Development

Multiple agents working on same codebase:

```json
{
  "agent": {
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    },
    "guardrails": {
      "max_file_operations": 200,
      "allowed_paths": ["/Users/john/project"]
    }
  }
}
```

### Long-Running Tasks

Monitor and control long-running agent tasks:

```json
{
  "agent": {
    "runtime": {
      "task_timeout": "2h",
      "auto_cleanup": false
    },
    "observatory": {
      "update_interval": "10s",
      "history_retention": "30d"
    }
  }
}
```

### Safety-Critical Operations

Enforce strict safety controls:

```json
{
  "agent": {
    "guardrails": {
      "enabled": true,
      "max_file_operations": 50,
      "blocked_commands": ["rm", "sudo", "chmod"],
      "require_approval": {
        "file_delete": true,
        "system_commands": true,
        "network_requests": true
      }
    }
  }
}
```

## Best Practices

1. **Enable guardrails** — Always use guardrails in production
2. **Set appropriate limits** — Configure limits based on use case
3. **Monitor actively** — Check observatory dashboard regularly
4. **Use file locking** — Enable coordinator for multi-agent workflows
5. **Configure approvals** — Require approval for destructive operations
6. **Review logs** — Audit agent activities regularly

## Limitations

1. **Performance overhead** — Monitoring and coordination add latency
2. **File locking** — Can cause delays in multi-agent scenarios
3. **Memory usage** — Session history consumes memory
4. **Complexity** — Requires understanding of agent workflows
5. **Beta status** — Features may change in future releases

## Troubleshooting

### Agent session not tracked

1. Verify `agent.enabled` is `true`
2. Check observatory is enabled
3. Ensure agent client is supported (Claude Code, Codex)
4. Review daemon logs for errors

### File locking issues

1. Check coordinator is enabled
2. Verify lock timeout is appropriate
3. Review active locks: `GET /api/v1/agent/locks`
4. Manually release stuck locks if needed

### Guardrails not enforcing

1. Verify guardrails are enabled
2. Check rule configuration is correct
3. Review blocked operations log
4. Ensure agent client respects guardrails

### High memory usage

1. Reduce history retention period
2. Decrease update interval
3. Limit max concurrent tasks
4. Enable auto cleanup

## Security Considerations

1. **Path restrictions** — Always configure allowed/blocked paths
2. **Command blocking** — Block dangerous commands
3. **Approval workflows** — Require approval for sensitive operations
4. **Audit logging** — Enable comprehensive logging
5. **Resource limits** — Set appropriate operation limits

## Future Enhancements

- Multi-agent collaboration protocols
- Advanced conflict resolution strategies
- Machine learning for anomaly detection
- Integration with external monitoring tools
- Agent behavior analytics
- Automated safety policy generation
