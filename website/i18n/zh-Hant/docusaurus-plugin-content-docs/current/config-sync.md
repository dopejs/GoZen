---
sidebar_position: 9
title: 組態同步
---

# 組態同步

跨裝置同步 provider、profile、預設 profile 與預設 client。認證令牌在上傳前使用 AES-256-GCM（PBKDF2-SHA256 金鑰衍生）加密。

## 支援的後端

| Backend | Description |
|---------|-------------|
| `webdav` | 任何 WebDAV 伺服器（如 Nextcloud、ownCloud） |
| `s3` | AWS S3 或 S3 相容儲存（如 MinIO、Cloudflare R2） |
| `gist` | 私有 gist（需要具有 gist 權限的 PAT） |
| `repo` | 透過 Contents API 儲存至倉庫檔案（需要具有 repo 權限的 PAT） |

## 設定

透過 Web UI 設定頁面配置同步：

```bash
# Open Web UI settings
zen web  # Settings → Config Sync
```

或透過 CLI 手動拉取：

```bash
zen config sync
```

## 設定範例

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

## 加密

設定密碼短語後，同步負載中的所有認證令牌將使用 AES-256-GCM 加密，金鑰透過 PBKDF2-SHA256（600k 次迭代）衍生。加密鹽值隨負載一起儲存。未設定密碼短語時，資料以明文 JSON 上傳。

## 衝突解決

- 按實體時間戳合併：較新的修改勝出
- 刪除的實體使用墓碑標記（30 天後過期）
- 純量值（預設 profile/client）：較新的時間戳勝出

## 同步範圍

**已同步：** Provider（令牌加密）、Profile、預設 profile、預設 client

**不同步：** 連接埠設定、Web 密碼、專案綁定、同步設定本身
