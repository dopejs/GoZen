---
sidebar_position: 12
title: Supervision de la santé
---

# Supervision de la santé et répartition de charge

Surveillez la santé des fournisseurs en temps réel et routez automatiquement les requêtes vers le meilleur fournisseur disponible.

## Fonctionnalités

- **Contrôles de santé en temps réel** — supervision périodique, à intervalle configurable
- **Suivi du taux de succès** — santé calculée à partir du taux de requêtes réussies
- **Supervision de la latence** — temps de réponse moyen par fournisseur
- **Plusieurs stratégies** — bascule, tourniquet, latence minimale, coût minimal
- **Bascule automatique** — passage aux fournisseurs de secours quand le principal est en mauvaise santé
- **Tableau de bord de santé** — indicateurs visuels dans l’interface web

## Configuration

### Activer la supervision de santé

```json
{
  "health_check": {
    "enabled": true,
    "interval": "5m",
    "timeout": "10s",
    "endpoint": "/v1/messages",
    "method": "POST"
  }
}
```

**Options :**
- `interval` — fréquence des contrôles de santé (par défaut : 5 minutes)
- `timeout` — délai d’expiration des requêtes de contrôle (par défaut : 10 secondes)
- `endpoint` — point d’API à tester (par défaut : `/v1/messages`)
- `method` — méthode HTTP du contrôle de santé (par défaut : `POST`)

### Configurer la répartition de charge

```json
{
  "load_balancing": {
    "strategy": "least-latency",
    "health_aware": true,
    "cache_ttl": "30s"
  }
}
```

## Stratégies de répartition de charge

### 1. Bascule (par défaut)

Utilise les fournisseurs dans l’ordre et passe au suivant en cas d’échec.

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-primary", "anthropic-backup", "openai"],
      "load_balancing": {
        "strategy": "failover"
      }
    }
  }
}
```

**Comportement :**
1. Essayer `anthropic-primary`
2. En cas d’échec, essayer `anthropic-backup`
3. En cas d’échec, essayer `openai`
4. Si tous échouent, renvoyer une erreur

**Idéal pour :** les charges de production avec une hiérarchie claire principal/secours

### 2. Tourniquet

Répartit les requêtes équitablement entre tous les fournisseurs en bonne santé.

```json
{
  "load_balancing": {
    "strategy": "round-robin"
  }
}
```

**Comportement :**
- Requête 1 → fournisseur A
- Requête 2 → fournisseur B
- Requête 3 → fournisseur C
- Requête 4 → fournisseur A (le cycle recommence)

**Idéal pour :** répartir la charge sur plusieurs comptes afin d’éviter les limites de débit

### 3. Latence minimale

Route vers le fournisseur dont la latence moyenne est la plus faible.

```json
{
  "load_balancing": {
    "strategy": "least-latency"
  }
}
```

**Comportement :**
- Suit le temps de réponse moyen par fournisseur
- Route vers le plus rapide
- Met à jour les métriques toutes les 30 secondes (configurable via `cache_ttl`)

**Idéal pour :** les applications sensibles à la latence et les interactions en temps réel

### 4. Coût minimal

Route vers le fournisseur le moins cher pour le modèle demandé.

```json
{
  "load_balancing": {
    "strategy": "least-cost"
  }
}
```

**Comportement :**
- Compare les tarifs des fournisseurs
- Route vers l’option la moins chère
- Tient compte du coût des jetons en entrée et en sortie

**Idéal pour :** l’optimisation des coûts et les traitements par lots

## État de santé

Les fournisseurs sont classés en quatre états de santé :

| État | Taux de succès | Comportement |
|--------|--------------|----------|
| **Sain** | ≥ 95 % | Priorité normale |
| **Dégradé** | 70 à 95 % | Priorité réduite, reste utilisable |
| **Défaillant** | < 70 % | Ignoré, sauf en l’absence de fournisseur sain |
| **Inconnu** | Aucune donnée | Considéré comme sain au départ |

### Routage sensible à la santé

Quand `health_aware: true` (valeur par défaut) :
- Les fournisseurs sains sont prioritaires
- Les fournisseurs dégradés servent de repli
- Les fournisseurs défaillants sont ignorés, sauf si tous les autres échouent

## Tableau de bord de l’interface web

Accédez au tableau de bord de santé sur `http://localhost:19840/health` :

### État des fournisseurs

- **Indicateur d’état** — vert (sain), jaune (dégradé), rouge (défaillant)
- **Taux de succès** — pourcentage de requêtes réussies
- **Latence moyenne** — temps de réponse moyen en millisecondes
- **Dernier contrôle** — horodatage du contrôle de santé le plus récent
- **Nombre d’erreurs** — nombre d’échecs récents

### Chronologie des métriques

- **Graphique de latence** — évolution des temps de réponse
- **Graphique du taux de succès** — évolution de la santé
- **Volume de requêtes** — requêtes par fournisseur

## Points d’API

### Obtenir la santé d’un fournisseur

```bash
GET /api/v1/health/providers
```

Réponse :
```json
{
  "providers": [
    {
      "name": "anthropic-primary",
      "status": "healthy",
      "success_rate": 98.5,
      "avg_latency_ms": 1250,
      "last_check": "2026-03-05T10:30:00Z",
      "error_count": 2,
      "total_requests": 150
    },
    {
      "name": "openai-backup",
      "status": "degraded",
      "success_rate": 85.0,
      "avg_latency_ms": 2100,
      "last_check": "2026-03-05T10:29:00Z",
      "error_count": 15,
      "total_requests": 100
    }
  ]
}
```

### Obtenir les métriques d’un fournisseur

```bash
GET /api/v1/health/providers/{name}/metrics?period=1h
```

Réponse :
```json
{
  "provider": "anthropic-primary",
  "period": "1h",
  "metrics": [
    {
      "timestamp": "2026-03-05T10:00:00Z",
      "latency_ms": 1200,
      "success_rate": 99.0,
      "requests": 25
    },
    {
      "timestamp": "2026-03-05T10:05:00Z",
      "latency_ms": 1300,
      "success_rate": 98.0,
      "requests": 28
    }
  ]
}
```

### Déclencher un contrôle de santé manuel

```bash
POST /api/v1/health/check
Content-Type: application/json

{
  "provider": "anthropic-primary"
}
```

## Notifications webhook

Recevez une alerte quand l’état d’un fournisseur change :

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": ["provider_down", "provider_up", "failover"]
    }
  ]
}
```

**Types d’événements :**
- `provider_down` — le fournisseur devient défaillant
- `provider_up` — le fournisseur redevient sain
- `failover` — la requête a basculé vers un fournisseur de secours

## Routage par scénario

Combinez supervision de santé et routage par scénario pour une distribution intelligente des requêtes :

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-primary", "anthropic-backup"],
      "scenarios": {
        "thinking": {
          "providers": ["anthropic-thinking"],
          "load_balancing": {
            "strategy": "least-latency"
          }
        },
        "image": {
          "providers": ["anthropic-vision", "openai-vision"],
          "load_balancing": {
            "strategy": "failover"
          }
        }
      }
    }
  }
}
```

Voir [Routage par scénario](./routing.md) pour les détails.

## Bonnes pratiques

1. **Choisissez le bon intervalle** — 5 minutes convient dans la plupart des cas, 1 minute pour les systèmes critiques
2. **Activez le routage sensible à la santé** — toujours, en production
3. **Surveillez les fournisseurs dégradés** — enquêtez dès que le taux de succès passe sous 95 %
4. **Combinez les stratégies** — bascule pour principal/secours, tourniquet pour répartir la charge
5. **Activez les webhooks** — soyez prévenu immédiatement quand un fournisseur tombe
6. **Consultez le tableau de bord** — passez en revue les tendances pour repérer les schémas

## Dépannage

### Les contrôles de santé échouent

1. Vérifiez la validité des clés d’API des fournisseurs
2. Vérifiez la connectivité réseau vers les points d’accès des fournisseurs
3. Augmentez le délai si les fournisseurs sont lents : `"timeout": "30s"`
4. Consultez les journaux du daemon pour les messages d’erreur précis

### Métriques de latence incorrectes

1. La latence inclut le temps réseau et le temps de traitement de l’API
2. Vérifiez si un proxy ou un VPN ajoute de la surcharge
3. Les métriques sont mises en cache 30 secondes par défaut (configurable via `cache_ttl`)

### La bascule ne fonctionne pas

1. Vérifiez `health_aware: true` dans la configuration de répartition de charge
2. Vérifiez que des fournisseurs de secours figurent dans le profil
3. Assurez-vous que les contrôles de santé sont activés et s’exécutent
4. Passez en revue les événements de bascule dans l’interface web ou les journaux

### Un fournisseur reste bloqué en état défaillant

1. Déclenchez un contrôle de santé manuel via l’API
2. Vérifiez si le fournisseur est réellement hors service (testez avec curl)
3. Redémarrez le daemon pour réinitialiser l’état : `zen daemon restart`
4. Cherchez la cause racine dans les journaux d’erreurs

## Impact sur les performances

- **Contrôles de santé** — surcharge minime, exécutés dans une goroutine en arrière-plan
- **Mise en cache des métriques** — un TTL de 30 secondes réduit les requêtes en base
- **Opérations atomiques** — compteurs sûrs pour les requêtes concurrentes
- **Sans blocage** — les contrôles de santé ne bloquent pas le traitement des requêtes

## Configuration avancée

### Charge utile personnalisée pour le contrôle de santé

```json
{
  "health_check": {
    "enabled": true,
    "custom_payload": {
      "model": "claude-3-haiku-20240307",
      "max_tokens": 10,
      "messages": [
        {
          "role": "user",
          "content": "ping"
        }
      ]
    }
  }
}
```

### Réglages de santé par fournisseur

```json
{
  "providers": {
    "anthropic-primary": {
      "health_check": {
        "interval": "1m",
        "timeout": "5s"
      }
    },
    "openai-backup": {
      "health_check": {
        "interval": "5m",
        "timeout": "10s"
      }
    }
  }
}
```
