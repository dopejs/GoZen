---
sidebar_position: 1
title: はじめに
---

# はじめに

## インストール

ワンライナーでインストールできます:

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

アンインストール:

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh
```

## 初回起動

まず最初のプロバイダーを追加します:

```bash
zen config add provider
```

デフォルトプロファイルで起動:

```bash
zen
```

特定のプロファイルで起動:

```bash
zen -p work
```

特定の CLI で起動:

```bash
zen --cli codex
```
