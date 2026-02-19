# GoZen

<p align="center">
  <img src="https://raw.githubusercontent.com/dopejs/GoZen/main/assets/gozen.svg" alt="GoZen Logo" width="120">
</p>

[English](../README.md) | [简体中文](README.zh-CN.md) | [Español](README.es.md)

> **Go Zen** — 進入禪意般的心流程式設計狀態。**Goes Env** — 無縫環境切換。諧音「夠禪」。

多 CLI 環境切換器，支援 Claude Code、Codex、OpenCode，具備 API 代理自動故障轉移。

## 功能

- **多 CLI 支援** — 支援 Claude Code、Codex、OpenCode 三種 CLI，可依專案設定
- **多組態管理** — 在 `~/.zen/zen.json` 中統一管理所有 API 組態
- **統一守護程序** — 單一 `zend` 程序同時託管代理伺服器與 Web 管理介面
- **代理故障轉移** — 內建 HTTP 代理，當主要 provider 無法使用時自動切換至備用
- **場景路由** — 根據請求特徵（thinking、image、longContext 等）智慧路由
- **專案綁定** — 將目錄綁定至特定 profile 與 CLI，實現專案層級自動組態
- **環境變數設定** — 在 provider 層級為每個 CLI 分別設定環境變數
- **Web 管理介面** — 瀏覽器視覺化管理，支援密碼保護存取
- **Web 安全** — 自動產生存取密碼、工作階段認證、RSA 加密令牌傳輸
- **組態同步** — 透過 WebDAV、S3、GitHub Gist 或 GitHub Repo 跨裝置同步 provider、profile 與設定，使用 AES-256-GCM 加密
- **版本更新檢查** — 啟動時自動非阻塞檢查新版本（24 小時快取）
- **自動更新** — `zen upgrade` 一鍵升級，支援 semver 版本比對（支援預發行版本）
- **Shell 補全** — 支援 zsh / bash / fish

### v3.0 新功能

- **使用量追蹤** — 依 provider、模型、專案追蹤 token 使用量與成本
- **預算控制** — 設定每日/每週/每月支出限制，支援警告/降級/阻止動作
- **Provider 健康監控** — 即時健康檢查，追蹤延遲與錯誤率
- **智慧負載均衡** — 多種策略：故障轉移、輪詢、最低延遲、最低成本
- **Webhook 通知** — 預算警報、provider 狀態變化、每日摘要通知（Slack/Discord/通用）
- **上下文壓縮** — token 數量超過閾值時自動壓縮上下文
- **中介軟體管道** — 可插拔中介軟體，用於請求/回應轉換
- **Agent 基礎設施** — 內建 agent 工作流程支援，具備工作階段管理

## 安裝

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

解除安裝：

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh
```

## 快速開始

```sh
# 新增第一個 provider
zen config add provider

# 啟動（使用預設 profile）
zen

# 使用指定 profile
zen -p work

# 使用指定 CLI
zen --cli codex
```

## 命令一覽

| 命令 | 說明 |
|------|------|
| `zen` | 啟動 CLI（使用專案綁定或預設組態） |
| `zen -p <profile>` | 使用指定 profile 啟動 |
| `zen -p` | 互動選擇 profile |
| `zen --cli <cli>` | 使用指定 CLI（claude/codex/opencode） |
| `zen -y` / `zen --yes` | 自動批准 CLI 權限（claude `--permission-mode acceptEdits`, codex `-a never`） |
| `zen use <provider>` | 直接使用指定 provider（不經代理） |
| `zen pick` | 互動選擇 provider 啟動 |
| `zen list` | 列出所有 provider 與 profile |
| `zen config` | 顯示設定子命令 |
| `zen config add provider` | 新增 provider |
| `zen config add profile` | 新增 profile |
| `zen config default-client` | 設定預設 CLI 用戶端 |
| `zen config default-profile` | 設定預設 profile |
| `zen config reset-password` | 重設 Web UI 存取密碼 |
| `zen config sync` | 從遠端同步後端拉取組態 |
| `zen daemon start` | 啟動 zend 守護程序 |
| `zen daemon stop` | 停止守護程序 |
| `zen daemon restart` | 重新啟動守護程序 |
| `zen daemon status` | 檢視守護程序狀態 |
| `zen daemon enable` | 安裝為系統服務 |
| `zen daemon disable` | 解除安裝系統服務 |
| `zen bind <profile>` | 將目前目錄綁定至 profile |
| `zen bind --cli <cli>` | 將目前目錄綁定使用指定 CLI |
| `zen unbind` | 解除目前目錄綁定 |
| `zen status` | 顯示目前目錄綁定狀態 |
| `zen web` | 在瀏覽器中開啟 Web 管理介面 |
| `zen upgrade` | 升級至最新版本 |
| `zen version` | 顯示版本 |

## 守護程序架構

v3.0 中，GoZen 使用統一的守護程序（`zend`）同時託管 HTTP 代理與 Web 管理介面：

- **代理伺服器** 執行於連接埠 `19841`（可透過 `proxy_port` 設定）
- **Web UI** 執行於連接埠 `19840`（可透過 `web_port` 設定）
- 執行 `zen` 或 `zen web` 時守護程序自動啟動
- 組態變更透過檔案監控實現熱重載
- 同步自動推送（防抖 2 秒）與自動拉取由守護程序管理

```sh
# 手動管理守護程序
zen daemon start          # 啟動守護程序
zen daemon stop           # 停止守護程序
zen daemon restart        # 重新啟動守護程序
zen daemon status         # 檢視守護程序狀態

# 系統服務（開機自動啟動）
zen daemon enable         # 安裝為系統服務
zen daemon disable        # 解除安裝系統服務
```

## 多 CLI 支援

zen 支援三種 AI 程式設計助手 CLI：

| CLI | 說明 | API 格式 |
|-----|------|---------|
| `claude` | Claude Code（預設） | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

### 設定預設 CLI

```sh
zen config default-client

# 透過 Web UI
zen web  # Settings 頁面
```

### 依專案設定 CLI

```sh
cd ~/work/project
zen bind --cli codex  # 此目錄使用 Codex
```

### 臨時使用其他 CLI

```sh
zen --cli opencode  # 本次使用 OpenCode
```

## Profile 管理

Profile 是一組 provider 的有序清單，用於故障轉移。

### 組態範例

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
# 使用預設 profile
zen

# 使用指定 profile
zen -p work

# 互動選擇
zen -p
```

## 專案綁定

將目錄綁定至特定 profile 和／或 CLI，實現專案層級自動組態。

```sh
cd ~/work/company-project

# 綁定 profile
zen bind work-profile

# 綁定 CLI
zen bind --cli codex

# 同時綁定
zen bind work-profile --cli codex

# 檢視狀態
zen status

# 解除綁定
zen unbind
```

**優先順序**：命令列參數 > 專案綁定 > 全域預設

## Web 管理介面

```sh
# 開啟瀏覽器（如需要會自動啟動守護程序）
zen web
```

Web UI 功能：
- Provider 與 Profile 管理
- 專案綁定管理
- 全域設定（預設用戶端、預設 Profile、連接埠）
- 組態同步設定
- 請求日誌檢視（支援自動重新整理）
- 模型欄位自動補全

### Web UI 安全

守護程序首次啟動時自動產生存取密碼。非本地請求（127.0.0.1/::1 以外）需要登入。

- **工作階段認證** 支援可設定的到期時間
- **暴力破解保護** 指數級退避
- **RSA 加密** 敏感令牌傳輸（API 金鑰在瀏覽器端加密後傳送）
- 本地存取（127.0.0.1）免認證

```sh
# 重設 Web UI 密碼
zen config reset-password

# 透過 Web UI 變更密碼
zen web  # Settings → Change Password
```

## 組態同步

跨裝置同步 provider、profile、預設 profile 與預設 client。認證令牌在上傳前使用 AES-256-GCM（PBKDF2-SHA256 金鑰衍生）加密。

支援的後端：
- **WebDAV** — 任何 WebDAV 伺服器（如 Nextcloud、ownCloud）
- **S3** — AWS S3 或 S3 相容儲存（如 MinIO、Cloudflare R2）
- **GitHub Gist** — 私有 gist（需要具有 `gist` 權限的 PAT）
- **GitHub Repo** — 透過 Contents API 儲存至倉庫檔案（需要具有 `repo` 權限的 PAT）

### 透過 Web UI 設定

```sh
zen web  # Settings → Config Sync
```

### 透過 CLI 手動拉取

```sh
zen config sync
```

### 衝突解決

- 按實體時間戳合併：較新的修改勝出
- 刪除的實體使用墓碑標記（30 天後過期）
- 純量值（預設 profile/client）：較新的時間戳勝出

## 環境變數設定

每個 provider 可以為不同 CLI 設定獨立的環境變數：

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

### Claude Code 常用環境變數

| 變數 | 說明 |
|------|------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | 最大輸出 token |
| `MAX_THINKING_TOKENS` | 擴展思考預算 |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | 最大上下文窗口 |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash 預設逾時 |

## 場景路由

根據請求特徵自動路由至不同 provider：

| 場景 | 觸發條件 |
|------|---------|
| `think` | 啟用 thinking 模式 |
| `image` | 包含圖片內容 |
| `longContext` | 內容超過閾值 |
| `webSearch` | 使用 web_search 工具 |
| `background` | 使用 Haiku 模型 |

**Fallback 機制**：若場景設定的 providers 全部失敗，會自動 fallback 至 profile 的預設 providers。

組態範例：

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

## 使用量追蹤與預算控制

追蹤 API 使用量並設定支出限制：

```json
{
  "pricing": {
    "claude-sonnet-4-20250514": {"input_per_million": 3.0, "output_per_million": 15.0},
    "claude-opus-4-20250514": {"input_per_million": 15.0, "output_per_million": 75.0}
  },
  "budgets": {
    "daily": {"amount": 10.0, "action": "warn"},
    "monthly": {"amount": 100.0, "action": "block"},
    "per_project": true
  }
}
```

預算動作：`warn`（記錄警告）、`downgrade`（切換至更便宜的模型）、`block`（拒絕請求）。

## Provider 健康監控

自動健康檢查與指標追蹤：

```json
{
  "health_check": {
    "enabled": true,
    "interval_secs": 60,
    "timeout_secs": 10
  }
}
```

透過 Web UI 或 API 檢視 provider 健康狀態：`GET /api/v1/health/providers`

## 智慧負載均衡

為每個 profile 設定負載均衡策略：

```json
{
  "profiles": {
    "balanced": {
      "providers": ["provider-a", "provider-b", "provider-c"],
      "strategy": "least-latency"
    }
  }
}
```

策略：
- `failover` — 依序嘗試 provider（預設）
- `round-robin` — 均勻分配請求
- `least-latency` — 路由至最快的 provider
- `least-cost` — 路由至該模型最便宜的 provider

## Webhook 通知

取得重要事件通知：

```json
{
  "webhooks": [
    {
      "name": "slack-alerts",
      "url": "https://hooks.slack.com/services/xxx",
      "events": ["budget_warning", "budget_exceeded", "provider_down", "provider_up"],
      "enabled": true
    }
  ]
}
```

事件：`budget_warning`、`budget_exceeded`、`provider_down`、`provider_up`、`failover`、`daily_summary`

## 上下文壓縮

當上下文超過閾值時自動壓縮：

```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 100000,
    "target_ratio": 0.5
  }
}
```

## 中介軟體管道

使用可插拔中介軟體轉換請求與回應：

```json
{
  "middleware": {
    "enabled": true,
    "middlewares": [
      {"name": "context-injection", "enabled": true, "config": {"inject_cursorrules": true}},
      {"name": "rate-limiter", "enabled": true, "config": {"requests_per_minute": 60}}
    ]
  }
}
```

內建中介軟體：`context-injection`、`request-logger`、`rate-limiter`、`compression`

## 組態檔案

| 檔案 | 說明 |
|------|------|
| `~/.zen/zen.json` | 主組態檔案 |
| `~/.zen/zend.log` | 守護程序日誌 |
| `~/.zen/zend.pid` | 守護程序 PID 檔案 |
| `~/.zen/logs.db` | 請求日誌資料庫（SQLite） |

### 完整組態範例

```json
{
  "version": 8,
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

## 升級

```sh
# 最新版本
zen upgrade

# 指定版本
zen upgrade 3.0
zen upgrade 3.0.0

# 預發行版本
zen upgrade 3.0.0-alpha.1
```

## 從舊版遷移

GoZen 會自動從舊版本遷移組態：
- `~/.opencc/opencc.json` → `~/.zen/zen.json`（從 OpenCC v1.x 遷移）
- `~/.cc_envs/` → `~/.zen/zen.json`（從舊格式遷移）

## 開發

```sh
# 建置
go build -o zen .

# 測試
go test ./...
```

發佈：推送 tag 後 GitHub Actions 自動建置。

```sh
git tag v3.0.0
git push origin v3.0.0
```

## License

MIT
