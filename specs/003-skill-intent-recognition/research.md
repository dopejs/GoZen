# Research: Skill-Based Intent Recognition

**Feature**: 003-skill-intent-recognition
**Date**: 2026-02-28

## R1: 混合匹配引擎架构

**Decision**: 两层匹配架构——本地快速匹配优先，LLM 回退兜底

**Rationale**:
- 本地匹配（关键词 + 同义词表 + 模糊匹配）覆盖 80%+ 的常见表达，零延迟零成本
- LLM 回退仅在本地置信度 < 0.7 时触发，处理长尾自然语言表达
- 复用现有 `LLMClient`，通过 Bot profile 路由，受 guardrails 预算控制

**Alternatives considered**:
- 纯 LLM 方案：准确率最高但每次请求都有延迟和成本，不适合高频命令场景
- 纯本地方案：零成本但难以处理复杂自然语言，准确率上限有限
- 嵌入向量匹配：需要额外依赖（向量数据库或嵌入模型），对 CLI 工具过重

## R2: 本地匹配策略

**Decision**: 基于关键词集合 + 同义词映射 + 字符串相似度的三层本地匹配

**Rationale**:
- 每个 Skill 定义一组触发关键词（多语言），支持精确匹配和包含匹配
- 同义词映射表将常见变体归一化（如"停止"→"stop"，"暂停"→"pause"）
- 字符串相似度（Levenshtein 或 Jaro-Winkler）处理拼写变体和近似表达
- 三层各自产生置信度分数，加权合并为最终本地置信度

**Alternatives considered**:
- 正则扩展：维护成本高，难以覆盖自然语言变体
- TF-IDF：需要语料库训练，对少量 Skill 场景过度设计
- 规则引擎（如 ANTLR）：引入外部依赖，学习成本高

## R3: LLM 回退的 Prompt 设计

**Decision**: 使用结构化 prompt，将所有已注册 Skill 的名称和描述作为上下文，要求 LLM 返回 JSON 格式的意图分类结果

**Rationale**:
- Prompt 包含 Skill 列表（名称 + 描述 + 关联意图），LLM 从中选择最匹配的
- 返回格式：`{"skill": "skill_name", "intent": "intent_type", "confidence": 0.85}`
- 使用 Haiku 模型（已有 `claude-3-haiku-20240307` 默认配置）降低成本和延迟
- 单次调用，不需要多轮对话

**Alternatives considered**:
- Few-shot 示例：增加 token 消耗，Skill 列表已提供足够上下文
- Function calling：增加复杂度，简单分类任务不需要

## R4: Skill 配置存储与热重载

**Decision**: Skill 配置嵌入 `BotConfig` 中，存储在 `zen.json`，通过现有 config watcher 实现热重载

**Rationale**:
- 遵循现有配置模式，无需引入新的存储机制
- daemon 已有 config file watcher（`internal/daemon/server.go`），修改 zen.json 自动触发重载
- 需要 bump config version (v9→v10)，添加迁移逻辑处理无 skills 字段的旧配置
- 内置 Skill 在代码中定义，用户自定义 Skill 在配置中定义，合并后使用

**Alternatives considered**:
- 独立 skills.json 文件：增加文件管理复杂度，需要额外的 watcher
- 数据库存储（SQLite）：已有 SQLite 用于日志，但 Skill 数量少，JSON 足够

## R5: 参数提取流程

**Decision**: Skill 匹配成功后，将意图类型注入到 LLM chat 的 system prompt 中，由 LLM 在对话处理中提取参数

**Rationale**:
- Skill 仅返回意图类型（如 IntentControl），不提取具体参数
- 对于需要参数的意图（control 需要 action/target），在 `handleChat` 流程中通过 system prompt 指导 LLM 提取
- 现有 `handleControl` 等 handler 需要 ParsedIntent 中的 action/target，Skill 匹配后需要一个轻量的参数提取步骤
- 方案：Skill 匹配成功 → 调用 LLM 提取参数（单次调用，prompt 明确指定需要的字段）→ 构造完整 ParsedIntent → 路由到对应 handler

**Alternatives considered**:
- 跳过参数提取直接进入 chat：会丢失结构化命令的精确执行能力
- 在本地匹配中提取参数：增加 Skill 定义复杂度，违反"Skill 仅负责分类"的设计决策

## R6: 匹配日志与调试

**Decision**: 使用内存环形缓冲区存储最近 N 条匹配日志，通过 Web API 暴露

**Rationale**:
- 匹配日志用于调试和优化，不需要持久化
- 环形缓冲区（默认 200 条）避免内存无限增长
- Web API 端点 `/api/v1/bot/skills/test` 提供实时测试功能
- 日志包含：输入消息、各 Skill 得分、本地/LLM 路径、最终结果、耗时

**Alternatives considered**:
- SQLite 持久化：已有日志存储但匹配日志量大且价值短暂，不值得持久化
- 文件日志：增加磁盘 I/O，清理策略复杂
