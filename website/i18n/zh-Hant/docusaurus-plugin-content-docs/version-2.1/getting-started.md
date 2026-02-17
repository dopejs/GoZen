---
sidebar_position: 1
title: 快速開始
---

# 快速開始

## 安裝

使用安裝腳本一鍵安裝：

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

解除安裝：

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh
```

## 首次使用

新增第一個 provider：

```bash
zen config add provider
```

使用預設 profile 啟動：

```bash
zen
```

使用指定 profile 啟動：

```bash
zen -p work
```

使用指定 CLI 啟動：

```bash
zen --cli codex
```
