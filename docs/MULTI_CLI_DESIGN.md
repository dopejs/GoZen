# Multi-CLI & Provider Transform Design

## 研究结论

### CLI API 格式对比

| CLI | API 格式 | 环境变量 | 备注 |
|-----|---------|---------|------|
| Claude Code | Anthropic Messages API | `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_MODEL` | 当前支持 |
| Codex | OpenAI Chat Completions API | `OPENAI_API_KEY`, `OPENAI_BASE_URL` | 需要 API 转换 |
| OpenCode | 多 provider 支持 | `ANTHROPIC_API_KEY`, `OPENAI_API_KEY` 等 | 原生支持 Anthropic |

### 关键发现

1. **Codex 使用 OpenAI API 格式**
   - 需要 OpenAI Chat Completions API 兼容的后端
   - 环境变量：`OPENAI_API_KEY`, `OPENAI_BASE_URL`
   - 如果要让 Codex 使用 Anthropic provider，需要 API 转换层

2. **OpenCode 原生支持多 provider**
   - 支持 Anthropic、OpenAI、Google 等 75+ providers
   - 使用 `anthropic/claude-sonnet-4-5` 格式指定模型
   - 可以直接使用 Anthropic API，无需转换

3. **Feature 1 和 Feature 2 的关系**
   - **有关系但不完全依赖**
   - OpenCode 可以直接使用 Anthropic provider，不需要转换
   - Codex 需要 OpenAI 格式，如果要用 Anthropic provider 则需要转换
   - Provider Transform 主要服务于：Codex + Anthropic provider 的组合

---

## Feature 1: Provider Request Transform

### 目标

支持不同 API 格式的 provider，让 opencc 可以：
- 将 Anthropic API 请求转换为 OpenAI 格式（供 Codex 使用）
- 将 OpenAI API 请求转换为 Anthropic 格式（供 Claude Code 使用）

### 设计

#### 1. Provider Type 字段

```go
type ProviderConfig struct {
    Type string `json:"type,omitempty"` // "anthropic", "openai", "azure", "bedrock"
    // ... existing fields
}
```

支持的类型：
- `anthropic` - Anthropic Messages API（默认）
- `openai` - OpenAI Chat Completions API
- `azure` - Azure OpenAI（OpenAI 格式 + Azure 认证）
- `bedrock` - AWS Bedrock（需要 AWS 签名）

#### 2. Transform 层

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────>│   Proxy     │────>│  Provider   │
│ (CLI 请求)   │     │ (Transform) │     │ (上游 API)  │
└─────────────┘     └─────────────┘     └─────────────┘
                          │
                    ┌─────┴─────┐
                    │ Transform │
                    │  Router   │
                    └───────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
    ┌──────────┐   ┌──────────┐   ┌──────────┐
    │ Anthropic│   │  OpenAI  │   │  Azure   │
    │ Transform│   │ Transform│   │ Transform│
    └──────────┘   └──────────┘   └──────────┘
```

#### 3. 转换逻辑

**Anthropic → OpenAI:**
```json
// Anthropic 请求
{
  "model": "claude-sonnet-4-5",
  "messages": [{"role": "user", "content": "Hello"}],
  "max_tokens": 1024
}

// 转换为 OpenAI 格式
{
  "model": "claude-sonnet-4-5",
  "messages": [{"role": "user", "content": "Hello"}],
  "max_completion_tokens": 1024
}
```

**OpenAI → Anthropic:**
```json
// OpenAI 请求
{
  "model": "gpt-4",
  "messages": [{"role": "user", "content": "Hello"}],
  "max_completion_tokens": 1024
}

// 转换为 Anthropic 格式
{
  "model": "gpt-4",
  "messages": [{"role": "user", "content": "Hello"}],
  "max_tokens": 1024
}
```

#### 4. 实现文件

```
internal/proxy/
├── transform/
│   ├── transform.go      # Transform 接口
│   ├── anthropic.go      # Anthropic 格式处理
│   ├── openai.go         # OpenAI 格式处理
│   ├── azure.go          # Azure 格式处理
│   └── bedrock.go        # Bedrock 格式处理
└── server.go             # 集成 transform
```

---

## Feature 2: Multi-CLI Support

### 目标

支持启动不同的 AI coding CLI：
- Claude Code（当前默认）
- Codex（OpenAI）
- OpenCode（多 provider）

### 设计

#### 1. CLI 配置

已实现：`default_cli` 字段

```json
{
  "default_cli": "claude"  // "claude", "codex", "opencode"
}
```

#### 2. CLI 环境变量映射

不同 CLI 需要不同的环境变量：

| CLI | 需要的环境变量 |
|-----|---------------|
| Claude Code | `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_MODEL` |
| Codex | `OPENAI_API_KEY`, `OPENAI_BASE_URL` |
| OpenCode | `ANTHROPIC_API_KEY` 或 `OPENAI_API_KEY`（取决于 provider） |

#### 3. 启动逻辑

```go
func startCLI(cli string, providers []*Provider) error {
    switch cli {
    case "claude":
        // 设置 ANTHROPIC_* 环境变量
        // 启动代理（Anthropic 格式）
        // 启动 claude
    case "codex":
        // 设置 OPENAI_* 环境变量
        // 启动代理（需要 Anthropic→OpenAI 转换）
        // 启动 codex
    case "opencode":
        // 设置对应 provider 的环境变量
        // 启动代理（根据 provider type）
        // 启动 opencode
    }
}
```

#### 4. 代理模式

| CLI | Provider Type | 代理行为 |
|-----|--------------|---------|
| claude | anthropic | 直接转发 |
| claude | openai | OpenAI→Anthropic 转换 |
| codex | anthropic | Anthropic→OpenAI 转换 |
| codex | openai | 直接转发 |
| opencode | any | 根据 provider type 处理 |

---

## 实现计划

### Phase 1: Provider Type 基础设施

1. 添加 `Type` 字段到 `ProviderConfig`
2. TUI/Web UI 支持选择 provider type
3. 配置迁移（默认 type = "anthropic"）

### Phase 2: Transform 层

1. 定义 Transform 接口
2. 实现 Anthropic transform（当前逻辑提取）
3. 实现 OpenAI transform
4. 集成到 proxy server

### Phase 3: CLI 环境变量映射

1. 定义 CLI 环境变量映射表
2. 根据 CLI + Provider Type 设置正确的环境变量
3. 更新启动逻辑

### Phase 4: 测试与文档

1. 单元测试 transform 逻辑
2. 集成测试各 CLI + Provider 组合
3. 更新 README 和配置文档

---

## 优先级建议

1. **高优先级**：OpenCode 支持（无需 transform，只需环境变量映射）
2. **中优先级**：Provider Type 字段和 UI
3. **低优先级**：Codex 支持（需要完整 transform 层）

OpenCode 原生支持 Anthropic，可以快速实现。Codex 需要 API 转换，工作量较大。

---

*Created: 2026-02-14*
