---
sidebar_position: 5
title: プロジェクトバインディング
---

# プロジェクトバインディング

ディレクトリを特定のプロファイルや CLI に紐付けることで、プロジェクト単位の設定を自動適用できます。

## 使い方

```bash
cd ~/work/company-project

# プロファイルをバインド
zen bind work-profile

# CLI をバインド
zen bind --cli codex

# 両方をバインド
zen bind work-profile --cli codex

# 状態を確認
zen status

# バインド解除
zen unbind
```

## 優先順位

CLI 引数 > プロジェクトバインディング > グローバルデフォルト
