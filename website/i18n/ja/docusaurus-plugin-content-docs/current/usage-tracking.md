---
sidebar_position: 11
title: 使用状況追跡と予算管理
---

# 使用状況追跡と予算管理

プロバイダー、モデル、プロジェクト全体でToken使用量とコストを追跡します。支出制限を設定し、自動的にアクションを実行します。

## 機能

- **リアルタイム追跡** — 各リクエストのToken使用量とコストを監視
- **多次元集計** — プロバイダー、モデル、プロジェクト、期間ごとに追跡
- **予算制限** — 日次、週次、月次の支出上限を設定
- **自動アクション** — 制限超過時に警告、ダウングレード、またはブロック
- **コスト見積もり** — すべての主要AIモデルの正確な価格設定
- **履歴データ** — SQLiteストレージ、パフォーマンス向上のため時間単位で集計

## 設定

### 使用状況追跡の有効化

```json
{
  "usage_tracking": {
    "enabled": true,
    "db_path": "~/.zen/usage.db"
  }
}
```

### モデル価格の設定

```json
{
  "pricing": {
    "models": {
      "claude-opus-4": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet-4": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4o": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    },
    "model_families": {
      "claude-opus": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    }
  }
}
```

**モデルマッチング**：まず正確なモデル名にマッチし、次にモデルファミリープレフィックスにフォールバックします。

### 予算制限の設定

```json
{
  "budget": {
    "daily": {
      "enabled": true,
      "limit": 10.0,
      "action": "warn"
    },
    "weekly": {
      "enabled": true,
      "limit": 50.0,
      "action": "downgrade"
    },
    "monthly": {
      "enabled": true,
      "limit": 200.0,
      "action": "block"
    }
  }
}
```

## 予算アクション

| アクション | 動作 |
|------|------|
| `warn` | 警告をログに記録しwebhook通知を送信するが、リクエストは許可 |
| `downgrade` | より安価なモデルに切り替え（例：opus → sonnet → haiku） |
| `block` | 429ステータスコードでリクエストを拒否 |

## Web UI

`http://localhost:19840/usage`で使用状況ダッシュボードにアクセス：

- **概要** — 現在のサイクルの総コスト、リクエスト、Token
- **プロバイダー別** — 各プロバイダーのコスト内訳
- **モデル別** — 各モデルの使用統計
- **プロジェクト別** — プロジェクトごとのコストを追跡（プロジェクトバインディング経由）
- **タイムライン** — 時間/日次のコストトレンド
- **予算ステータス** — 日次/週次/月次制限の視覚的インジケーター

## APIエンドポイント

### 使用状況サマリーの取得

```bash
GET /api/v1/usage/summary?period=daily
```

レスポンス：
```json
{
  "period": "daily",
  "start": "2026-03-05T00:00:00Z",
  "end": "2026-03-05T23:59:59Z",
  "total_cost": 8.45,
  "total_requests": 42,
  "total_input_tokens": 125000,
  "total_output_tokens": 35000,
  "by_provider": {
    "anthropic": 6.20,
    "openai": 2.25
  },
  "by_model": {
    "claude-sonnet-4": 5.10,
    "claude-opus-4": 1.10,
    "gpt-4o": 2.25
  }
}
```

### 予算ステータスの取得

```bash
GET /api/v1/budget/status
```

レスポンス：
```json
{
  "daily": {
    "enabled": true,
    "limit": 10.0,
    "spent": 8.45,
    "percent": 84.5,
    "action": "warn",
    "exceeded": false
  },
  "weekly": {
    "enabled": true,
    "limit": 50.0,
    "spent": 32.10,
    "percent": 64.2,
    "action": "downgrade",
    "exceeded": false
  },
  "monthly": {
    "enabled": true,
    "limit": 200.0,
    "spent": 145.80,
    "percent": 72.9,
    "action": "block",
    "exceeded": false
  }
}
```

### 予算制限の更新

```bash
PUT /api/v1/budget/limits
Content-Type: application/json

{
  "daily": {
    "enabled": true,
    "limit": 15.0,
    "action": "warn"
  }
}
```

## プロジェクトレベルの追跡

ディレクトリバインディングを使用してプロジェクトごとのコストを追跡：

```bash
# 現在のディレクトリをプロファイルにバインド
zen bind work-profile

# このディレクトリからのすべてのリクエストにプロジェクトパスがタグ付けされます
# Web UIの"By Project"でコストを確認
```

## Webhook通知

予算超過時にアラートを受信：

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": ["budget_warning", "budget_exceeded"]
    }
  ]
}
```

完全な設定については[Webhooks](./webhooks.md)を参照してください。

## ベストプラクティス

1. **警告から始める** — 最初は`warn`アクションを使用して使用パターンを理解
2. **現実的な制限を設定** — 履歴使用データに基づいて制限を設定
3. **開発時はダウングレードを使用** — テスト時に自動的により安価なモデルに切り替え
4. **本番環境ではブロックを保持** — ハード支出上限にのみ`block`アクションを使用
5. **毎日監視** — 予期しない事態を避けるため使用ダッシュボードを定期的に確認
6. **Webhookを有効化** — 制限に近づいた時にリアルタイムアラートを取得

## トラブルシューティング

### 使用状況が追跡されない

1. 設定で`usage_tracking.enabled`が`true`であることを確認
2. データベースパスが書き込み可能であることを確認：`~/.zen/usage.db`
3. デーモンを再起動：`zen daemon restart`

### コストが不正確

1. 設定のモデル価格が現在のレートと一致することを確認
2. モデル名のマッチングを確認（完全一致 vs ファミリープレフィックス）
3. プロバイダーがレートを変更した場合、価格設定を更新

### 予算が実行されない

1. 予算設定が有効化されているか確認
2. アクションが設定されているか確認（`warn`、`downgrade`、または`block`）
3. デーモンログで予算チェッカーエラーを確認

## パフォーマンス

- **時間単位の集計** — 生データは時間単位で集計してクエリ負荷を削減
- **インデックス付きクエリ** — データベースはプロバイダー、モデル、プロジェクト、タイムスタンプにインデックス
- **効率的なストレージ** — リクエストあたり約1KB、30,000リクエストで約30MB
- **高速ダッシュボード** — 典型的な使用パターンでクエリ時間は1秒未満
