---
sidebar_position: 11
title: Suivi de l’usage et budget
---

# Suivi de l’usage et contrôle du budget

Suivez la consommation de jetons et les coûts par fournisseur, modèle et projet. Fixez des plafonds de dépense appliqués automatiquement.

## Fonctionnalités

- **Suivi en temps réel** — surveillez jetons et coûts requête par requête
- **Agrégation multidimensionnelle** — par fournisseur, modèle, projet et période
- **Plafonds budgétaires** — limites de dépense quotidiennes, hebdomadaires et mensuelles
- **Actions automatiques** — avertir, rétrograder ou bloquer les requêtes au dépassement
- **Estimation des coûts** — tarification précise pour tous les grands modèles
- **Historique** — stockage SQLite avec agrégation horaire pour la performance

## Configuration

### Activer le suivi de l’usage

```json
{
  "usage_tracking": {
    "enabled": true,
    "db_path": "~/.zen/usage.db"
  }
}
```

### Configurer la tarification des modèles

```json
{
  "pricing": {
    "models": {
      "claude-opus-4": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet-4": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4o": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    },
    "model_families": {
      "claude-opus": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    }
  }
}
```

**Correspondance des modèles** : les noms exacts sont testés d’abord, puis les préfixes de famille de modèles.

### Définir les plafonds budgétaires

```json
{
  "budget": {
    "daily": {
      "enabled": true,
      "limit": 10.0,
      "action": "warn"
    },
    "weekly": {
      "enabled": true,
      "limit": 50.0,
      "action": "downgrade"
    },
    "monthly": {
      "enabled": true,
      "limit": 200.0,
      "action": "block"
    }
  }
}
```

## Actions budgétaires

| Action | Comportement |
|--------|--------------|
| `warn` | Journalise un avertissement et envoie une notification webhook, mais laisse passer la requête |
| `downgrade` | Bascule vers un modèle moins cher (par exemple opus → sonnet → haiku) |
| `block` | Rejette la requête avec le code d’état 429 |

## Interface web

Accédez au tableau de bord d’usage sur `http://localhost:19840/usage` :

- **Vue d’ensemble** — coût total, requêtes et jetons de la période en cours
- **Par fournisseur** — répartition des coûts par fournisseur
- **Par modèle** — statistiques d’usage par modèle
- **Par projet** — coûts par projet (via les liaisons de projet)
- **Chronologie** — tendances de coût horaires et quotidiennes
- **État du budget** — indicateurs visuels des limites quotidiennes, hebdomadaires et mensuelles

## Points d’API

### Obtenir le résumé d’usage

```bash
GET /api/v1/usage/summary?period=daily
```

Réponse :
```json
{
  "period": "daily",
  "start": "2026-03-05T00:00:00Z",
  "end": "2026-03-05T23:59:59Z",
  "total_cost": 8.45,
  "total_requests": 42,
  "total_input_tokens": 125000,
  "total_output_tokens": 35000,
  "by_provider": {
    "anthropic": 6.20,
    "openai": 2.25
  },
  "by_model": {
    "claude-sonnet-4": 5.10,
    "claude-opus-4": 1.10,
    "gpt-4o": 2.25
  }
}
```

### Obtenir l’état du budget

```bash
GET /api/v1/budget/status
```

Réponse :
```json
{
  "daily": {
    "enabled": true,
    "limit": 10.0,
    "spent": 8.45,
    "percent": 84.5,
    "action": "warn",
    "exceeded": false
  },
  "weekly": {
    "enabled": true,
    "limit": 50.0,
    "spent": 32.10,
    "percent": 64.2,
    "action": "downgrade",
    "exceeded": false
  },
  "monthly": {
    "enabled": true,
    "limit": 200.0,
    "spent": 145.80,
    "percent": 72.9,
    "action": "block",
    "exceeded": false
  }
}
```

### Mettre à jour les plafonds

```bash
PUT /api/v1/budget/limits
Content-Type: application/json

{
  "daily": {
    "enabled": true,
    "limit": 15.0,
    "action": "warn"
  }
}
```

## Suivi au niveau du projet

Suivez les coûts par projet grâce aux liaisons de répertoire :

```bash
# Bind current directory to a profile
zen bind work-profile

# All requests from this directory are tagged with the project path
# View costs in Web UI under "By Project"
```

## Notifications webhook

Recevez une alerte quand un budget est dépassé :

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": ["budget_warning", "budget_exceeded"]
    }
  ]
}
```

Voir [Webhooks](./webhooks.md) pour la configuration complète.

## Bonnes pratiques

1. **Commencez par des avertissements** — utilisez d’abord `warn` pour comprendre vos usages
2. **Fixez des limites réalistes** — appuyez-vous sur l’historique de consommation
3. **Utilisez la rétrogradation en développement** — passez automatiquement à des modèles moins chers pendant les tests
4. **Réservez le blocage à la production** — n’utilisez `block` que pour des plafonds stricts
5. **Surveillez au quotidien** — consultez régulièrement le tableau de bord pour éviter les surprises
6. **Activez les webhooks** — soyez alerté en temps réel à l’approche des limites

## Dépannage

### L’usage n’est pas suivi

1. Vérifiez que `usage_tracking.enabled` vaut `true` dans la configuration
2. Vérifiez que le chemin de la base est accessible en écriture : `~/.zen/usage.db`
3. Redémarrez le daemon : `zen daemon restart`

### Coûts incorrects

1. Vérifiez que la tarification des modèles correspond aux tarifs actuels
2. Vérifiez la correspondance des noms de modèles (exacte ou par préfixe de famille)
3. Mettez à jour la tarification si les fournisseurs changent leurs tarifs

### Budget non appliqué

1. Vérifiez que la configuration du budget est activée
2. Vérifiez qu’une action est définie (`warn`, `downgrade` ou `block`)
3. Consultez les journaux du daemon à la recherche d’erreurs du vérificateur de budget

## Performance

- **Agrégation horaire** — les données brutes sont agrégées chaque heure pour alléger les requêtes
- **Requêtes indexées** — index sur fournisseur, modèle, projet et horodatage
- **Stockage efficace** — environ 1 Ko par requête, environ 30 Mo pour 30 000 requêtes
- **Tableau de bord rapide** — temps de requête inférieur à la seconde dans les cas courants
