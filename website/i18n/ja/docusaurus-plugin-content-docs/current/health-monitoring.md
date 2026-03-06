---
sidebar_position: 12
title: ヘルスモニタリングとロードバランシング
---

# ヘルスモニタリングとロードバランシング

プロバイダーのヘルス状態をリアルタイムで監視し、リクエストを最適な利用可能プロバイダーに自動的にルーティングします。

## 機能

- **リアルタイムヘルスチェック** — 設定可能なチェック間隔での定期的なヘルス監視
- **成功率追跡** — リクエスト成功率に基づいてプロバイダーのヘルス状態を計算
- **レイテンシ監視** — 各プロバイダーの平均応答時間を追跡
- **複数の戦略** — フェイルオーバー、ラウンドロビン、最低レイテンシ、最低コスト
- **自動フェイルオーバー** — プライマリプロバイダーが不健全な時にバックアッププロバイダーに切り替え
- **ヘルスダッシュボード** — Web UIでの視覚的なステータスインジケーター

## 設定

### ヘルスモニタリングの有効化

```json
{
  "health_check": {
    "enabled": true,
    "interval": "5m",
    "timeout": "10s",
    "endpoint": "/v1/messages",
    "method": "POST"
  }
}
```

**オプション：**
- `interval` — プロバイダーのヘルス状態をチェックする頻度（デフォルト：5分）
- `timeout` — ヘルスチェックのリクエストタイムアウト（デフォルト：10秒）
- `endpoint` — テストするAPIエンドポイント（デフォルト：`/v1/messages`）
- `method` — ヘルスチェックのHTTPメソッド（デフォルト：`POST`）

### ロードバランシングの設定

```json
{
  "load_balancing": {
    "strategy": "least-latency",
    "health_aware": true,
    "cache_ttl": "30s"
  }
}
```

## ロードバランシング戦略

### 1. フェイルオーバー（デフォルト）

プロバイダーを順番に使用し、失敗時に次のプロバイダーに切り替えます。

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-primary", "anthropic-backup", "openai"],
      "load_balancing": {
        "strategy": "failover"
      }
    }
  }
}
```

**動作：**
1. `anthropic-primary`を試行
2. 失敗した場合、`anthropic-backup`を試行
3. 失敗した場合、`openai`を試行
4. すべて失敗した場合、エラーを返す

**最適な用途：** 明確なプライマリ/バックアップ階層を持つ本番ワークロード

### 2. ラウンドロビン

すべての健全なプロバイダー間でリクエストを均等に分散します。

```json
{
  "load_balancing": {
    "strategy": "round-robin"
  }
}
```

**動作：**
- リクエスト1 → プロバイダーA
- リクエスト2 → プロバイダーB
- リクエスト3 → プロバイダーC
- リクエスト4 → プロバイダーA（循環）

**最適な用途：** レート制限を回避するため複数のアカウント間で負荷を分散

### 3. 最低レイテンシ

平均レイテンシが最も低いプロバイダーにルーティングします。

```json
{
  "load_balancing": {
    "strategy": "least-latency"
  }
}
```

**動作：**
- 各プロバイダーの平均応答時間を追跡
- 最速のプロバイダーにルーティング
- 30秒ごとにメトリクスを更新（`cache_ttl`で設定可能）

**最適な用途：** レイテンシに敏感なアプリケーション、リアルタイムインタラクション

### 4. 最低コスト

リクエストされたモデルに対して最も安価なプロバイダーにルーティングします。

```json
{
  "load_balancing": {
    "strategy": "least-cost"
  }
}
```

**動作：**
- プロバイダー間の価格を比較
- 最も安価なオプションにルーティング
- 入力と出力の両方のTokenコストを考慮

**最適な用途：** コスト最適化、バッチ処理

## ヘルスステータス

プロバイダーは4つのヘルスステータスに分類されます：

| ステータス | 成功率 | 動作 |
|------|--------|------|
| **健全** | ≥ 95% | 通常の優先度 |
| **低下** | 70-95% | 低い優先度、まだ使用可能 |
| **不健全** | < 70% | スキップ、健全なプロバイダーがない場合を除く |
| **不明** | データなし | 初期状態では健全として扱う |

### ヘルス認識ルーティング

`health_aware: true`（デフォルト）の場合：
- 健全なプロバイダーが優先される
- 低下したプロバイダーはバックアップとして使用
- 不健全なプロバイダーは、他のすべてが失敗しない限りスキップ

## Web UIダッシュボード

`http://localhost:19840/health`でヘルスダッシュボードにアクセス：

### プロバイダーステータス

- **ステータスインジケーター** — 緑（健全）、黄（低下）、赤（不健全）
- **成功率** — 成功したリクエストの割合
- **平均レイテンシ** — 平均応答時間（ミリ秒）
- **最終チェック** — 最後のヘルスチェックのタイムスタンプ
- **エラー数** — 最近の失敗回数

### メトリクスタイムライン

- **レイテンシグラフ** — 時間経過に伴う応答時間のトレンド
- **成功率グラフ** — 時間経過に伴うヘルス状態のトレンド
- **リクエスト量** — プロバイダーごとのリクエスト数

## APIエンドポイント

### プロバイダーヘルス状態の取得

```bash
GET /api/v1/health/providers
```

レスポンス：
```json
{
  "providers": [
    {
      "name": "anthropic-primary",
      "status": "healthy",
      "success_rate": 98.5,
      "avg_latency_ms": 1250,
      "last_check": "2026-03-05T10:30:00Z",
      "error_count": 2,
      "total_requests": 150
    },
    {
      "name": "openai-backup",
      "status": "degraded",
      "success_rate": 85.0,
      "avg_latency_ms": 2100,
      "last_check": "2026-03-05T10:29:00Z",
      "error_count": 15,
      "total_requests": 100
    }
  ]
}
```

### プロバイダーメトリクスの取得

```bash
GET /api/v1/health/providers/{name}/metrics?period=1h
```

レスポンス：
```json
{
  "provider": "anthropic-primary",
  "period": "1h",
  "metrics": [
    {
      "timestamp": "2026-03-05T10:00:00Z",
      "latency_ms": 1200,
      "success_rate": 99.0,
      "requests": 25
    },
    {
      "timestamp": "2026-03-05T10:05:00Z",
      "latency_ms": 1300,
      "success_rate": 98.0,
      "requests": 28
    }
  ]
}
```

### 手動ヘルスチェックのトリガー

```bash
POST /api/v1/health/check
Content-Type: application/json

{
  "provider": "anthropic-primary"
}
```

## Webhook通知

プロバイダーのステータス変更時にアラートを受信：

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": ["provider_down", "provider_up", "failover"]
    }
  ]
}
```

**イベントタイプ：**
- `provider_down` — プロバイダーが不健全になった
- `provider_up` — プロバイダーが健全な状態に回復
- `failover` — リクエストがバックアッププロバイダーにフェイルオーバー

## シナリオベースのルーティング

ヘルスモニタリングとシナリオルーティングを組み合わせてインテリジェントなリクエスト分散を実現：

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-primary", "anthropic-backup"],
      "scenarios": {
        "thinking": {
          "providers": ["anthropic-thinking"],
          "load_balancing": {
            "strategy": "least-latency"
          }
        },
        "image": {
          "providers": ["anthropic-vision", "openai-vision"],
          "load_balancing": {
            "strategy": "failover"
          }
        }
      }
    }
  }
}
```

詳細は[シナリオルーティング](./routing.md)を参照してください。

## ベストプラクティス

1. **適切な間隔を設定** — ほとんどの場合5分で十分、重要なシステムでは1分を使用
2. **ヘルス認識ルーティングを使用** — 本番ワークロードでは常に有効化
3. **低下したプロバイダーを監視** — 成功率が95%を下回った時に調査
4. **戦略を組み合わせる** — プライマリ/バックアップにはフェイルオーバー、負荷分散にはラウンドロビンを使用
5. **Webhookを有効化** — プロバイダーダウン時に即座に通知を受ける
6. **ダッシュボードを定期的に確認** — ヘルストレンドを確認してパターンを識別

## トラブルシューティング

### ヘルスチェックが失敗する

1. プロバイダーのAPIキーが有効であることを確認
2. プロバイダーエンドポイントへのネットワーク接続を確認
3. プロバイダーの応答が遅い場合、タイムアウトを増やす：`"timeout": "30s"`
4. デーモンログで具体的なエラーメッセージを確認

### レイテンシメトリクスが不正確

1. レイテンシにはネットワーク時間 + API処理時間が含まれます
2. プロキシやVPNがオーバーヘッドを追加していないか確認
3. メトリクスはデフォルトで30秒キャッシュされます（`cache_ttl`で設定可能）

### フェイルオーバーが機能しない

1. ロードバランシング設定で`health_aware: true`であることを確認
2. 設定ファイルでバックアッププロバイダーが設定されているか確認
3. ヘルスチェックが有効で実行中であることを確認
4. Web UIまたはログでフェイルオーバーイベントを確認

### プロバイダーが不健全な状態でスタック

1. API経由で手動でヘルスチェックをトリガー
2. プロバイダーが実際にダウンしているか確認（curlでテスト）
3. デーモンを再起動してヘルス状態をリセット：`zen daemon restart`
4. エラーログで根本原因を確認

## パフォーマンスへの影響

- **ヘルスチェック** — 最小限のオーバーヘッド、バックグラウンドgoroutineで実行
- **メトリクスキャッシュ** — 30秒のTTLでデータベースクエリを削減
- **アトミック操作** — 並行リクエストのためのスレッドセーフカウンター
- **ノンブロッキング** — ヘルスチェックはリクエスト処理をブロックしない

## 高度な設定

### カスタムヘルスチェックペイロード

```json
{
  "health_check": {
    "enabled": true,
    "custom_payload": {
      "model": "claude-3-haiku-20240307",
      "max_tokens": 10,
      "messages": [
        {
          "role": "user",
          "content": "ping"
        }
      ]
    }
  }
}
```

### プロバイダーごとのヘルス設定

```json
{
  "providers": {
    "anthropic-primary": {
      "health_check": {
        "interval": "1m",
        "timeout": "5s"
      }
    },
    "openai-backup": {
      "health_check": {
        "interval": "5m",
        "timeout": "10s"
      }
    }
  }
}
```
