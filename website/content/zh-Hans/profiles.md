---
sidebar_position: 3
title: Profile 与故障转移
---

# Profile 与故障转移

Profile 是一组 provider 的有序列表，用于故障转移。当列表中的第一个 provider 不可用时，会自动切换到下一个。

## 配置示例

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
