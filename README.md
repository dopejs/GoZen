# opencc

Claude Code 多环境切换器，支持 API 代理自动故障转移。

## 功能

- **多配置管理** — 在 `~/.opencc/opencc.json` 中统一管理所有 API 配置，随时切换
- **代理故障转移** — 内置 HTTP 代理，当主 provider 不可用时自动切换到备用
- **场景路由** — 根据请求特征（thinking、image、longContext、webSearch、background）智能路由到不同 provider
- **环境变量配置** — 在 provider 级别配置 Claude Code 环境变量（max_output_tokens、effort_level 等）
- **Fallback Profiles** — 多个命名的故障转移配置，按场景快速切换（work / staging / …）
- **TUI 配置界面** — 交互式终端界面管理配置、profile 和故障转移顺序（v1.5 全新设计）
- **Web 管理界面** — 浏览器可视化管理 provider 和 profile，支持拖拽排序
- **全局设置** — 配置默认 Profile、默认 CLI、Web UI 端口
- **智能 Profile 分配** — 添加 provider 后自动弹出 profile 选择（TUI 和 Web）
- **模型自动补全** — Web 端模型字段带有官方 Claude Model ID 候选提示
- **自更新** — `opencc upgrade` 一键升级，支持 semver 版本匹配，带下载进度条
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

通过 TUI 界面创建：

```sh
opencc config
```

或手动编辑 `~/.opencc/opencc.json`：

```json
{
  "providers": {
    "work": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "reasoning_model": "claude-sonnet-4-5",
      "haiku_model": "claude-haiku-4-5",
      "opus_model": "claude-opus-4-5",
      "sonnet_model": "claude-sonnet-4-5"
    },
    "backup": {
      "base_url": "https://backup.example.com",
      "auth_token": "sk-..."
    }
  },
  "profiles": {
    "default": ["work", "backup"],
    "staging": ["staging-provider"]
  }
}
```

### 命令一览

| 命令 | 说明 |
|------|------|
| `opencc` | 以代理模式启动 claude（使用 default profile 或项目绑定） |
| `opencc -f work` | 使用名为 "work" 的 fallback profile 启动 |
| `opencc -f` | 交互选择一个 profile 后启动 |
| `opencc use <config>` | 使用指定配置直接启动 claude |
| `opencc pick` | 交互选择 provider 启动（不保存） |
| `opencc list` | 列出所有配置（按 fallback 顺序排列） |
| `opencc config` | 打开 TUI 配置管理界面（主菜单） |
| `opencc config --legacy` | 使用旧版 TUI 界面 |
| `opencc config add provider [name]` | 添加 provider（可预填名称） |
| `opencc config add profile [name]` | 添加 profile（可预填名称） |
| `opencc bind <profile>` | 绑定当前目录到指定 profile |
| `opencc unbind` | 解除当前目录的 profile 绑定 |
| `opencc status` | 显示当前目录的绑定状态 |
| `opencc web start` | 启动 Web 管理界面（后台守护进程） |
| `opencc web stop` | 停止 Web 管理界面 |
| `opencc web open` | 在浏览器中打开 Web 管理界面 |
| `opencc upgrade` | 升级到最新版本 |
| `opencc upgrade 1.2` | 升级到 1.2.x 最新版本 |
| `opencc version` | 显示当前版本 |
| `opencc completion zsh/bash/fish` | 生成 shell 补全脚本 |

### 故障转移

opencc 支持多个命名的 fallback profile，用于不同使用场景。

Profile 配置在 `~/.opencc/opencc.json` 的 `profiles` 字段中：

```json
{
  "profiles": {
    "default": ["work", "backup", "personal"],
    "work": ["work-primary", "work-secondary"],
    "staging": ["staging-provider"]
  }
}
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

### 项目级 Profile 绑定

opencc 支持将目录绑定到特定 profile，实现项目级别的自动配置切换。

#### 使用方式

```sh
# 在项目目录下绑定 profile
cd /path/to/project
opencc bind work-profile

# 之后在该目录运行 opencc 会自动使用 work-profile
opencc

# 查看当前目录的绑定状态
opencc status

# 解除绑定
opencc unbind
```

#### 工作原理

- 绑定信息存储在 `~/.opencc/opencc.json` 的 `project_bindings` 字段
- 每个用户有自己的绑定配置（不会提交到项目仓库）
- 优先级：`-f` 参数 > 项目绑定 > default profile

#### 使用场景

```sh
# 工作项目使用公司 API
cd ~/work/company-project
opencc bind work-profile
opencc  # 自动使用 work-profile

# 个人项目使用个人 API
cd ~/personal/side-project
opencc bind personal-profile
opencc  # 自动使用 personal-profile

# 临时使用其他 profile（覆盖绑定）
opencc -f staging-profile
```

#### 配置示例

```json
{
  "providers": {...},
  "profiles": {
    "work-profile": ["work-api"],
    "personal-profile": ["personal-api"]
  },
  "project_bindings": {
    "/Users/john/work/company-project": "work-profile",
    "/Users/john/personal/side-project": "personal-profile"
  }
}
```

**注意事项**：
- 如果绑定的 profile 被删除，会自动降级到 default profile
- 绑定使用绝对路径，确保跨 shell 会话一致性

### Web 管理界面

除了 TUI，opencc 还提供浏览器管理界面：

```sh
# 启动 Web 服务（后台守护进程，端口 19840）
opencc web start

# 在浏览器中打开
opencc web open

# 停止
opencc web stop
```

Web UI 支持：
- Provider 和 Profile 的增删改查
- 拖拽调整 provider 排序
- 添加 provider 后选择要加入的 profile
- 模型字段自动补全（Claude 官方 Model ID）

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
| `~/.opencc/opencc.json` | 统一 JSON 配置文件（providers + profiles） |
| `~/.opencc/proxy.log` | 代理运行日志 |
| `~/.opencc/web.log` | Web 服务运行日志 |

#### 配置文件版本管理

从 v1.3.2 开始，配置文件包含 `version` 字段用于版本管理：

```json
{
  "version": 2,
  "providers": {...},
  "profiles": {...}
}
```

**版本兼容性**：
- ✅ 新版本可以读取老版本配置（自动升级）
- ❌ 老版本无法读取新版本配置（会提示升级）
- 建议所有机器使用相同版本的 opencc

**版本历史**：
- Version 1（隐式）：v1.3.1 及之前，profiles 为字符串数组
- Version 2：v1.3.2+，profiles 支持 routing 和 long_context_threshold
- Version 3：v1.4.0+，project_bindings 支持
- Version 4：v1.5.0+，default_profile、default_cli、web_port 全局设置

每个 provider 支持以下字段：

| 字段 | 必填 | 说明 |
|------|------|------|
| `base_url` | 是 | API 地址 |
| `auth_token` | 是 | API 密钥 |
| `model` | 否 | 主模型，默认 `claude-sonnet-4-5` |
| `reasoning_model` | 否 | 推理模型，默认 `claude-sonnet-4-5` |
| `haiku_model` | 否 | Haiku 模型，默认 `claude-haiku-4-5` |
| `opus_model` | 否 | Opus 模型，默认 `claude-opus-4-5` |
| `sonnet_model` | 否 | Sonnet 模型，默认 `claude-sonnet-4-5` |
| `env_vars` | 否 | 自定义环境变量（map[string]string） |

#### 环境变量配置

每个 provider 可以配置 `env_vars` 字段，用于设置任意环境变量。这些变量会：
1. 在使用 `opencc use` 时导出到系统环境
2. 在代理转发请求时作为 HTTP 头传递（格式：`x-env-变量名小写`）

配置示例：

```json
{
  "providers": {
    "anthropic-high-performance": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5-20250929",
      "env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000",
        "ANTHROPIC_MAX_CONTEXT_WINDOW": "1000000",
        "CLAUDE_CODE_EFFORT_LEVEL": "high",
        "MY_CUSTOM_VAR": "custom_value"
      }
    }
  }
}
```

**常用环境变量参考**：

| 环境变量 | 说明 | 示例值 |
|---------|------|--------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | 最大输出 token 数 | `64000` |
| `MAX_THINKING_TOKENS` | 扩展思考预算 | `50000` |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | 最大上下文窗口 | `1000000` |
| `CLAUDE_CODE_EFFORT_LEVEL` | 努力级别 (Opus 4.6) | `high` / `medium` / `low` |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash 默认超时（毫秒） | `180000` |
| `BASH_MAX_TIMEOUT_MS` | Bash 最大超时（毫秒） | `600000` |
| `BASH_MAX_OUTPUT_LENGTH` | Bash 输出最大字符数 | `50000` |
| `CLAUDE_CODE_SUBAGENT_MODEL` | 子代理模型 | `claude-haiku-4-5` |
| `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` | 自动压缩百分比 (1-100) | `85` |
| `MAX_MCP_OUTPUT_TOKENS` | MCP 工具响应最大 token | `30000` |

你可以添加任何自定义环境变量，不限于上述列表。

### 场景路由

opencc 支持基于请求特征的智能路由，可以为不同类型的请求配置专用的 provider 链。

#### 支持的场景

| 场景 | 触发条件 | 优先级 |
|------|---------|--------|
| `webSearch` | 请求包含 `web_search` 工具 | 1（最高） |
| `think` | 请求启用 thinking 模式 | 2 |
| `image` | 请求包含图片内容 | 3 |
| `longContext` | 请求内容超过阈值 | 4 |
| `background` | 请求使用 Haiku 模型 | 5 |
| `default` | 其他所有请求 | 6（最低） |

#### 配置示例

在 profile 中添加 `routing` 字段：

```json
{
  "profiles": {
    "smart-routing": {
      "providers": ["anthropic-main"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [
            {
              "name": "anthropic-thinking",
              "model": "claude-opus-4-5-20250514"
            }
          ]
        },
        "image": {
          "providers": [
            {"name": "anthropic-vision"}
          ]
        },
        "longContext": {
          "providers": [
            {"name": "anthropic-long-context"}
          ]
        },
        "background": {
          "providers": [
            {"name": "local-haiku"}
          ]
        },
        "webSearch": {
          "providers": [
            {"name": "search-provider"}
          ]
        }
      }
    }
  }
}
```

#### 配置说明

- **`long_context_threshold`**：longContext 场景的触发阈值（字符数），默认 32000
- **`routing.<scenario>.providers`**：该场景使用的 provider 列表
- **`routing.<scenario>.providers[].model`**：可选，覆盖该 provider 的默认模型

完整示例见 `example-scenario-routing-config.json`。

通过 `opencc config` 进入 TUI，按 `f` 进入 profile 编辑器，再按 `r` 可以配置场景路由。

### 从旧版迁移

如果之前使用 `~/.cc_envs/` 格式的配置文件，opencc 会在首次运行时自动迁移到 `~/.opencc/opencc.json`。旧目录不会被删除，可以手动清理。

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
git tag v1.3.1
git push origin v1.3.1
```

## 目录结构

```
├── main.go              # 入口
├── cmd/                 # CLI 命令 (cobra)
├── internal/
│   ├── config/          # 统一 JSON 配置管理（Store + 迁移）
│   ├── daemon/          # Web 守护进程管理（启停、平台适配）
│   ├── proxy/           # HTTP 代理服务器
│   └── web/             # Web 管理 API + 嵌入式静态前端
├── tui/                 # TUI 界面 (bubbletea)
├── install.sh           # 用户安装脚本
└── deploy.sh            # 构建发布脚本
```

## License

MIT
