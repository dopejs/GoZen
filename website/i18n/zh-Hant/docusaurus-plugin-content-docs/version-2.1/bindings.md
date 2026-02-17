---
sidebar_position: 5
title: 專案綁定
---

# 專案綁定

將目錄綁定到特定 profile 和/或 CLI，實現專案級自動設定。

## 使用方法

```bash
cd ~/work/company-project

# Bind profile
zen bind work-profile

# Bind CLI
zen bind --cli codex

# Bind both
zen bind work-profile --cli codex

# Check status
zen status

# Unbind
zen unbind
```

## 優先級

命令列參數 > 專案綁定 > 全域預設
