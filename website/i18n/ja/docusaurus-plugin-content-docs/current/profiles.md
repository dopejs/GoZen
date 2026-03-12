---
sidebar_position: 3
title: プロファイルとフェイルオーバー
---

# プロファイルとフェイルオーバー

プロファイルはフェイルオーバー用に順序付けされたプロバイダーのリストです。先頭のプロバイダーが利用できない場合、自動的に次のプロバイダーへ切り替わります。

## 設定例

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-main", "anthropic-backup"]
    },
    "work": {
      "providers": ["company-api"],
      "routing": {
        "think": {"providers": [{"name": "thinking-api"}]}
      }
    }
  }
}
```

## プロファイルの使い方

```bash
# デフォルトプロファイルを使う
zen

# 指定したプロファイルを使う
zen -p work

# 対話的に選択
zen -p
```
