---
sidebar_position: 9
title: 配置同步
---

# 配置同步

跨设备同步 provider、profile、默认 profile 和默认 client。认证令牌在上传前使用 AES-256-GCM（PBKDF2-SHA256 密钥派生）加密。

## 支持的后端

| Backend | Description |
|---------|-------------|
| `webdav` | 任何 WebDAV 服务器（如 Nextcloud、ownCloud） |
| `s3` | AWS S3 或 S3 兼容存储（如 MinIO、Cloudflare R2） |
| `gist` | 私有 gist（需要具有 gist 权限的 PAT） |
| `repo` | 通过 Contents API 存储到仓库文件（需要具有 repo 权限的 PAT） |

## 设置

通过 Web UI 设置页面配置同步：

```bash
# Open Web UI settings
zen web  # Settings → Config Sync
```

或通过 CLI 手动拉取：

```bash
zen config sync
```

## 配置示例

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

设置密码短语后，同步负载中的所有认证令牌将使用 AES-256-GCM 加密，密钥通过 PBKDF2-SHA256（600k 次迭代）派生。加密盐值随负载一起存储。未设置密码短语时，数据以明文 JSON 上传。

## 冲突解决

- 按实体时间戳合并：较新的修改胜出
- 删除的实体使用墓碑标记（30 天后过期）
- 标量值（默认 profile/client）：较新的时间戳胜出

## 同步范围

**已同步：** Provider（令牌加密）、Profile、默认 profile、默认 client

**不同步：** 端口设置、Web 密码、项目绑定、同步配置本身
