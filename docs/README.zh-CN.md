# GoZen

<p align="center">
  <img src="https://raw.githubusercontent.com/dopejs/GoZen/main/assets/gozen.svg" alt="GoZen Logo" width="120">
</p>

[English](../README.md) | [繁體中文](README.zh-TW.md) | [Español](README.es.md)

> **Go Zen** — 进入禅意般的心流编程状态。**Goes Env** — 无缝环境切换。谐音"够禅"。

多 CLI 环境切换器，支持 Claude Code、Codex、OpenCode，带 API 代理自动故障转移。

## 功能

- **多 CLI 支持** — 支持 Claude Code、Codex、OpenCode 三种 CLI，可按项目配置
- **多配置管理** — 在 `~/.zen/zen.json` 中统一管理所有 API 配置
- **统一守护进程** — 单个 `zend` 进程同时托管代理服务和 Web 管理界面
- **代理故障转移** — 内置 HTTP 代理，当主 provider 不可用时自动切换到备用
- **场景路由** — 根据请求特征（thinking、image、longContext 等）智能路由
- **项目绑定** — 将目录绑定到特定 profile 和 CLI，实现项目级自动配置
- **环境变量配置** — 在 provider 级别为每个 CLI 单独配置环境变量
- **Web 管理界面** — 浏览器可视化管理，支持密码保护访问
- **Web 安全** — 自动生成访问密码、会话认证、RSA 加密令牌传输
- **配置同步** — 通过 WebDAV、S3、GitHub Gist 或 GitHub Repo 跨设备同步 provider、profile 和设置，使用 AES-256-GCM 加密
- **版本更新检查** — 启动时自动非阻塞检查新版本（24 小时缓存）
- **自更新** — `zen upgrade` 一键升级，支持 semver 版本匹配（支持预发布版本）
- **Shell 补全** — 支持 zsh / bash / fish

## 安装

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

卸载：

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh
```

## 快速开始

```sh
# 添加第一个 provider
zen config add provider

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
| `zen -y` / `zen --yes` | 自动批准 CLI 权限（claude `--permission-mode acceptEdits`, codex `-a never`） |
| `zen use <provider>` | 直接使用指定 provider（无代理） |
| `zen pick` | 交互选择 provider 启动 |
| `zen list` | 列出所有 provider 和 profile |
| `zen config` | 显示配置子命令 |
| `zen config add provider` | 添加新 provider |
| `zen config add profile` | 添加新 profile |
| `zen config default-client` | 设置默认 CLI 客户端 |
| `zen config default-profile` | 设置默认 profile |
| `zen config reset-password` | 重置 Web UI 访问密码 |
| `zen config sync` | 从远程同步后端拉取配置 |
| `zen daemon start` | 启动 zend 守护进程 |
| `zen daemon stop` | 停止守护进程 |
| `zen daemon restart` | 重启守护进程 |
| `zen daemon status` | 查看守护进程状态 |
| `zen daemon enable` | 安装为系统服务 |
| `zen daemon disable` | 卸载系统服务 |
| `zen bind <profile>` | 绑定当前目录到 profile |
| `zen bind --cli <cli>` | 绑定当前目录使用指定 CLI |
| `zen unbind` | 解除当前目录绑定 |
| `zen status` | 显示当前目录绑定状态 |
| `zen web` | 在浏览器中打开 Web 管理界面 |
| `zen upgrade` | 升级到最新版本 |
| `zen version` | 显示版本 |

## 守护进程架构

v2.1 中，GoZen 使用统一的守护进程（`zend`）同时托管 HTTP 代理和 Web 管理界面：

- **代理服务** 运行在端口 `19841`（可通过 `proxy_port` 配置）
- **Web UI** 运行在端口 `19840`（可通过 `web_port` 配置）
- 运行 `zen` 或 `zen web` 时守护进程自动启动
- 配置变更通过文件监控实现热重载
- 同步自动推送（防抖 2 秒）和自动拉取由守护进程管理

```sh
# 手动管理守护进程
zen daemon start          # 启动守护进程
zen daemon stop           # 停止守护进程
zen daemon restart        # 重启守护进程
zen daemon status         # 查看守护进程状态

# 系统服务（开机自启）
zen daemon enable         # 安装为系统服务
zen daemon disable        # 卸载系统服务
```

## 多 CLI 支持

zen 支持三种 AI 编程助手 CLI：

| CLI | 说明 | API 格式 |
|-----|------|---------|
| `claude` | Claude Code（默认） | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

### 设置默认 CLI

```sh
zen config default-client

# 通过 Web UI
zen web  # Settings 页面
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

## Web 管理界面

```sh
# 打开浏览器（如需要会自动启动守护进程）
zen web
```

Web UI 功能：
- Provider 和 Profile 管理
- 项目绑定管理
- 全局设置（默认客户端、默认 Profile、端口）
- 配置同步设置
- 请求日志查看（支持自动刷新）
- 模型字段自动补全

### Web UI 安全

守护进程首次启动时自动生成访问密码。非本地请求（127.0.0.1/::1 以外）需要登录。

- **会话认证** 支持可配置的过期时间
- **暴力破解保护** 指数级退避
- **RSA 加密** 敏感令牌传输（API 密钥在浏览器端加密后发送）
- 本地访问（127.0.0.1）免认证

```sh
# 重置 Web UI 密码
zen config reset-password

# 通过 Web UI 修改密码
zen web  # Settings → Change Password
```

## 配置同步

跨设备同步 provider、profile、默认 profile 和默认 client。认证令牌在上传前使用 AES-256-GCM（PBKDF2-SHA256 密钥派生）加密。

支持的后端：
- **WebDAV** — 任何 WebDAV 服务器（如 Nextcloud、ownCloud）
- **S3** — AWS S3 或 S3 兼容存储（如 MinIO、Cloudflare R2）
- **GitHub Gist** — 私有 gist（需要具有 `gist` 权限的 PAT）
- **GitHub Repo** — 通过 Contents API 存储到仓库文件（需要具有 `repo` 权限的 PAT）

### 通过 Web UI 设置

```sh
zen web  # Settings → Config Sync
```

### 通过 CLI 手动拉取

```sh
zen config sync
```

### 冲突解决

- 按实体时间戳合并：较新的修改胜出
- 删除的实体使用墓碑标记（30 天后过期）
- 标量值（默认 profile/client）：较新的时间戳胜出

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
| `~/.zen/zend.log` | 守护进程日志 |
| `~/.zen/zend.pid` | 守护进程 PID 文件 |
| `~/.zen/logs.db` | 请求日志数据库（SQLite） |

### 完整配置示例

```json
{
  "version": 7,
  "default_profile": "default",
  "default_client": "claude",
  "proxy_port": 19841,
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
      "client": "codex"
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

# 预发布版本
zen upgrade 2.1.0-alpha.1
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
git tag v2.1.0
git push origin v2.1.0
```

## License

MIT
