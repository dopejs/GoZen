---
sidebar_position: 4
title: Routage par scénario
---

# Routage par scénario

Acheminez automatiquement les requêtes vers différents fournisseurs selon leurs caractéristiques.

## Scénarios pris en charge

| Scénario | Description |
|----------|-------------|
| `think` | Mode raisonnement activé |
| `image` | Contient du contenu image |
| `longContext` | Le contenu dépasse le seuil |
| `webSearch` | Utilise l’outil web_search |
| `background` | Utilise le modèle Haiku |

## Mécanisme de repli

Si tous les fournisseurs d’un scénario échouent, le routage revient automatiquement aux fournisseurs par défaut du profil.

## Exemple de configuration

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
