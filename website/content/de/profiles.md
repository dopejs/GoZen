---
sidebar_position: 3
title: Profile und Failover
---

# Profile und Failover

Ein Profil ist eine geordnete Liste von Anbietern für das Failover. Ist der erste Anbieter nicht verfügbar, wird automatisch auf den nächsten gewechselt.

## Beispielkonfiguration

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

## Profile verwenden

```bash
# Standardprofil verwenden
zen

# Bestimmtes Profil verwenden
zen -p work

# Interaktiv auswählen
zen -p
```
