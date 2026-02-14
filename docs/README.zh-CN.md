# GoZen

[English](../README.md) | [繁體中文](README.zh-TW.md) | [Español](README.es.md)

> **Go Zen** — 进入禅意般的心流编程状态。**Goes Env** — 无缝环境切换。谐音"够禅"。

多 CLI 环境切换器，支持 Claude Code、Codex、OpenCode，带 API 代理自动故障转移。

## 功能

- **多 CLI 支持** — 支持 Claude Code、Codex、OpenCode 三种 CLI，可按项目配置
- **多配置管理** — 在 `~/.zen/zen.json` 中统一管理所有 API 配置
- **代理故障转移** — 内置 HTTP 代理，当主 provider 不可用时自动切换到备用
- **场景路由** — 根据请求特征（thinking、image、longContext 等）智能路由
- **项目绑定** — 将目录绑定到特定 profile 和 CLI，实现项目级自动配置
- **环境变量配置** — 在 provider 级别为每个 CLI 单独配置环境变量
- **TUI 配置界面** — 交互式终端界面，支持 Dashboard 和传统两种模式
- **Web 管理界面** — 浏览器可视化管理 provider、profile 和项目绑定
- **版本更新检查** — 启动时自动非阻塞检查新版本（24 小时缓存）
- **自更新** — `zen upgrade` 一键升级，支持 semver 版本匹配
- **Shell 补全** — 支持 zsh / bash / fish

## 安装

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

卸载：

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh -s -- --uninstall
```

## 快速开始

```sh
# 打开 TUI 配置界面，创建第一个 provider
zen config

# 启动（使用默认 profile）
zen

# 使用指定 profile
zen -p work

# 使用指定 CLI
zen --cli codex
```

## 命令一览

| 命令 | 说明 |
|------|------|
| `zen` | 启动 CLI（使用项目绑定或默认配置） |
| `zen -p <profile>` | 使用指定 profile 启动 |
| `zen -p` | 交互选择 profile |
| `zen --cli <cli>` | 使用指定 CLI（claude/codex/opencode） |
| `zen use <provider>` | 直接使用指定 provider（无代理） |
| `zen pick` | 交互选择 provider 启动 |
| `zen list` | 列出所有 provider 和 profile |
| `zen config` | 打开 TUI 配置界面 |
| `zen config --legacy` | 使用传统 TUI 界面 |
| `zen bind <profile>` | 绑定当前目录到 profile |
| `zen bind --cli <cli>` | 绑定当前目录使用指定 CLI |
| `zen unbind` | 解除当前目录绑定 |
| `zen status` | 显示当前目录绑定状态 |
| `zen web start` | 启动 Web 管理界面 |
| `zen web open` | 在浏览器中打开 Web 界面 |
| `zen web stop` | 停止 Web 服务 |
| `zen web restart` | 重启 Web 服务 |
| `zen upgrade` | 升级到最新版本 |
| `zen version` | 显示版本 |

## 多 CLI 支持

zen 支持三种 AI 编程助手 CLI：

| CLI | 说明 | API 格式 |
|-----|------|---------|
| `claude` | Claude Code（默认） | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

### 设置默认 CLI

```sh
# 通过 TUI
zen config  # Settings → Default CLI

# 通过 Web UI
zen web open  # Settings 页面
```

### 按项目配置 CLI

```sh
cd ~/work/project
zen bind --cli codex  # 该目录使用 Codex
```

### 临时使用其他 CLI

```sh
zen --cli opencode  # 本次使用 OpenCode
```

## Profile 管理

Profile 是一组 provider 的有序列表，用于故障转移。

### 配置示例

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-main", "anthropic-backup"]
    },
    "work": {
      "providers": ["company-api"],
      "routing": {
        "think": {"providers": [{"name": "thinking-api"}]}
      }
    }
  }
}
```

### 使用 Profile

```sh
# 使用默认 profile
zen

# 使用指定 profile
zen -p work

# 交互选择
zen -p
```

## 项目绑定

将目录绑定到特定 profile 和/或 CLI，实现项目级自动配置。

```sh
cd ~/work/company-project

# 绑定 profile
zen bind work-profile

# 绑定 CLI
zen bind --cli codex

# 同时绑定
zen bind work-profile --cli codex

# 查看状态
zen status

# 解除绑定
zen unbind
```

**优先级**：命令行参数 > 项目绑定 > 全局默认

## TUI 配置界面

```sh
zen config
```

v1.5 提供全新 Dashboard 界面：

- **左侧列表**：Providers、Profiles、Project Bindings
- **右侧详情**：选中项的详细信息
- **快捷键**：
  - `a` - 添加新项
  - `e` - 编辑选中项
  - `d` - 删除选中项
  - `Tab` - 切换焦点
  - `q` - 返回/退出

使用 `--legacy` 切换到传统界面。

## Web 管理界面

```sh
# 启动（后台运行，端口 19840）
zen web start

# 打开浏览器
zen web open

# 停止
zen web stop

# 重启
zen web restart
```

Web UI 功能：
- Provider 和 Profile 管理
- 项目绑定管理
- 全局设置（默认 CLI、默认 Profile、端口）
- 请求日志查看
- 模型字段自动补全

## 环境变量配置

每个 provider 可以为不同 CLI 配置独立的环境变量：

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000"
      },
      "codex_env_vars": {
        "CODEX_SOME_VAR": "value"
      },
      "opencode_env_vars": {
        "OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": "64000"
      }
    }
  }
}
```

### Claude Code 常用环境变量

| 变量 | 说明 |
|------|------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | 最大输出 token |
| `MAX_THINKING_TOKENS` | 扩展思考预算 |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | 最大上下文窗口 |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash 默认超时 |

## 场景路由

根据请求特征自动路由到不同 provider：

| 场景 | 触发条件 |
|------|---------|
| `think` | 启用 thinking 模式 |
| `image` | 包含图片内容 |
| `longContext` | 内容超过阈值 |
| `webSearch` | 使用 web_search 工具 |
| `background` | 使用 Haiku 模型 |

**Fallback 机制**：如果场景配置的 providers 全部失败，会自动 fallback 到 profile 的默认 providers。

配置示例：

```json
{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}
```

## 配置文件

| 文件 | 说明 |
|------|------|
| `~/.zen/zen.json` | 主配置文件 |
| `~/.zen/proxy.log` | 代理日志 |
| `~/.zen/web.log` | Web 服务日志 |

### 完整配置示例

```json
{
  "version": 6,
  "default_profile": "default",
  "default_cli": "claude",
  "web_port": 19840,
  "providers": {
    "anthropic": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000"
      }
    }
  },
  "profiles": {
    "default": {
      "providers": ["anthropic"]
    }
  },
  "project_bindings": {
    "/path/to/project": {
      "profile": "work",
      "cli": "codex"
    }
  }
}
```

## 升级

```sh
# 最新版本
zen upgrade

# 指定版本
zen upgrade 2.1
zen upgrade 2.1.0
```

## 从旧版迁移

GoZen 会自动从旧版本迁移配置：
- `~/.opencc/opencc.json` → `~/.zen/zen.json`（从 OpenCC v1.x 迁移）
- `~/.cc_envs/` → `~/.zen/zen.json`（从旧格式迁移）

## 开发

```sh
# 构建
go build -o zen .

# 测试
go test ./...
```

发布：打 tag 后 GitHub Actions 自动构建。

```sh
git tag v2.0.0
git push origin v2.0.0
```

## License

MIT
