---
sidebar_position: 16
title: 代理基礎設施 (BETA)
---

# 代理基礎設施 (BETA)

:::warning BETA 功能
代理基礎設施目前處於測試階段。預設情況下已停用，需要明確配置才能啟用。
:::

內建支援自主代理工作流程，包括會話管理、檔案協調、即時監控和安全控制。

## 功能特性

- **代理執行時期** — 執行自主代理任務，具有完整的生命週期管理
- **觀測站** — 即時監控代理會話和活動
- **護欄** — 代理行為的安全控制和約束
- **協調器** — 基於檔案的多代理工作流程協調
- **任務佇列** — 管理具有優先順序和相依性的代理任務
- **會話管理** — 跨多個專案追蹤代理會話

## 架構

```
代理客戶端 (Claude Code, Codex 等)
    ↓
代理執行時期
    ↓
┌─────────────┬──────────────┬─────────────┐
│ 觀測站      │ 護欄         │ 協調器      │
│ (監控)      │ (安全)       │ (同步)      │
└─────────────┴──────────────┴─────────────┘
    ↓
任務佇列 → 提供商 API
```

## 配置

### 啟用代理基礎設施

```json
{
  "agent": {
    "enabled": true,
    "runtime": {
      "max_concurrent_tasks": 5,
      "task_timeout": "30m",
      "auto_cleanup": true
    },
    "observatory": {
      "enabled": true,
      "update_interval": "5s",
      "history_retention": "7d"
    },
    "guardrails": {
      "enabled": true,
      "max_file_operations": 100,
      "max_api_calls": 1000,
      "allowed_paths": ["/Users/john/projects"],
      "blocked_commands": ["rm -rf", "sudo"]
    },
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    }
  }
}
```

## 元件

### 1. 代理執行時期

管理代理任務執行生命週期。

**功能特性：**
- 任務排程和執行
- 並行任務管理
- 逾時處理
- 自動清理
- 錯誤復原

**配置：**
```json
{
  "runtime": {
    "max_concurrent_tasks": 5,
    "task_timeout": "30m",
    "auto_cleanup": true,
    "retry_failed_tasks": true,
    "max_retries": 3
  }
}
```

**API：**
```bash
# 啟動代理任務
POST /api/v1/agent/tasks
Content-Type: application/json

{
  "name": "code-review",
  "description": "Review pull request #123",
  "priority": 1,
  "config": {
    "model": "claude-opus-4",
    "max_tokens": 100000
  }
}

# 取得任務狀態
GET /api/v1/agent/tasks/{task_id}

# 取消任務
DELETE /api/v1/agent/tasks/{task_id}
```

### 2. 觀測站

即時監控代理活動。

**功能特性：**
- 會話追蹤
- 活動日誌
- 效能指標
- 狀態更新
- 歷史資料

**配置：**
```json
{
  "observatory": {
    "enabled": true,
    "update_interval": "5s",
    "history_retention": "7d",
    "metrics": {
      "track_tokens": true,
      "track_costs": true,
      "track_latency": true
    }
  }
}
```

**監控指標：**
- 活躍會話
- 進行中的任務
- Token 使用量
- API 呼叫
- 檔案操作
- 錯誤率
- 平均延遲

**API：**
```bash
# 取得所有活躍會話
GET /api/v1/agent/sessions

# 取得會話詳情
GET /api/v1/agent/sessions/{session_id}

# 取得會話指標
GET /api/v1/agent/sessions/{session_id}/metrics
```

### 3. 護欄

代理行為的安全控制和約束。

**功能特性：**
- 操作限制
- 路徑限制
- 命令阻止
- 資源配額
- 審批工作流程

**配置：**
```json
{
  "guardrails": {
    "enabled": true,
    "max_file_operations": 100,
    "max_api_calls": 1000,
    "max_tokens_per_session": 1000000,
    "allowed_paths": [
      "/Users/john/projects",
      "/tmp/agent-workspace"
    ],
    "blocked_paths": [
      "/etc",
      "/System",
      "~/.ssh"
    ],
    "blocked_commands": [
      "rm -rf /",
      "sudo",
      "chmod 777"
    ],
    "require_approval": {
      "file_delete": true,
      "system_commands": true,
      "network_requests": false
    }
  }
}
```

**執行機制：**
- 執行前驗證
- 即時監控
- 自動阻止
- 審批提示
- 稽核日誌

**API：**
```bash
# 取得護欄狀態
GET /api/v1/agent/guardrails

# 更新護欄規則
PUT /api/v1/agent/guardrails
Content-Type: application/json

{
  "max_file_operations": 200,
  "blocked_commands": ["rm -rf", "sudo", "dd"]
}
```

### 4. 協調器

基於檔案的多代理工作流程協調。

**功能特性：**
- 檔案鎖定
- 變更偵測
- 衝突解決
- 狀態同步
- 事件通知

**配置：**
```json
{
  "coordinator": {
    "enabled": true,
    "lock_timeout": "5m",
    "change_detection": true,
    "conflict_resolution": "last-write-wins",
    "notification_webhook": "https://hooks.slack.com/..."
  }
}
```

**使用場景：**
- 多個代理編輯相同檔案
- 防止並行修改
- 偵測外部檔案變更
- 協調代理工作流程

**API：**
```bash
# 取得檔案鎖
POST /api/v1/agent/locks
Content-Type: application/json

{
  "path": "/path/to/file.go",
  "session_id": "sess_123",
  "timeout": "5m"
}

# 釋放檔案鎖
DELETE /api/v1/agent/locks/{lock_id}

# 取得檔案變更事件
GET /api/v1/agent/changes?since=2026-03-05T10:00:00Z
```

### 5. 任務佇列

管理具有優先順序和相依性的代理任務。

**功能特性：**
- 優先順序排程
- 任務相依性
- 佇列管理
- 狀態追蹤
- 重試邏輯

**配置：**
```json
{
  "task_queue": {
    "enabled": true,
    "max_queue_size": 100,
    "priority_levels": 5,
    "enable_dependencies": true,
    "retry_policy": {
      "max_retries": 3,
      "backoff": "exponential"
    }
  }
}
```

**API：**
```bash
# 新增任務到佇列
POST /api/v1/agent/queue
Content-Type: application/json

{
  "name": "run-tests",
  "priority": 2,
  "depends_on": ["build-project"],
  "config": {}
}

# 取得佇列狀態
GET /api/v1/agent/queue

# 從佇列中移除任務
DELETE /api/v1/agent/queue/{task_id}
```

## Web UI

存取代理儀表板：`http://localhost:19840/agent`

### 會話標籤

- **活躍會話** — 目前執行的代理會話
- **會話詳情** — 任務進度、指標、日誌
- **會話控制** — 暫停、繼續、取消

### 任務標籤

- **任務佇列** — 待處理和進行中的任務
- **任務歷史** — 已完成和失敗的任務
- **任務詳情** — 配置、日誌、結果

### 護欄標籤

- **操作限制** — 目前使用量 vs. 限制
- **被阻止的操作** — 最近被阻止的嘗試
- **審批佇列** — 等待審批的操作

### 指標標籤

- **Token 使用量** — 每個會話和總計
- **API 呼叫** — 請求計數和速率
- **檔案操作** — 讀取/寫入/刪除計數
- **效能** — 延遲和吞吐量

## 與 Claude Code 整合

GoZen 自動偵測 Claude Code 會話並提供代理基礎設施：

```bash
# 啟動帶有代理支援的 Claude Code
zen --agent

# 代理功能自動啟用：
# - 會話追蹤
# - 檔案協調
# - 護欄執行
# - 即時監控
```

**優勢：**
- 防止並行檔案修改
- 追蹤 token 使用量和成本
- 執行安全約束
- 監控代理活動
- 協調多代理工作流程

## 使用場景

### 多代理開發

多個代理在同一程式碼庫上工作：

```json
{
  "agent": {
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    },
    "guardrails": {
      "max_file_operations": 200,
      "allowed_paths": ["/Users/john/project"]
    }
  }
}
```

### 長時間執行的任務

監控和控制長時間執行的代理任務：

```json
{
  "agent": {
    "runtime": {
      "task_timeout": "2h",
      "auto_cleanup": false
    },
    "observatory": {
      "update_interval": "10s",
      "history_retention": "30d"
    }
  }
}
```

### 安全關鍵操作

執行嚴格的安全控制：

```json
{
  "agent": {
    "guardrails": {
      "enabled": true,
      "max_file_operations": 50,
      "blocked_commands": ["rm", "sudo", "chmod"],
      "require_approval": {
        "file_delete": true,
        "system_commands": true,
        "network_requests": true
      }
    }
  }
}
```

## 最佳實務

1. **啟用護欄** — 在生產環境中始終使用護欄
2. **設定適當的限制** — 根據使用場景配置限制
3. **主動監控** — 定期檢查觀測站儀表板
4. **使用檔案鎖定** — 為多代理工作流程啟用協調器
5. **配置審批** — 對破壞性操作要求審批
6. **審查日誌** — 定期稽核代理活動

## 限制

1. **效能開銷** — 監控和協調會增加延遲
2. **檔案鎖定** — 在多代理場景中可能導致延遲
3. **記憶體使用** — 會話歷史消耗記憶體
4. **複雜性** — 需要理解代理工作流程
5. **Beta 狀態** — 功能可能在未來版本中變更

## 疑難排解

### 代理會話未被追蹤

1. 驗證 `agent.enabled` 為 `true`
2. 檢查觀測站已啟用
3. 確保代理客戶端受支援（Claude Code、Codex）
4. 查看守護程式日誌中的錯誤

### 檔案鎖定問題

1. 檢查協調器已啟用
2. 驗證鎖定逾時是否合適
3. 查看活躍鎖：`GET /api/v1/agent/locks`
4. 如需要，手動釋放卡住的鎖

### 護欄未執行

1. 驗證護欄已啟用
2. 檢查規則配置是否正確
3. 查看被阻止的操作日誌
4. 確保代理客戶端遵守護欄

### 高記憶體使用

1. 減少歷史保留期
2. 降低更新間隔
3. 限制最大並行任務數
4. 啟用自動清理

## 安全考量

1. **路徑限制** — 始終配置允許/阻止的路徑
2. **命令阻止** — 阻止危險命令
3. **審批工作流程** — 對敏感操作要求審批
4. **稽核日誌** — 啟用全面的日誌記錄
5. **資源限制** — 設定適當的操作限制

## 未來增強

- 多代理協作協定
- 進階衝突解決策略
- 用於異常偵測的機器學習
- 與外部監控工具整合
- 代理行為分析
- 自動安全政策產生
