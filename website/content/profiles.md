---
sidebar_position: 3
title: Profiles & Failover
---

# Profiles & Failover

A profile is an ordered list of providers for failover. When the first provider is unavailable, it automatically switches to the next one.

## Configuration Example

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

## Using Profiles

```bash
# Use default profile
zen

# Use specified profile
zen -p work

# Interactively select
zen -p
```
