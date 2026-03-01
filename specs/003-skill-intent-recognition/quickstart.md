# Quickstart: Skill-Based Intent Recognition

**Feature**: 003-skill-intent-recognition

## 开发环境准备

```sh
# 切换到 feature 分支
git checkout 003-skill-intent-recognition

# 构建并启动 dev daemon
./scripts/dev.sh restart

# 验证 dev daemon 运行
./scripts/dev.sh status
```

Dev 环境端口：Web UI `http://localhost:29840`，Proxy `http://localhost:29841`

## 核心开发流程

### 1. Skill 定义与匹配引擎

```sh
# TDD: 先写测试
go test ./internal/bot/ -run TestSkill -v

# 运行全部 bot 测试
go test ./internal/bot/... -v

# 检查覆盖率（需 ≥80%）
go test -cover ./internal/bot/
```

### 2. Config 变更

修改 `internal/config/config.go` 后：
- Bump `CurrentConfigVersion` (v9→v10)
- 添加迁移逻辑
- 运行 config 测试：`go test ./internal/config/ -v`

### 3. Web API

```sh
# 重启 dev daemon 加载新 API
./scripts/dev.sh restart

# 测试 Skill API
curl http://localhost:29840/api/v1/bot/skills
curl -X POST http://localhost:29840/api/v1/bot/skills/test \
  -H "Content-Type: application/json" \
  -d '{"message": "帮我暂停一下"}'
```

### 4. 全量验证

```sh
go build ./...
go test ./...
go test -cover ./internal/bot/
go test -cover ./internal/config/
```

## 关键文件

| File | Purpose |
|------|---------|
| `internal/bot/skill.go` | Skill 类型定义、注册、管理 |
| `internal/bot/matcher.go` | 混合匹配引擎 |
| `internal/bot/builtin_skills.go` | 内置 Skill 定义 |
| `internal/bot/nlu.go` | 修改：集成 Skill 匹配 |
| `internal/config/config.go` | 修改：SkillsConfig 类型 |
| `internal/web/api_bot_skills.go` | Skill Web API |
