---
sidebar_position: 7
title: Web 管理介面
---

# Web 管理介面

透過瀏覽器視覺化管理所有設定。守護程序會在需要時自動啟動。

## 使用方法

```bash
# Open in browser (auto-starts daemon if needed)
zen web
```

## 功能

- Provider 和 Profile 管理
- 專案綁定管理
- 全域設定（預設用戶端、預設 Profile、連接埠）
- 組態同步設定
- 請求日誌檢視（支援自動重新整理）
- 模型欄位自動補全

## 安全

守護程序首次啟動時自動產生存取密碼。非本地請求（127.0.0.1/::1 以外）需要登入。

- 工作階段認證，支援可設定的到期時間
- 暴力破解保護，指數級退避
- RSA 加密敏感令牌傳輸（API 金鑰在瀏覽器端加密）
- 本地存取（127.0.0.1）免認證

### Password Management

```bash
# Reset the Web UI password
zen config reset-password

# Change password via Web UI
zen web  # Settings → Change Password
```
