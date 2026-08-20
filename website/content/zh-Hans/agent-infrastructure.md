---
sidebar_position: 16
title: 代理基础设施 (BETA)
---

# 代理基础设施 (BETA)

:::warning BETA 功能
代理基础设施目前处于测试阶段。默认情况下已禁用，需要显式配置才能启用。
:::

内置支持自主代理工作流，包括会话管理、文件协调、实时监控和安全控制。

## 功能特性

- **代理运行时** — 执行自主代理任务，具有完整的生命周期管理
- **观测站** — 实时监控代理会话和活动
- **护栏** — 代理行为的安全控制和约束
- **协调器** — 基于文件的多代理工作流协调
- **任务队列** — 管理具有优先级和依赖关系的代理任务
- **会话管理** — 跨多个项目跟踪代理会话

## 架构

```
代理客户端 (Claude Code, Codex 等)
    ↓
代理运行时
    ↓
┌─────────────┬──────────────┬─────────────┐
│ 观测站      │ 护栏         │ 协调器      │
│ (监控)      │ (安全)       │ (同步)      │
└─────────────┴──────────────┴─────────────┘
    ↓
任务队列 → 提供商 API
```

## 配置

### 启用代理基础设施

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

## 组件

### 1. 代理运行时

管理代理任务执行生命周期。

**功能特性：**
- 任务调度和执行
- 并发任务管理
- 超时处理
- 自动清理
- 错误恢复

**配置：**
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

**API：**
```bash
# 启动代理任务
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

# 获取任务状态
GET /api/v1/agent/tasks/{task_id}

# 取消任务
DELETE /api/v1/agent/tasks/{task_id}
```

### 2. 观测站

实时监控代理活动。

**功能特性：**
- 会话跟踪
- 活动日志
- 性能指标
- 状态更新
- 历史数据

**配置：**
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

**监控指标：**
- 活跃会话
- 进行中的任务
- Token 使用量
- API 调用
- 文件操作
- 错误率
- 平均延迟

**API：**
```bash
# 获取所有活跃会话
GET /api/v1/agent/sessions

# 获取会话详情
GET /api/v1/agent/sessions/{session_id}

# 获取会话指标
GET /api/v1/agent/sessions/{session_id}/metrics
```

### 3. 护栏

代理行为的安全控制和约束。

**功能特性：**
- 操作限制
- 路径限制
- 命令阻止
- 资源配额
- 审批工作流

**配置：**
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

**执行机制：**
- 执行前验证
- 实时监控
- 自动阻止
- 审批提示
- 审计日志

**API：**
```bash
# 获取护栏状态
GET /api/v1/agent/guardrails

# 更新护栏规则
PUT /api/v1/agent/guardrails
Content-Type: application/json

{
  "max_file_operations": 200,
  "blocked_commands": ["rm -rf", "sudo", "dd"]
}
```

### 4. 协调器

基于文件的多代理工作流协调。

**功能特性：**
- 文件锁定
- 变更检测
- 冲突解决
- 状态同步
- 事件通知

**配置：**
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

**使用场景：**
- 多个代理编辑相同文件
- 防止并发修改
- 检测外部文件更改
- 协调代理工作流

**API：**
```bash
# 获取文件锁
POST /api/v1/agent/locks
Content-Type: application/json

{
  "path": "/path/to/file.go",
  "session_id": "sess_123",
  "timeout": "5m"
}

# 释放文件锁
DELETE /api/v1/agent/locks/{lock_id}

# 获取文件变更事件
GET /api/v1/agent/changes?since=2026-03-05T10:00:00Z
```

### 5. 任务队列

管理具有优先级和依赖关系的代理任务。

**功能特性：**
- 优先级调度
- 任务依赖
- 队列管理
- 状态跟踪
- 重试逻辑

**配置：**
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

**API：**
```bash
# 添加任务到队列
POST /api/v1/agent/queue
Content-Type: application/json

{
  "name": "run-tests",
  "priority": 2,
  "depends_on": ["build-project"],
  "config": {}
}

# 获取队列状态
GET /api/v1/agent/queue

# 从队列中移除任务
DELETE /api/v1/agent/queue/{task_id}
```

## Web UI

访问代理仪表板：`http://localhost:19840/agent`

### 会话标签

- **活跃会话** — 当前运行的代理会话
- **会话详情** — 任务进度、指标、日志
- **会话控制** — 暂停、恢复、取消

### 任务标签

- **任务队列** — 待处理和进行中的任务
- **任务历史** — 已完成和失败的任务
- **任务详情** — 配置、日志、结果

### 护栏标签

- **操作限制** — 当前使用量 vs. 限制
- **被阻止的操作** — 最近被阻止的尝试
- **审批队列** — 等待审批的操作

### 指标标签

- **Token 使用量** — 每个会话和总计
- **API 调用** — 请求计数和速率
- **文件操作** — 读/写/删除计数
- **性能** — 延迟和吞吐量

## 与 Claude Code 集成

GoZen 自动检测 Claude Code 会话并提供代理基础设施：

```bash
# 启动带有代理支持的 Claude Code
zen --agent

# 代理功能自动启用：
# - 会话跟踪
# - 文件协调
# - 护栏执行
# - 实时监控
```

**优势：**
- 防止并发文件修改
- 跟踪 token 使用量和成本
- 执行安全约束
- 监控代理活动
- 协调多代理工作流

## 使用场景

### 多代理开发

多个代理在同一代码库上工作：

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

### 长时间运行的任务

监控和控制长时间运行的代理任务：

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

### 安全关键操作

执行严格的安全控制：

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

## 最佳实践

1. **启用护栏** — 在生产环境中始终使用护栏
2. **设置适当的限制** — 根据使用场景配置限制
3. **主动监控** — 定期检查观测站仪表板
4. **使用文件锁定** — 为多代理工作流启用协调器
5. **配置审批** — 对破坏性操作要求审批
6. **审查日志** — 定期审计代理活动

## 限制

1. **性能开销** — 监控和协调会增加延迟
2. **文件锁定** — 在多代理场景中可能导致延迟
3. **内存使用** — 会话历史消耗内存
4. **复杂性** — 需要理解代理工作流
5. **Beta 状态** — 功能可能在未来版本中更改

## 故障排除

### 代理会话未被跟踪

1. 验证 `agent.enabled` 为 `true`
2. 检查观测站已启用
3. 确保代理客户端受支持（Claude Code、Codex）
4. 查看守护进程日志中的错误

### 文件锁定问题

1. 检查协调器已启用
2. 验证锁定超时是否合适
3. 查看活跃锁：`GET /api/v1/agent/locks`
4. 如需要，手动释放卡住的锁

### 护栏未执行

1. 验证护栏已启用
2. 检查规则配置是否正确
3. 查看被阻止的操作日志
4. 确保代理客户端遵守护栏

### 高内存使用

1. 减少历史保留期
2. 降低更新间隔
3. 限制最大并发任务数
4. 启用自动清理

## 安全考虑

1. **路径限制** — 始终配置允许/阻止的路径
2. **命令阻止** — 阻止危险命令
3. **审批工作流** — 对敏感操作要求审批
4. **审计日志** — 启用全面的日志记录
5. **资源限制** — 设置适当的操作限制

## 未来增强

- 多代理协作协议
- 高级冲突解决策略
- 用于异常检测的机器学习
- 与外部监控工具集成
- 代理行为分析
- 自动安全策略生成
