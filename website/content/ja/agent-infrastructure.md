---
sidebar_position: 16
title: エージェントインフラストラクチャ (BETA)
---

# エージェントインフラストラクチャ (BETA)

:::warning BETA機能
エージェントインフラストラクチャは現在ベータ版です。デフォルトでは無効になっており、明示的な設定が必要です。
:::

自律エージェントワークフローの組み込みサポート。セッション管理、ファイル調整、リアルタイム監視、セキュリティ制御を含みます。

## 機能

- **エージェントランタイム** — 完全なライフサイクル管理を備えた自律エージェントタスクの実行
- **オブザーバトリー** — エージェントセッションとアクティビティのリアルタイム監視
- **ガードレール** — エージェント動作のセキュリティ制御と制約
- **コーディネーター** — ファイルベースのマルチエージェントワークフロー調整
- **タスクキュー** — 優先度と依存関係を持つエージェントタスクの管理
- **セッション管理** — 複数のプロジェクトにわたるエージェントセッションの追跡

## アーキテクチャ

```
エージェントクライアント (Claude Code, Codex など)
    ↓
エージェントランタイム
    ↓
┌─────────────┬──────────────┬─────────────┐
│ オブザーバ  │ ガードレール │ コーディネー│
│ トリー      │              │ ター        │
│ (監視)      │ (セキュリティ)│ (同期)      │
└─────────────┴──────────────┴─────────────┘
    ↓
タスクキュー → プロバイダーAPI
```

## 設定

### エージェントインフラストラクチャの有効化

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

## コンポーネント

### 1. エージェントランタイム

エージェントタスク実行のライフサイクルを管理します。

**機能：**
- タスクのスケジューリングと実行
- 並行タスク管理
- タイムアウト処理
- 自動クリーンアップ
- エラー回復

**設定：**
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
# エージェントタスクの開始
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

# タスクステータスの取得
GET /api/v1/agent/tasks/{task_id}

# タスクのキャンセル
DELETE /api/v1/agent/tasks/{task_id}
```

### 2. オブザーバトリー

エージェントアクティビティのリアルタイム監視。

**機能：**
- セッション追跡
- アクティビティログ
- パフォーマンスメトリクス
- ステータス更新
- 履歴データ

**設定：**
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

**監視メトリクス：**
- アクティブセッション
- 進行中のタスク
- Token使用量
- API呼び出し
- ファイル操作
- エラー率
- 平均レイテンシ

**API：**
```bash
# すべてのアクティブセッションを取得
GET /api/v1/agent/sessions

# セッション詳細を取得
GET /api/v1/agent/sessions/{session_id}

# セッションメトリクスを取得
GET /api/v1/agent/sessions/{session_id}/metrics
```

### 3. ガードレール

エージェント動作のセキュリティ制御と制約。

**機能：**
- 操作制限
- パス制限
- コマンドブロック
- リソースクォータ
- 承認ワークフロー

**設定：**
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

**実行メカニズム：**
- 実行前の検証
- リアルタイム監視
- 自動ブロック
- 承認プロンプト
- 監査ログ

**API：**
```bash
# ガードレールステータスの取得
GET /api/v1/agent/guardrails

# ガードレールルールの更新
PUT /api/v1/agent/guardrails
Content-Type: application/json

{
  "max_file_operations": 200,
  "blocked_commands": ["rm -rf", "sudo", "dd"]
}
```

### 4. コーディネーター

ファイルベースのマルチエージェントワークフロー調整。

**機能：**
- ファイルロック
- 変更検出
- 競合解決
- ステート同期
- イベント通知

**設定：**
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

**使用シナリオ：**
- 複数のエージェントが同じファイルを編集
- 並行変更の防止
- 外部ファイル変更の検出
- エージェントワークフローの調整

**API：**
```bash
# ファイルロックの取得
POST /api/v1/agent/locks
Content-Type: application/json

{
  "path": "/path/to/file.go",
  "session_id": "sess_123",
  "timeout": "5m"
}

# ファイルロックの解放
DELETE /api/v1/agent/locks/{lock_id}

# ファイル変更イベントの取得
GET /api/v1/agent/changes?since=2026-03-05T10:00:00Z
```

### 5. タスクキュー

優先度と依存関係を持つエージェントタスクの管理。

**機能：**
- 優先度スケジューリング
- タスク依存関係
- キュー管理
- ステータス追跡
- リトライロジック

**設定：**
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
# キューにタスクを追加
POST /api/v1/agent/queue
Content-Type: application/json

{
  "name": "run-tests",
  "priority": 2,
  "depends_on": ["build-project"],
  "config": {}
}

# キューステータスの取得
GET /api/v1/agent/queue

# キューからタスクを削除
DELETE /api/v1/agent/queue/{task_id}
```

## Web UI

エージェントダッシュボードにアクセス：`http://localhost:19840/agent`

### セッションタブ

- **アクティブセッション** — 現在実行中のエージェントセッション
- **セッション詳細** — タスク進捗、メトリクス、ログ
- **セッション制御** — 一時停止、再開、キャンセル

### タスクタブ

- **タスクキュー** — 保留中および進行中のタスク
- **タスク履歴** — 完了および失敗したタスク
- **タスク詳細** — 設定、ログ、結果

### ガードレールタブ

- **操作制限** — 現在の使用量 vs. 制限
- **ブロックされた操作** — 最近ブロックされた試行
- **承認キュー** — 承認待ちの操作

### メトリクスタブ

- **Token使用量** — セッションごとおよび合計
- **API呼び出し** — リクエスト数とレート
- **ファイル操作** — 読み取り/書き込み/削除数
- **パフォーマンス** — レイテンシとスループット

## Claude Codeとの統合

GoZenはClaude Codeセッションを自動検出し、エージェントインフラストラクチャを提供します：

```bash
# エージェントサポート付きでClaude Codeを起動
zen --agent

# エージェント機能が自動的に有効化されます：
# - セッション追跡
# - ファイル調整
# - ガードレール実行
# - リアルタイム監視
```

**利点：**
- 並行ファイル変更の防止
- Token使用量とコストの追跡
- セキュリティ制約の実行
- エージェントアクティビティの監視
- マルチエージェントワークフローの調整

## 使用シナリオ

### マルチエージェント開発

複数のエージェントが同じコードベースで作業：

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

### 長時間実行タスク

長時間実行エージェントタスクの監視と制御：

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

### セキュリティクリティカルな操作

厳格なセキュリティ制御の実行：

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

## ベストプラクティス

1. **ガードレールを有効化** — 本番環境では常にガードレールを使用
2. **適切な制限を設定** — 使用シナリオに基づいて制限を設定
3. **積極的に監視** — オブザーバトリーダッシュボードを定期的に確認
4. **ファイルロックを使用** — マルチエージェントワークフローではコーディネーターを有効化
5. **承認を設定** — 破壊的操作には承認を要求
6. **ログをレビュー** — エージェントアクティビティを定期的に監査

## 制限事項

1. **パフォーマンスオーバーヘッド** — 監視と調整によりレイテンシが増加
2. **ファイルロック** — マルチエージェントシナリオで遅延が発生する可能性
3. **メモリ使用量** — セッション履歴がメモリを消費
4. **複雑性** — エージェントワークフローの理解が必要
5. **ベータステータス** — 将来のバージョンで機能が変更される可能性

## トラブルシューティング

### エージェントセッションが追跡されない

1. `agent.enabled`が`true`であることを確認
2. オブザーバトリーが有効化されていることを確認
3. エージェントクライアントがサポートされていることを確認（Claude Code、Codex）
4. デーモンログでエラーを確認

### ファイルロックの問題

1. コーディネーターが有効化されていることを確認
2. ロックタイムアウトが適切であることを確認
3. アクティブなロックを確認：`GET /api/v1/agent/locks`
4. 必要に応じてスタックしたロックを手動で解放

### ガードレールが実行されない

1. ガードレールが有効化されていることを確認
2. ルール設定が正しいことを確認
3. ブロックされた操作のログを確認
4. エージェントクライアントがガードレールに従っていることを確認

### 高メモリ使用量

1. 履歴保持期間を短縮
2. 更新間隔を長くする
3. 最大並行タスク数を制限
4. 自動クリーンアップを有効化

## セキュリティ考慮事項

1. **パス制限** — 常に許可/ブロックパスを設定
2. **コマンドブロック** — 危険なコマンドをブロック
3. **承認ワークフロー** — 機密操作には承認を要求
4. **監査ログ** — 包括的なログ記録を有効化
5. **リソース制限** — 適切な操作制限を設定

## 今後の機能強化

- マルチエージェント協調プロトコル
- 高度な競合解決戦略
- 異常検出のための機械学習
- 外部監視ツールとの統合
- エージェント動作分析
- 自動セキュリティポリシー生成
