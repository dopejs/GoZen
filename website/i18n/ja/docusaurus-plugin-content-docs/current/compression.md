---
sidebar_position: 14
title: コンテキスト圧縮 (BETA)
---

# コンテキスト圧縮 (BETA)

:::warning BETA機能
コンテキスト圧縮は現在ベータ版です。デフォルトでは無効になっており、明示的な設定が必要です。
:::

Token数が閾値を超えた時に会話コンテキストを自動圧縮し、会話品質を維持しながらコストを削減します。

## 機能

- **自動圧縮** — Token数が閾値を超えた時にトリガー
- **インテリジェント要約** — 安価なモデル（claude-3-haiku）を使用して古いメッセージを要約
- **最近のメッセージを保持** — コンテキストの連続性を保つため最近のメッセージを完全に保持
- **Token推定** — API呼び出し前に正確なToken数を計算
- **統計追跡** — 圧縮効果を監視
- **透過的な操作** — すべてのAIクライアントとシームレスに連携

## 動作原理

1. **Token推定** — 会話履歴のToken数を計算
2. **閾値チェック** — 設定された閾値と比較（デフォルト：50,000）
3. **メッセージ選択** — 圧縮が必要な古いメッセージを識別
4. **要約生成** — 安価なモデルを使用して簡潔な要約を作成
5. **コンテキスト置換** — 古いメッセージを要約で置き換え
6. **リクエスト転送** — 圧縮されたコンテキストをターゲットモデルに送信

## 設定

### 圧縮の有効化

```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 50000,
    "target_tokens": 20000,
    "summarizer_model": "claude-3-haiku-20240307",
    "preserve_recent_messages": 5,
    "tokens_per_char": 0.25
  }
}
```

**オプション：**

| オプション | デフォルト値 | 説明 |
|------|--------|------|
| `enabled` | `false` | コンテキスト圧縮を有効化 |
| `threshold_tokens` | `50000` | この値を超えた時に圧縮をトリガー |
| `target_tokens` | `20000` | 圧縮後の目標Token数 |
| `summarizer_model` | `claude-3-haiku-20240307` | 要約に使用するモデル |
| `preserve_recent_messages` | `5` | 完全に保持する最近のメッセージ数 |
| `tokens_per_char` | `0.25` | Token数推定の比率 |

### プロファイルごとの設定

特定のプロファイルで圧縮を有効化：

```json
{
  "profiles": {
    "long-context": {
      "providers": ["anthropic"],
      "compression": {
        "enabled": true,
        "threshold_tokens": 100000,
        "target_tokens": 40000
      }
    },
    "short-context": {
      "providers": ["openai"],
      "compression": {
        "enabled": false
      }
    }
  }
}
```

## Token推定

GoZenは高速なToken数計算のために文字ベースの推定を使用：

```
estimated_tokens = character_count * tokens_per_char
```

**デフォルト比率：** 文字あたり0.25 Token（1 Token ≈ 4文字）

**精度：** 英語テキストで±10%、他の言語では異なる場合があります

正確なToken数計算のため、GoZenは利用可能な場合`tiktoken-go`ライブラリを使用します。

## 圧縮戦略

### メッセージ選択

1. **システムメッセージ** — 常に保持
2. **最近のメッセージ** — 最後のN件のメッセージを保持（デフォルト：5）
3. **古いメッセージ** — 圧縮候補

### 要約プロンプト

```
以下の会話履歴を簡潔に要約し、重要な情報、決定、コンテキストを保持してください：

[古いメッセージ]

要点を捉えた短い要約を提供してください。
```

### 結果

```
元：45,000 tokens（30メッセージ）
圧縮後：22,000 tokens（要約 + 5件の最近のメッセージ）
節約：23,000 tokens（51%）
```

## Web UI

`http://localhost:19840/settings`で圧縮設定にアクセス：

1. "Compression"タブに移動（BETAバッジ付き）
2. "Enable Compression"を切り替え
3. 閾値と目標Token数を調整
4. 要約モデルを選択
5. 保持する最近のメッセージ数を設定
6. "Save"をクリック

### 統計ダッシュボード

圧縮統計を表示：

- **総圧縮回数** — 圧縮がトリガーされた回数
- **節約されたToken** — すべての圧縮で節約された総Token数
- **平均節約** — 圧縮ごとの平均Token削減量
- **圧縮率** — 圧縮がトリガーされたリクエストの割合

## APIエンドポイント

### 圧縮統計の取得

```bash
GET /api/v1/compression/stats
```

レスポンス：
```json
{
  "enabled": true,
  "total_compressions": 42,
  "tokens_saved": 1250000,
  "average_savings": 29761,
  "compression_rate": 0.15,
  "last_compression": "2026-03-05T10:30:00Z"
}
```

### 圧縮設定の更新

```bash
PUT /api/v1/compression/settings
Content-Type: application/json

{
  "enabled": true,
  "threshold_tokens": 60000,
  "target_tokens": 25000
}
```

### 統計のリセット

```bash
POST /api/v1/compression/stats/reset
```

## 使用シナリオ

### 長時間のコーディングセッション

**シナリオ：** Claude Codeで数時間のコーディングセッション

**設定：**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 80000,
    "target_tokens": 30000,
    "preserve_recent_messages": 10
  }
}
```

**利点：** コンテキスト制限に達することなく会話の連続性を維持

### バッチ処理

**シナリオ：** AIで複数のドキュメントを処理

**設定：**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 40000,
    "target_tokens": 15000,
    "preserve_recent_messages": 3
  }
}
```

**利点：** 大規模なドキュメントセットを処理する際のコスト削減

### 研究と分析

**シナリオ：** 複数のトピックにわたる長時間の研究セッション

**設定：**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 100000,
    "target_tokens": 40000,
    "preserve_recent_messages": 8
  }
}
```

**利点：** 初期のコンテキストを保持しながら最近のトピックに会話を集中

## ベストプラクティス

1. **デフォルト値から開始** — デフォルト設定はほとんどの使用シナリオに適しています
2. **統計を監視** — 圧縮率と節約を定期的に確認
3. **閾値を調整** — 長コンテキストモデル（Claude Opus）では増やし、短コンテキストでは減らす
4. **十分なメッセージを保持** — コンテキストの連続性のため5-10件の最近のメッセージを保持
5. **安価な要約器を使用** — Haikuは高速でコスト効率が良く、要約に適しています
6. **本番前にテスト** — 特定の使用ケースで圧縮品質を検証

## 制限事項

1. **品質の損失** — 要約により微妙な詳細が失われる可能性
2. **レイテンシの増加** — 要約API呼び出しのオーバーヘッドを追加
3. **コストのトレードオフ** — 要約コスト vs. Token節約
4. **言語サポート** — 英語に最適、他の言語では異なる場合があります
5. **コンテキストウィンドウ** — モデルの最大コンテキストウィンドウを超えることはできません

## トラブルシューティング

### 圧縮がトリガーされない

1. `compression.enabled`が`true`であることを確認
2. Token数が閾値を超えているか確認
3. 会話に圧縮可能な十分なメッセージがあることを確認
4. デーモンログで圧縮エラーを確認

### 要約品質が低い

1. 異なる要約モデルを試す（例：claude-3-sonnet）
2. より多くのコンテキストを保持するため`preserve_recent_messages`を増やす
3. より長い要約を許可するため`target_tokens`を調整
4. 要約モデルが利用可能で正常に動作しているか確認

### レイテンシの増加

1. 圧縮により追加のAPI呼び出し（要約）が1回追加されます
2. より高速な要約モデルを使用（haikuが最速）
3. 圧縮頻度を減らすため閾値を増やす
4. レイテンシに敏感なアプリケーションでは圧縮を無効化することを検討

### 予期しないコスト

1. 使用ダッシュボードで要約コストを監視
2. 節約 vs. 要約コストを比較
3. 圧縮頻度を減らすため閾値を調整
4. 要約に最も安価な利用可能モデルを使用

## パフォーマンスへの影響

- **Token推定** — リクエストあたり約1ms（無視できる）
- **要約生成** — 1-3秒（モデルとメッセージ数に依存）
- **メモリオーバーヘッド** — 最小（圧縮あたり約1KB）
- **コスト節約** — 通常30-50%のToken削減

## 高度な設定

### カスタム要約プロンプト

```json
{
  "compression": {
    "enabled": true,
    "custom_prompt": "以下の会話の技術的な要約を作成し、コード変更、決定、アクションアイテムに焦点を当ててください：\n\n{messages}\n\n要約："
  }
}
```

### 条件付き圧縮

特定のシナリオでのみ圧縮を有効化：

```json
{
  "profiles": {
    "default": {
      "scenarios": {
        "longContext": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": true,
            "threshold_tokens": 100000
          }
        },
        "default": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": false
          }
        }
      }
    }
  }
}
```

### 多段階圧縮

非常に長い会話のための複数回の圧縮：

```json
{
  "compression": {
    "enabled": true,
    "stages": [
      {
        "threshold_tokens": 50000,
        "target_tokens": 30000
      },
      {
        "threshold_tokens": 80000,
        "target_tokens": 40000
      }
    ]
  }
}
```

## 今後の機能強化

- インテリジェントなメッセージ選択のためのセマンティック類似度マッチング
- 品質比較のためのマルチモデル要約
- 圧縮品質メトリクスとフィードバック
- 各使用ケースに合わせたカスタム圧縮戦略
- 外部コンテキストストレージのためのRAG統合
