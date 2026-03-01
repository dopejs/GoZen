# Implementation Plan: Skill-Based Intent Recognition

**Branch**: `003-skill-intent-recognition` | **Date**: 2026-02-28 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-skill-intent-recognition/spec.md`

## Summary

当前 Bot 的 NLU 系统仅依赖硬编码正则表达式进行意图识别，无法处理自然语言的多样性表达。本功能引入 Skill 系统作为正则匹配的补充层，采用混合匹配模式（本地关键词/同义词/模糊匹配优先，低置信度时回退 LLM 语义理解），提升意图识别准确率至 85% 以上。Skill 仅负责意图分类，参数提取交由已有 LLM 流程处理。

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Bubble Tea (TUI), Cobra (CLI), net/http (proxy/web)
**Storage**: JSON config at `~/.zen/zen.json`, match logs in memory (ring buffer)
**Testing**: `go test ./...`, table-driven tests, 80% coverage threshold for `internal/bot`
**Target Platform**: macOS, Linux (CLI tool)
**Project Type**: CLI tool with daemon, proxy, TUI, and Web UI
**Performance Goals**: 本地匹配 ≤500ms (95th percentile), LLM 回退 ≤2s
**Constraints**: 不破坏现有正则匹配行为，Skill 系统作为补充层
**Scale/Scope**: 预计 10-20 个内置 Skill，用户可自定义扩展

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | PASS | 新增 skill_test.go、matcher_test.go，在现有 nlu_test.go 中扩展集成测试 |
| II. Simplicity & YAGNI | PASS | Skill 匹配器作为 NLUParser 的扩展层，不引入新框架或抽象 |
| III. Config Migration Safety | PASS | 新增 SkillConfig 到 BotConfig，需 bump config version (v9→v10) |
| IV. Branch Protection & Commit Discipline | PASS | 在 feature branch 上开发，每个任务单独提交 |
| V. Minimal Artifacts | PASS | 无额外文档文件，Skill 定义在 zen.json 中 |
| VI. Test Coverage (NON-NEGOTIABLE) | PASS | internal/bot 需维持 80% 覆盖率 |

## Project Structure

### Documentation (this feature)

```text
specs/003-skill-intent-recognition/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── api-bot-skills.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── bot/
│   ├── skill.go           # Skill 定义、注册、管理
│   ├── skill_test.go      # Skill 单元测试
│   ├── matcher.go         # 混合匹配引擎（本地 + LLM 回退）
│   ├── matcher_test.go    # 匹配引擎测试
│   ├── builtin_skills.go  # 内置 Skill 定义
│   ├── nlu.go             # 修改：集成 Skill 匹配层
│   ├── nlu_test.go        # 修改：扩展测试覆盖 Skill 路径
│   ├── handlers.go        # 修改：Skill 匹配后的参数提取流程
│   └── gateway.go         # 修改：初始化 Skill 系统
├── config/
│   └── config.go          # 修改：新增 SkillConfig 类型，bump version
└── web/
    ├── api_bot_skills.go  # Skill 管理 API 端点
    └── static/
        └── app.js         # 修改：Skill 管理 UI 组件
```

**Structure Decision**: 新代码集中在 `internal/bot/` 包内，遵循现有模式。Skill 匹配引擎作为独立文件（`matcher.go`）便于测试，通过 `NLUParser` 集成到现有消息处理流程。Web API 新增 `api_bot_skills.go` 遵循现有 `api_bot.go` 模式。

## Complexity Tracking

> No constitution violations requiring justification.
