---
sidebar_position: 15
title: 中间件管道 (BETA)
---

# 中间件管道 (BETA)

:::warning BETA 功能
中间件管道目前处于测试阶段。默认情况下已禁用，需要显式配置才能启用。
:::

使用可插拔中间件扩展 GoZen，实现请求/响应转换、日志记录、速率限制和自定义处理。

## 功能特性

- **可插拔架构** — 无需修改核心代码即可添加自定义处理逻辑
- **基于优先级的执行** — 控制中间件执行顺序
- **请求/响应钩子** — 在发送前处理请求，在接收后处理响应
- **内置中间件** — 上下文注入、日志记录、速率限制、压缩
- **插件加载器** — 从本地文件或远程 URL 加载中间件
- **错误处理** — 优雅的错误处理和回退行为

## 架构

```
客户端请求
    ↓
[中间件 1: 优先级 100]
    ↓
[中间件 2: 优先级 200]
    ↓
[中间件 3: 优先级 300]
    ↓
提供商 API
    ↓
[中间件 3: 响应]
    ↓
[中间件 2: 响应]
    ↓
[中间件 1: 响应]
    ↓
客户端响应
```

## 配置

### 启用中间件管道

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "context-injection",
        "enabled": true,
        "priority": 100,
        "config": {}
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info"
        }
      }
    ]
  }
}
```

**选项：**

| 选项 | 描述 |
|------|------|
| `enabled` | 启用中间件管道 |
| `pipeline` | 中间件配置数组 |
| `name` | 中间件标识符 |
| `priority` | 执行顺序（越小越早） |
| `config` | 中间件特定配置 |

## 内置中间件

### 1. 上下文注入

向请求中注入自定义上下文。

```json
{
  "name": "context-injection",
  "enabled": true,
  "priority": 100,
  "config": {
    "system_prompt": "你是一个有用的编码助手。",
    "metadata": {
      "session_id": "sess_123",
      "user_id": "user_456"
    }
  }
}
```

**使用场景：**
- 添加系统提示
- 注入会话元数据
- 添加用户上下文

### 2. 请求日志记录器

记录所有请求和响应。

```json
{
  "name": "request-logger",
  "enabled": true,
  "priority": 200,
  "config": {
    "log_level": "info",
    "log_body": false,
    "log_headers": true
  }
}
```

**使用场景：**
- 调试
- 审计跟踪
- 性能监控

### 3. 速率限制器

限制每个提供商或全局的请求速率。

```json
{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60,
    "burst": 10,
    "per_provider": true
  }
}
```

**使用场景：**
- 防止速率限制错误
- 控制 API 使用
- 防止滥用

### 4. 压缩 (BETA)

当 token 数量超过阈值时压缩上下文。

```json
{
  "name": "compression",
  "enabled": true,
  "priority": 400,
  "config": {
    "threshold_tokens": 50000,
    "target_tokens": 20000
  }
}
```

详见[上下文压缩](./compression.md)。

### 5. 会话内存 (BETA)

跨会话维护对话记忆。

```json
{
  "name": "session-memory",
  "enabled": true,
  "priority": 150,
  "config": {
    "max_memories": 100,
    "ttl_hours": 24,
    "storage": "sqlite"
  }
}
```

**使用场景：**
- 记住用户偏好
- 跟踪对话历史
- 跨会话维护上下文

### 6. 编排 (BETA)

将请求路由到多个提供商并聚合响应。

```json
{
  "name": "orchestration",
  "enabled": true,
  "priority": 500,
  "config": {
    "strategy": "parallel",
    "providers": ["anthropic", "openai"],
    "consensus": "longest"
  }
}
```

**使用场景：**
- 比较模型输出
- 关键请求的冗余
- 通过共识提高质量

## 自定义中间件

### 中间件接口

```go
type Middleware interface {
    Name() string
    Priority() int
    ProcessRequest(ctx *RequestContext) error
    ProcessResponse(ctx *ResponseContext) error
}

type RequestContext struct {
    Provider  string
    Model     string
    Messages  []Message
    Metadata  map[string]interface{}
}

type ResponseContext struct {
    Provider  string
    Model     string
    Response  *APIResponse
    Latency   time.Duration
    Metadata  map[string]interface{}
}
```

### 示例：自定义头部注入

```go
package main

import (
    "github.com/dopejs/gozen/internal/middleware"
)

type CustomHeaderMiddleware struct {
    headers map[string]string
}

func (m *CustomHeaderMiddleware) Name() string {
    return "custom-headers"
}

func (m *CustomHeaderMiddleware) Priority() int {
    return 250
}

func (m *CustomHeaderMiddleware) ProcessRequest(ctx *middleware.RequestContext) error {
    for k, v := range m.headers {
        ctx.Metadata[k] = v
    }
    return nil
}

func (m *CustomHeaderMiddleware) ProcessResponse(ctx *middleware.ResponseContext) error {
    // 不需要响应处理
    return nil
}

func init() {
    middleware.Register("custom-headers", func(config map[string]interface{}) middleware.Middleware {
        return &CustomHeaderMiddleware{
            headers: config["headers"].(map[string]string),
        }
    })
}
```

### 加载自定义中间件

#### 本地插件

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "local",
        "path": "/path/to/custom-middleware.so",
        "config": {
          "headers": {
            "X-Custom-Header": "value"
          }
        }
      }
    ]
  }
}
```

#### 远程插件

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "remote",
        "url": "https://example.com/middleware/custom-headers.so",
        "checksum": "sha256:abc123...",
        "config": {}
      }
    ]
  }
}
```

## Web UI

在 `http://localhost:19840/settings` 访问中间件设置：

1. 导航到 "Middleware" 标签（标有 BETA 徽章）
2. 切换 "Enable Middleware Pipeline"
3. 从管道中添加/删除中间件
4. 调整优先级和配置
5. 启用/禁用单个中间件
6. 点击 "Save"

## API 端点

### 列出中间件

```bash
GET /api/v1/middleware
```

响应：
```json
{
  "enabled": true,
  "pipeline": [
    {
      "name": "context-injection",
      "enabled": true,
      "priority": 100,
      "type": "builtin"
    },
    {
      "name": "request-logger",
      "enabled": true,
      "priority": 200,
      "type": "builtin"
    }
  ]
}
```

### 添加中间件

```bash
POST /api/v1/middleware
Content-Type: application/json

{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60
  }
}
```

### 更新中间件

```bash
PUT /api/v1/middleware/{name}
Content-Type: application/json

{
  "enabled": false
}
```

### 删除中间件

```bash
DELETE /api/v1/middleware/{name}
```

### 重新加载管道

```bash
POST /api/v1/middleware/reload
```

## 使用场景

### 开发环境

添加调试日志和请求检查：

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 100,
        "config": {
          "log_level": "debug",
          "log_body": true
        }
      }
    ]
  }
}
```

### 生产环境

添加速率限制和监控：

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "rate-limiter",
        "enabled": true,
        "priority": 100,
        "config": {
          "requests_per_minute": 100,
          "burst": 20
        }
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info",
          "log_body": false
        }
      }
    ]
  }
}
```

### 多提供商比较

使用编排来比较输出：

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "orchestration",
        "enabled": true,
        "priority": 500,
        "config": {
          "strategy": "parallel",
          "providers": ["anthropic", "openai", "google"],
          "consensus": "longest"
        }
      }
    ]
  }
}
```

## 最佳实践

1. **使用适当的优先级** — 较小的数字先执行
2. **保持中间件专注** — 每个中间件应该做好一件事
3. **优雅地处理错误** — 不要因错误而破坏管道
4. **彻底测试** — 在生产前验证中间件行为
5. **监控性能** — 跟踪中间件开销
6. **记录配置** — 清楚地记录配置选项

## 限制

1. **性能开销** — 每个中间件都会增加延迟
2. **复杂性** — 太多中间件会使调试变得困难
3. **插件安全** — 远程插件需要信任和验证
4. **错误传播** — 中间件错误会影响所有请求
5. **配置复杂性** — 复杂的管道更难维护

## 故障排除

### 中间件未执行

1. 验证 `middleware.enabled` 为 `true`
2. 检查中间件在管道中已启用
3. 验证优先级设置正确
4. 查看守护进程日志中的中间件错误

### 意外行为

1. 检查中间件执行顺序（优先级）
2. 验证配置是否正确
3. 单独测试中间件
4. 查看中间件日志

### 性能问题

1. 识别慢速中间件（检查日志）
2. 减少中间件数量
3. 优化中间件实现
4. 考虑禁用非必要的中间件

### 插件加载失败

1. 验证插件路径是否正确
2. 检查插件是否为正确的架构编译
3. 验证校验和匹配（对于远程插件）
4. 查看插件日志中的错误

## 安全考虑

1. **验证插件** — 仅加载受信任的插件
2. **验证校验和** — 始终验证远程插件校验和
3. **沙箱插件** — 考虑在隔离环境中运行插件
4. **审计中间件** — 在部署前审查中间件代码
5. **监控行为** — 注意意外的中间件行为

## 未来增强

- WebAssembly 插件支持以实现跨平台兼容性
- 用于共享社区插件的中间件市场
- Web UI 中的可视化管道编辑器
- 中间件性能分析
- 插件更新的热重载
- 中间件测试框架
