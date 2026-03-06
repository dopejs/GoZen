---
sidebar_position: 14
title: 上下文压缩 (BETA)
---

# 上下文压缩 (BETA)

:::warning BETA 功能
上下文压缩目前处于测试阶段。默认情况下已禁用，需要显式配置才能启用。
:::

当 token 数量超过阈值时自动压缩对话上下文，在保持对话质量的同时降低成本。

## 功能特性

- **自动压缩** — 当 token 数量超过阈值时触发
- **智能摘要** — 使用廉价模型（claude-3-haiku）总结旧消息
- **保留最近消息** — 保持最近的消息完整以保证上下文连续性
- **Token 估算** — 在 API 调用前准确计算 token 数量
- **统计跟踪** — 监控压缩效果
- **透明操作** — 与所有 AI 客户端无缝协作

## 工作原理

1. **Token 估算** — 计算对话历史中的 token 数量
2. **阈值检查** — 与配置的阈值比较（默认：50,000）
3. **消息选择** — 识别需要压缩的旧消息
4. **摘要生成** — 使用廉价模型创建简洁摘要
5. **上下文替换** — 用摘要替换旧消息
6. **请求转发** — 将压缩后的上下文发送到目标模型

## 配置

### 启用压缩

```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 50000,
    "target_tokens": 20000,
    "summarizer_model": "claude-3-haiku-20240307",
    "preserve_recent_messages": 5,
    "tokens_per_char": 0.25
  }
}
```

**选项：**

| 选项 | 默认值 | 描述 |
|------|--------|------|
| `enabled` | `false` | 启用上下文压缩 |
| `threshold_tokens` | `50000` | 当上下文超过此值时触发压缩 |
| `target_tokens` | `20000` | 压缩后的目标 token 数量 |
| `summarizer_model` | `claude-3-haiku-20240307` | 用于摘要的模型 |
| `preserve_recent_messages` | `5` | 保持完整的最近消息数量 |
| `tokens_per_char` | `0.25` | Token 计数的估算比率 |

### 按配置文件配置

为特定配置文件启用压缩：

```json
{
  "profiles": {
    "long-context": {
      "providers": ["anthropic"],
      "compression": {
        "enabled": true,
        "threshold_tokens": 100000,
        "target_tokens": 40000
      }
    },
    "short-context": {
      "providers": ["openai"],
      "compression": {
        "enabled": false
      }
    }
  }
}
```

## Token 估算

GoZen 使用基于字符的估算进行快速 token 计数：

```
estimated_tokens = character_count * tokens_per_char
```

**默认比率：** 每字符 0.25 个 token（1 个 token ≈ 4 个字符）

**准确度：** 英文文本 ±10%，其他语言可能有所不同

对于精确的 token 计数，GoZen 在可用时使用 `tiktoken-go` 库。

## 压缩策略

### 消息选择

1. **系统消息** — 始终保留
2. **最近消息** — 保留最后 N 条消息（默认：5）
3. **旧消息** — 压缩候选

### 摘要提示

```
简洁地总结以下对话历史，同时保留关键信息、决策和上下文：

[旧消息]

提供一个捕捉要点的简短摘要。
```

### 结果

```
原始：45,000 tokens（30 条消息）
压缩后：22,000 tokens（摘要 + 5 条最近消息）
节省：23,000 tokens（51%）
```

## Web UI

在 `http://localhost:19840/settings` 访问压缩设置：

1. 导航到 "Compression" 标签（标有 BETA 徽章）
2. 切换 "Enable Compression"
3. 调整阈值和目标 token 数量
4. 选择摘要模型
5. 设置要保留的最近消息数量
6. 点击 "Save"

### 统计仪表板

查看压缩统计：

- **总压缩次数** — 触发压缩的次数
- **节省的 Token** — 所有压缩中节省的总 token 数
- **平均节省** — 每次压缩的平均 token 减少量
- **压缩率** — 触发压缩的请求百分比

## API 端点

### 获取压缩统计

```bash
GET /api/v1/compression/stats
```

响应：
```json
{
  "enabled": true,
  "total_compressions": 42,
  "tokens_saved": 1250000,
  "average_savings": 29761,
  "compression_rate": 0.15,
  "last_compression": "2026-03-05T10:30:00Z"
}
```

### 更新压缩设置

```bash
PUT /api/v1/compression/settings
Content-Type: application/json

{
  "enabled": true,
  "threshold_tokens": 60000,
  "target_tokens": 25000
}
```

### 重置统计

```bash
POST /api/v1/compression/stats/reset
```

## 使用场景

### 长时间编码会话

**场景：** 使用 Claude Code 进行多小时编码会话

**配置：**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 80000,
    "target_tokens": 30000,
    "preserve_recent_messages": 10
  }
}
```

**优势：** 在不触及上下文限制的情况下保持对话连续性

### 批量处理

**场景：** 使用 AI 处理多个文档

**配置：**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 40000,
    "target_tokens": 15000,
    "preserve_recent_messages": 3
  }
}
```

**优势：** 在处理大型文档集时降低成本

### 研究与分析

**场景：** 涉及多个主题的长时间研究会话

**配置：**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 100000,
    "target_tokens": 40000,
    "preserve_recent_messages": 8
  }
}
```

**优势：** 保持对话专注于最近的主题，同时保留早期上下文

## 最佳实践

1. **从默认值开始** — 默认设置适用于大多数使用场景
2. **监控统计** — 定期检查压缩率和节省情况
3. **调整阈值** — 对于长上下文模型（Claude Opus）增加，对于短上下文减少
4. **保留足够的消息** — 保留 5-10 条最近消息以保证上下文连续性
5. **使用廉价摘要器** — Haiku 快速且成本效益高，适合摘要
6. **生产前测试** — 使用您的特定用例验证压缩质量

## 限制

1. **质量损失** — 摘要可能会丢失细微的细节
2. **延迟增加** — 增加摘要 API 调用开销
3. **成本权衡** — 摘要成本 vs. token 节省
4. **语言支持** — 最适合英语，其他语言可能有所不同
5. **上下文窗口** — 不能超过模型的最大上下文窗口

## 故障排除

### 压缩未触发

1. 验证 `compression.enabled` 为 `true`
2. 检查 token 数量是否超过阈值
3. 确保对话有足够的消息可压缩
4. 查看守护进程日志中的压缩错误

### 摘要质量差

1. 尝试不同的摘要模型（例如 claude-3-sonnet）
2. 增加 `preserve_recent_messages` 以保留更多上下文
3. 调整 `target_tokens` 以允许更长的摘要
4. 检查摘要模型是否可用且正常工作

### 延迟增加

1. 压缩会增加一次额外的 API 调用（摘要）
2. 使用更快的摘要模型（haiku 最快）
3. 增加阈值以减少压缩频率
4. 考虑对延迟敏感的应用禁用压缩

### 意外成本

1. 在使用仪表板中监控摘要成本
2. 比较节省 vs. 摘要成本
3. 调整阈值以减少压缩频率
4. 使用最便宜的可用模型进行摘要

## 性能影响

- **Token 估算** — 每个请求约 1ms（可忽略）
- **摘要生成** — 1-3 秒（取决于模型和消息数量）
- **内存开销** — 最小（每次压缩约 1KB）
- **成本节省** — 通常减少 30-50% 的 token

## 高级配置

### 自定义摘要提示

```json
{
  "compression": {
    "enabled": true,
    "custom_prompt": "创建以下对话的技术摘要，重点关注代码更改、决策和行动项：\n\n{messages}\n\n摘要："
  }
}
```

### 条件压缩

仅为特定场景启用压缩：

```json
{
  "profiles": {
    "default": {
      "scenarios": {
        "longContext": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": true,
            "threshold_tokens": 100000
          }
        },
        "default": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": false
          }
        }
      }
    }
  }
}
```

### 多阶段压缩

对于非常长的对话进行多次压缩：

```json
{
  "compression": {
    "enabled": true,
    "stages": [
      {
        "threshold_tokens": 50000,
        "target_tokens": 30000
      },
      {
        "threshold_tokens": 80000,
        "target_tokens": 40000
      }
    ]
  }
}
```

## 未来增强

- 用于智能消息选择的语义相似度匹配
- 用于质量比较的多模型摘要
- 压缩质量指标和反馈
- 针对每个用例的自定义压缩策略
- 与 RAG 集成以进行外部上下文存储
