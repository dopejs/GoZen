---
sidebar_position: 15
title: ミドルウェアパイプライン (BETA)
---

# ミドルウェアパイプライン (BETA)

:::warning BETA機能
ミドルウェアパイプラインは現在ベータ版です。デフォルトでは無効になっており、明示的な設定が必要です。
:::

プラグ可能なミドルウェアでGoZenを拡張し、リクエスト/レスポンス変換、ログ記録、レート制限、カスタム処理を実現します。

## 機能

- **プラグ可能なアーキテクチャ** — コアコードを変更せずにカスタム処理ロジックを追加
- **優先度ベースの実行** — ミドルウェアの実行順序を制御
- **リクエスト/レスポンスフック** — 送信前にリクエストを処理、受信後にレスポンスを処理
- **組み込みミドルウェア** — コンテキスト注入、ログ記録、レート制限、圧縮
- **プラグインローダー** — ローカルファイルまたはリモートURLからミドルウェアをロード
- **エラーハンドリング** — 優雅なエラー処理とフォールバック動作

## アーキテクチャ

```
クライアントリクエスト
    ↓
[ミドルウェア1: 優先度100]
    ↓
[ミドルウェア2: 優先度200]
    ↓
[ミドルウェア3: 優先度300]
    ↓
プロバイダーAPI
    ↓
[ミドルウェア3: レスポンス]
    ↓
[ミドルウェア2: レスポンス]
    ↓
[ミドルウェア1: レスポンス]
    ↓
クライアントレスポンス
```

## 設定

### ミドルウェアパイプラインの有効化

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

**オプション：**

| オプション | 説明 |
|------|------|
| `enabled` | ミドルウェアパイプラインを有効化 |
| `pipeline` | ミドルウェア設定の配列 |
| `name` | ミドルウェア識別子 |
| `priority` | 実行順序（小さいほど早い） |
| `config` | ミドルウェア固有の設定 |

## 組み込みミドルウェア

### 1. コンテキスト注入

リクエストにカスタムコンテキストを注入します。

```json
{
  "name": "context-injection",
  "enabled": true,
  "priority": 100,
  "config": {
    "system_prompt": "あなたは有用なコーディングアシスタントです。",
    "metadata": {
      "session_id": "sess_123",
      "user_id": "user_456"
    }
  }
}
```

**使用シナリオ：**
- システムプロンプトの追加
- セッションメタデータの注入
- ユーザーコンテキストの追加

### 2. リクエストロガー

すべてのリクエストとレスポンスをログに記録します。

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

**使用シナリオ：**
- デバッグ
- 監査証跡
- パフォーマンス監視

### 3. レート制限

プロバイダーごとまたはグローバルでリクエストレートを制限します。

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

**使用シナリオ：**
- レート制限エラーの防止
- API使用の制御
- 乱用の防止

### 4. 圧縮 (BETA)

Token数が閾値を超えた時にコンテキストを圧縮します。

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

詳細は[コンテキスト圧縮](./compression.md)を参照してください。

### 5. セッションメモリ (BETA)

セッション間で会話メモリを維持します。

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

**使用シナリオ：**
- ユーザー設定の記憶
- 会話履歴の追跡
- セッション間でのコンテキスト維持

### 6. オーケストレーション (BETA)

リクエストを複数のプロバイダーにルーティングし、レスポンスを集約します。

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

**使用シナリオ：**
- モデル出力の比較
- 重要なリクエストの冗長性
- コンセンサスによる品質向上

## カスタムミドルウェア

### ミドルウェアインターフェース

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

### 例：カスタムヘッダー注入

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
    // レスポンス処理は不要
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

### カスタムミドルウェアのロード

#### ローカルプラグイン

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

#### リモートプラグイン

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

`http://localhost:19840/settings`でミドルウェア設定にアクセス：

1. "Middleware"タブに移動（BETAバッジ付き）
2. "Enable Middleware Pipeline"を切り替え
3. パイプラインからミドルウェアを追加/削除
4. 優先度と設定を調整
5. 個別のミドルウェアを有効化/無効化
6. "Save"をクリック

## APIエンドポイント

### ミドルウェアのリスト表示

```bash
GET /api/v1/middleware
```

レスポンス：
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

### ミドルウェアの追加

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

### ミドルウェアの更新

```bash
PUT /api/v1/middleware/{name}
Content-Type: application/json

{
  "enabled": false
}
```

### ミドルウェアの削除

```bash
DELETE /api/v1/middleware/{name}
```

### パイプラインのリロード

```bash
POST /api/v1/middleware/reload
```

## 使用シナリオ

### 開発環境

デバッグログとリクエスト検査を追加：

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

### 本番環境

レート制限と監視を追加：

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

### マルチプロバイダー比較

オーケストレーションを使用して出力を比較：

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

## ベストプラクティス

1. **適切な優先度を使用** — 小さい数字が先に実行される
2. **ミドルウェアを集中させる** — 各ミドルウェアは1つのことをうまく行うべき
3. **エラーを優雅に処理** — エラーでパイプラインを壊さない
4. **徹底的にテスト** — 本番前にミドルウェアの動作を検証
5. **パフォーマンスを監視** — ミドルウェアのオーバーヘッドを追跡
6. **設定を文書化** — 設定オプションを明確に文書化

## 制限事項

1. **パフォーマンスオーバーヘッド** — 各ミドルウェアがレイテンシを追加
2. **複雑性** — 多すぎるミドルウェアはデバッグを困難にする
3. **プラグインセキュリティ** — リモートプラグインには信頼と検証が必要
4. **エラー伝播** — ミドルウェアエラーがすべてのリクエストに影響
5. **設定の複雑性** — 複雑なパイプラインは維持が困難

## トラブルシューティング

### ミドルウェアが実行されない

1. `middleware.enabled`が`true`であることを確認
2. ミドルウェアがパイプラインで有効化されているか確認
3. 優先度が正しく設定されているか確認
4. デーモンログでミドルウェアエラーを確認

### 予期しない動作

1. ミドルウェアの実行順序（優先度）を確認
2. 設定が正しいことを確認
3. ミドルウェアを個別にテスト
4. ミドルウェアログを確認

### パフォーマンスの問題

1. 遅いミドルウェアを特定（ログを確認）
2. ミドルウェアの数を減らす
3. ミドルウェアの実装を最適化
4. 必須でないミドルウェアの無効化を検討

### プラグインのロード失敗

1. プラグインパスが正しいことを確認
2. プラグインが正しいアーキテクチャでコンパイルされているか確認
3. チェックサムが一致することを確認（リモートプラグインの場合）
4. プラグインログでエラーを確認

## セキュリティ考慮事項

1. **プラグインを検証** — 信頼できるプラグインのみをロード
2. **チェックサムを検証** — リモートプラグインのチェックサムを常に検証
3. **プラグインをサンドボックス化** — 隔離された環境でプラグインを実行することを検討
4. **ミドルウェアを監査** — デプロイ前にミドルウェアコードをレビュー
5. **動作を監視** — 予期しないミドルウェアの動作に注意

## 今後の機能強化

- クロスプラットフォーム互換性のためのWebAssemblyプラグインサポート
- コミュニティプラグインを共有するためのミドルウェアマーケットプレイス
- Web UIでのビジュアルパイプラインエディター
- ミドルウェアパフォーマンスプロファイリング
- プラグイン更新のホットリロード
- ミドルウェアテストフレームワーク
