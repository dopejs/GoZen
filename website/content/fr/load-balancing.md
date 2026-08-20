---
title: Répartition de charge
---

# Répartition de charge

GoZen propose plusieurs stratégies de sélection des fournisseurs au-delà de la simple bascule. Vous choisissez une stratégie par profil, la combinez avec les contrôles de santé et orientez le trafic selon la disponibilité, la latence ou le coût.

## Stratégies disponibles

### Bascule

Essaie les fournisseurs dans l’ordre jusqu’à ce que l’un réponde. C’est la stratégie par défaut, adaptée aux configurations principal/secours.

```json
{
  "profiles": {
    "default": {
      "providers": ["primary", "backup"],
      "strategy": "failover"
    }
  }
}
```

### Tourniquet

Répartit les requêtes équitablement entre plusieurs fournisseurs équivalents.

```json
{
  "profiles": {
    "balanced": {
      "providers": ["provider-a", "provider-b", "provider-c"],
      "strategy": "round-robin"
    }
  }
}
```

### Latence minimale

Privilégie le fournisseur au temps de réponse récent le plus faible.

```json
{
  "profiles": {
    "fast": {
      "providers": ["us-east", "us-west", "eu"],
      "strategy": "least-latency"
    }
  }
}
```

### Coût minimal

Privilégie le fournisseur le moins cher pour le modèle demandé.

```json
{
  "profiles": {
    "budget": {
      "providers": ["cheap-provider", "premium-provider"],
      "strategy": "least-cost"
    }
  }
}
```

## Routage sensible à la santé

Toutes les stratégies fonctionnent avec la supervision de santé. Quand `health_aware` est activé, les fournisseurs en mauvaise santé sont ignorés automatiquement jusqu’à leur rétablissement.

```json
{
  "profiles": {
    "production": {
      "providers": ["primary", "secondary", "tertiary"],
      "strategy": "least-latency",
      "health_aware": true
    }
  }
}
```

## Choisir une stratégie

- Utilisez `failover` quand la fiabilité prime.
- Utilisez `round-robin` quand les fournisseurs sont interchangeables.
- Utilisez `least-latency` pour les charges interactives ou sensibles au temps.
- Utilisez `least-cost` quand le budget compte plus que la vitesse brute.

## Documents liés

- [Profils](/docs/profiles) explique comment les groupes de fournisseurs sont définis.
- [Routage](/docs/routing) traite de la sélection par scénario.
- [Supervision de la santé](/docs/health-monitoring) explique l’effet des contrôles de santé sur le routage.
