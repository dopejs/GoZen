---
sidebar_position: 9
title: 設定同期
---

# 設定同期

プロバイダー、プロファイル、デフォルトプロファイル、デフォルトクライアントをデバイス間で同期できます。認証トークンはアップロード前に AES-256-GCM（PBKDF2-SHA256 による鍵導出）で暗号化されます。

## 対応バックエンド

| バックエンド | 説明 |
|-------------|------|
| `webdav` | 任意の WebDAV サーバー（例: Nextcloud、ownCloud） |
| `s3` | AWS S3 または S3 互換ストレージ（例: MinIO、Cloudflare R2） |
| `gist` | 非公開 gist（`gist` スコープ付き PAT が必要） |
| `repo` | Contents API 経由のリポジトリファイル（`repo` スコープ付き PAT が必要） |

## セットアップ

Web UI の設定ページから同期を設定します:

```bash
# Web UI の設定を開く
zen web  # Settings → Config Sync
```

CLI から手動で pull することもできます:

```bash
zen config sync
```

## 設定例

```json
{
  "sync": {
    "backend": "gist",
    "gist_id": "abc123def456",
    "token": "ghp_xxxxxxxxxxxx",
    "passphrase": "my-secret-passphrase",
    "auto_pull": true,
    "pull_interval": 300
  }
}
```

## 暗号化

passphrase を設定すると、同期ペイロード内のすべての認証トークンは PBKDF2-SHA256（60 万回反復）で導出した鍵を用いて AES-256-GCM で暗号化されます。暗号化に使う salt はペイロードと一緒に保存されます。passphrase が未設定の場合、データは平文 JSON のままアップロードされます。

## 競合解決

- エンティティ単位のタイムスタンプマージ: 新しい変更が優先
- 削除済みエンティティは tombstone を使用（30 日後に失効）
- スカラー値（デフォルトプロファイル / デフォルトクライアント）: 新しいタイムスタンプが優先

## 同期対象

**同期されるもの:** プロバイダー（暗号化されたトークンを含む）、プロファイル、デフォルトプロファイル、デフォルトクライアント

**同期されないもの:** ポート設定、Web UI パスワード、プロジェクトバインディング、同期設定そのもの
