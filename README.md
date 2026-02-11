# opencc

Claude Code 多环境切换器，支持 API 代理自动故障转移。

## 功能

- **多配置管理** — 在 `~/.cc_envs/` 中维护多个 API 配置，随时切换
- **代理故障转移** — 内置 HTTP 代理，当主 provider 不可用时自动切换到备用
- **Fallback Profiles** — 多个命名的故障转移配置，按场景快速切换（work / staging / …）
- **TUI 配置界面** — 交互式终端界面管理配置、profile 和故障转移顺序
- **自更新** — `opencc upgrade` 一键升级，支持 semver 版本匹配
- **Shell 补全** — 支持 zsh / bash / fish

## 安装

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh
```

卸载：

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh -s -- --uninstall
```

## 使用

### 创建配置

在 `~/.cc_envs/` 下创建 `.env` 文件，例如 `~/.cc_envs/work.env`：

```env
ANTHROPIC_BASE_URL=https://api.anthropic.com
ANTHROPIC_AUTH_TOKEN=sk-ant-xxx
ANTHROPIC_MODEL=claude-sonnet-4-20250514
```

也可以通过 TUI 界面创建：

```sh
opencc config
```

### 命令一览

| 命令 | 说明 |
|------|------|
| `opencc` | 以代理模式启动 claude（使用 default profile） |
| `opencc -f work` | 使用名为 "work" 的 fallback profile 启动 |
| `opencc -f` | 交互选择一个 profile 后启动 |
| `opencc use <config>` | 使用指定配置直接启动 claude |
| `opencc pick` | 交互选择 provider 启动（不保存） |
| `opencc list` | 列出所有配置（按 fallback 顺序排列） |
| `opencc config` | 打开 TUI 配置管理界面 |
| `opencc upgrade` | 升级到最新版本 |
| `opencc upgrade 1.2` | 升级到 1.2.x 最新版本 |
| `opencc version` | 显示当前版本 |
| `opencc completion zsh/bash/fish` | 生成 shell 补全脚本 |

### 故障转移

opencc 支持多个命名的 fallback profile，用于不同使用场景。

#### Profile 文件

| 文件 | Profile |
|------|---------|
| `~/.cc_envs/fallback.conf` | default（默认） |
| `~/.cc_envs/fallback.work.conf` | work |
| `~/.cc_envs/fallback.staging.conf` | staging |

每个文件格式相同，每行一个 provider 名称：

```
work
backup
personal
```

#### 使用 Profile

```sh
# 使用 default profile（等同于之前的行为）
opencc

# 使用指定 profile
opencc -f work

# 交互选择 profile
opencc -f
```

通过 `opencc config` 进入 TUI，按 `f` 键管理 fallback profiles — 可创建、编辑、删除 profile 及调整各 profile 内的 provider 顺序。

启动时 opencc 会启动一个本地 HTTP 代理，按顺序尝试各 provider。当前 provider 返回 429 或 5xx 时自动切换到下一个，并对失败的 provider 进行指数退避。

### 升级

```sh
# 升级到最新版本
opencc upgrade

# 升级到 1.x.x 最新版本
opencc upgrade 1

# 升级到 1.2.x 最新版本
opencc upgrade 1.2

# 升级到精确版本
opencc upgrade 1.2.3
```

### 配置文件说明

| 文件 | 说明 |
|------|------|
| `~/.cc_envs/*.env` | API 配置文件，每个文件对应一个 provider |
| `~/.cc_envs/fallback.conf` | default profile 的故障转移顺序 |
| `~/.cc_envs/fallback.<name>.conf` | 命名 profile 的故障转移顺序 |
| `~/.cc_envs/proxy.log` | 代理运行日志 |

每个 `.env` 文件支持以下变量：

| 变量 | 必填 | 说明 |
|------|------|------|
| `ANTHROPIC_BASE_URL` | 是 | API 地址 |
| `ANTHROPIC_AUTH_TOKEN` | 是 | API 密钥 |
| `ANTHROPIC_MODEL` | 否 | 模型名称，默认 claude-sonnet-4-20250514 |

## 开发

需要 Go 1.25+。

```sh
# 构建
go build -o opencc .

# 测试
go test ./...

# 构建当前平台二进制
./deploy.sh

# 构建所有平台二进制
./deploy.sh --all
```

发布流程：打 tag 后 GitHub Actions 自动构建并创建 Release。

```sh
git tag v1.2.0
git push origin v1.2.0
```

## 目录结构

```
├── main.go              # 入口
├── cmd/                 # CLI 命令 (cobra)
├── internal/
│   ├── config/          # fallback profile 管理
│   ├── envfile/         # .env 文件解析
│   └── proxy/           # HTTP 代理服务器
├── tui/                 # TUI 界面 (bubbletea)
├── install.sh           # 用户安装脚本
└── deploy.sh            # 构建发布脚本
```

## License

MIT
