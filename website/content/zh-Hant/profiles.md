---
sidebar_position: 3
title: Profile 與故障轉移
---

# Profile 與故障轉移

Profile 是一組 provider 的有序列表，用於故障轉移。當列表中的第一個 provider 不可用時，會自動切換到下一個。

## 設定範例

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

## 使用 Profile

```bash
# Use default profile
zen

# Use specified profile
zen -p work

# Interactively select
zen -p
```
