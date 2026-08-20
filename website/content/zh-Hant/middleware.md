---
sidebar_position: 15
title: 中間件管道 (BETA)
---

# 中間件管道 (BETA)

:::warning BETA 功能
中間件管道目前處於測試階段。預設情況下已停用，需要明確配置才能啟用。
:::

使用可插拔中間件擴展 GoZen，實現請求/回應轉換、日誌記錄、速率限制和自訂處理。

## 功能特性

- **可插拔架構** — 無需修改核心程式碼即可新增自訂處理邏輯
- **基於優先順序的執行** — 控制中間件執行順序
- **請求/回應鉤子** — 在傳送前處理請求，在接收後處理回應
- **內建中間件** — 上下文注入、日誌記錄、速率限制、壓縮
- **插件載入器** — 從本地檔案或遠端 URL 載入中間件
- **錯誤處理** — 優雅的錯誤處理和回退行為

## 架構

```
客戶端請求
    ↓
[中間件 1: 優先順序 100]
    ↓
[中間件 2: 優先順序 200]
    ↓
[中間件 3: 優先順序 300]
    ↓
提供商 API
    ↓
[中間件 3: 回應]
    ↓
[中間件 2: 回應]
    ↓
[中間件 1: 回應]
    ↓
客戶端回應
```

## 配置

### 啟用中間件管道

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "context-injection",
        "enabled": true,
        "priority": 100,
        "config": {}
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info"
        }
      }
    ]
  }
}
```

**選項：**

| 選項 | 描述 |
|------|------|
| `enabled` | 啟用中間件管道 |
| `pipeline` | 中間件配置陣列 |
| `name` | 中間件識別符 |
| `priority` | 執行順序（越小越早） |
| `config` | 中間件特定配置 |

## 內建中間件

### 1. 上下文注入

向請求中注入自訂上下文。

```json
{
  "name": "context-injection",
  "enabled": true,
  "priority": 100,
  "config": {
    "system_prompt": "你是一個有用的編碼助手。",
    "metadata": {
      "session_id": "sess_123",
      "user_id": "user_456"
    }
  }
}
```

**使用場景：**
- 新增系統提示
- 注入會話中繼資料
- 新增使用者上下文

### 2. 請求日誌記錄器

記錄所有請求和回應。

```json
{
  "name": "request-logger",
  "enabled": true,
  "priority": 200,
  "config": {
    "log_level": "info",
    "log_body": false,
    "log_headers": true
  }
}
```

**使用場景：**
- 除錯
- 稽核追蹤
- 效能監控

### 3. 速率限制器

限制每個提供商或全域的請求速率。

```json
{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60,
    "burst": 10,
    "per_provider": true
  }
}
```

**使用場景：**
- 防止速率限制錯誤
- 控制 API 使用
- 防止濫用

### 4. 壓縮 (BETA)

當 token 數量超過閾值時壓縮上下文。

```json
{
  "name": "compression",
  "enabled": true,
  "priority": 400,
  "config": {
    "threshold_tokens": 50000,
    "target_tokens": 20000
  }
}
```

詳見[上下文壓縮](./compression.md)。

### 5. 會話記憶 (BETA)

跨會話維護對話記憶。

```json
{
  "name": "session-memory",
  "enabled": true,
  "priority": 150,
  "config": {
    "max_memories": 100,
    "ttl_hours": 24,
    "storage": "sqlite"
  }
}
```

**使用場景：**
- 記住使用者偏好
- 追蹤對話歷史
- 跨會話維護上下文

### 6. 編排 (BETA)

將請求路由到多個提供商並聚合回應。

```json
{
  "name": "orchestration",
  "enabled": true,
  "priority": 500,
  "config": {
    "strategy": "parallel",
    "providers": ["anthropic", "openai"],
    "consensus": "longest"
  }
}
```

**使用場景：**
- 比較模型輸出
- 關鍵請求的冗餘
- 透過共識提高品質

## 自訂中間件

### 中間件介面

```go
type Middleware interface {
    Name() string
    Priority() int
    ProcessRequest(ctx *RequestContext) error
    ProcessResponse(ctx *ResponseContext) error
}

type RequestContext struct {
    Provider  string
    Model     string
    Messages  []Message
    Metadata  map[string]interface{}
}

type ResponseContext struct {
    Provider  string
    Model     string
    Response  *APIResponse
    Latency   time.Duration
    Metadata  map[string]interface{}
}
```

### 範例：自訂標頭注入

```go
package main

import (
    "github.com/dopejs/gozen/internal/middleware"
)

type CustomHeaderMiddleware struct {
    headers map[string]string
}

func (m *CustomHeaderMiddleware) Name() string {
    return "custom-headers"
}

func (m *CustomHeaderMiddleware) Priority() int {
    return 250
}

func (m *CustomHeaderMiddleware) ProcessRequest(ctx *middleware.RequestContext) error {
    for k, v := range m.headers {
        ctx.Metadata[k] = v
    }
    return nil
}

func (m *CustomHeaderMiddleware) ProcessResponse(ctx *middleware.ResponseContext) error {
    // 不需要回應處理
    return nil
}

func init() {
    middleware.Register("custom-headers", func(config map[string]interface{}) middleware.Middleware {
        return &CustomHeaderMiddleware{
            headers: config["headers"].(map[string]string),
        }
    })
}
```

### 載入自訂中間件

#### 本地插件

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "local",
        "path": "/path/to/custom-middleware.so",
        "config": {
          "headers": {
            "X-Custom-Header": "value"
          }
        }
      }
    ]
  }
}
```

#### 遠端插件

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "remote",
        "url": "https://example.com/middleware/custom-headers.so",
        "checksum": "sha256:abc123...",
        "config": {}
      }
    ]
  }
}
```

## Web UI

在 `http://localhost:19840/settings` 存取中間件設定：

1. 導覽到 "Middleware" 標籤（標有 BETA 徽章）
2. 切換 "Enable Middleware Pipeline"
3. 從管道中新增/刪除中間件
4. 調整優先順序和配置
5. 啟用/停用單個中間件
6. 點選 "Save"

## API 端點

### 列出中間件

```bash
GET /api/v1/middleware
```

回應：
```json
{
  "enabled": true,
  "pipeline": [
    {
      "name": "context-injection",
      "enabled": true,
      "priority": 100,
      "type": "builtin"
    },
    {
      "name": "request-logger",
      "enabled": true,
      "priority": 200,
      "type": "builtin"
    }
  ]
}
```

### 新增中間件

```bash
POST /api/v1/middleware
Content-Type: application/json

{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60
  }
}
```

### 更新中間件

```bash
PUT /api/v1/middleware/{name}
Content-Type: application/json

{
  "enabled": false
}
```

### 刪除中間件

```bash
DELETE /api/v1/middleware/{name}
```

### 重新載入管道

```bash
POST /api/v1/middleware/reload
```

## 使用場景

### 開發環境

新增除錯日誌和請求檢查：

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 100,
        "config": {
          "log_level": "debug",
          "log_body": true
        }
      }
    ]
  }
}
```

### 生產環境

新增速率限制和監控：

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "rate-limiter",
        "enabled": true,
        "priority": 100,
        "config": {
          "requests_per_minute": 100,
          "burst": 20
        }
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info",
          "log_body": false
        }
      }
    ]
  }
}
```

### 多提供商比較

使用編排來比較輸出：

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "orchestration",
        "enabled": true,
        "priority": 500,
        "config": {
          "strategy": "parallel",
          "providers": ["anthropic", "openai", "google"],
          "consensus": "longest"
        }
      }
    ]
  }
}
```

## 最佳實務

1. **使用適當的優先順序** — 較小的數字先執行
2. **保持中間件專注** — 每個中間件應該做好一件事
3. **優雅地處理錯誤** — 不要因錯誤而破壞管道
4. **徹底測試** — 在生產前驗證中間件行為
5. **監控效能** — 追蹤中間件開銷
6. **記錄配置** — 清楚地記錄配置選項

## 限制

1. **效能開銷** — 每個中間件都會增加延遲
2. **複雜性** — 太多中間件會使除錯變得困難
3. **插件安全** — 遠端插件需要信任和驗證
4. **錯誤傳播** — 中間件錯誤會影響所有請求
5. **配置複雜性** — 複雜的管道更難維護

## 疑難排解

### 中間件未執行

1. 驗證 `middleware.enabled` 為 `true`
2. 檢查中間件在管道中已啟用
3. 驗證優先順序設定正確
4. 查看守護程式日誌中的中間件錯誤

### 意外行為

1. 檢查中間件執行順序（優先順序）
2. 驗證配置是否正確
3. 單獨測試中間件
4. 查看中間件日誌

### 效能問題

1. 識別慢速中間件（檢查日誌）
2. 減少中間件數量
3. 最佳化中間件實作
4. 考慮停用非必要的中間件

### 插件載入失敗

1. 驗證插件路徑是否正確
2. 檢查插件是否為正確的架構編譯
3. 驗證校驗和匹配（對於遠端插件）
4. 查看插件日誌中的錯誤

## 安全考量

1. **驗證插件** — 僅載入受信任的插件
2. **驗證校驗和** — 始終驗證遠端插件校驗和
3. **沙箱插件** — 考慮在隔離環境中執行插件
4. **稽核中間件** — 在部署前審查中間件程式碼
5. **監控行為** — 注意意外的中間件行為

## 未來增強

- WebAssembly 插件支援以實現跨平台相容性
- 用於共享社群插件的中間件市場
- Web UI 中的視覺化管道編輯器
- 中間件效能分析
- 插件更新的熱重載
- 中間件測試框架
