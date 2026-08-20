---
sidebar_position: 4
title: シナリオルーティング
---

# シナリオルーティング

リクエストの特性に応じて、異なるプロバイダーへ自動的にルーティングします。

## 対応シナリオ

| シナリオ | 説明 |
|----------|------|
| `think` | Thinking mode が有効 |
| `image` | 画像コンテンツを含む |
| `longContext` | コンテンツがしきい値を超える |
| `webSearch` | `web_search` ツールを使用 |
| `background` | Haiku モデルを使用 |

## フォールバック機構

あるシナリオに割り当てられたすべてのプロバイダーが失敗した場合、自動的にそのプロファイルのデフォルトプロバイダーへフォールバックします。

## 設定例

```json
{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}
```
