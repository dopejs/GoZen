---
sidebar_position: 4
title: Szenario-Routing
---

# Szenario-Routing

Anfragen werden anhand ihrer Merkmale automatisch an unterschiedliche Anbieter geleitet.

## Unterstützte Szenarien

| Szenario | Beschreibung |
|----------|--------------|
| `think` | Nachdenk-Modus aktiviert |
| `image` | Enthält Bildinhalte |
| `longContext` | Inhalt überschreitet den Schwellenwert |
| `webSearch` | Nutzt das Werkzeug web_search |
| `background` | Nutzt das Haiku-Modell |

## Rückfallmechanismus

Schlagen alle Anbieter eines Szenarios fehl, greift automatisch die Standardliste des Profils.

## Beispielkonfiguration

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
