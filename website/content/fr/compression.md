---
sidebar_position: 14
title: Compression du contexte (BÊTA)
---

# Compression du contexte (BÊTA)

:::warning Fonctionnalité BÊTA
La compression du contexte est en bêta. Elle est désactivée par défaut et doit être activée explicitement.
:::

Compresse automatiquement le contexte de conversation quand le nombre de jetons dépasse un seuil, ce qui réduit les coûts tout en préservant la qualité de l’échange.

## Fonctionnalités

- **Compression automatique** — déclenchée au dépassement du seuil de jetons
- **Résumé intelligent** — un modèle bon marché (claude-3-haiku) résume les messages anciens
- **Préservation des messages récents** — les derniers messages restent intacts pour la continuité
- **Estimation des jetons** — comptage précis avant les appels d’API
- **Suivi statistique** — mesurez l’efficacité de la compression
- **Fonctionnement transparent** — compatible avec tous les clients d’IA

## Fonctionnement

1. **Estimation des jetons** — compter les jetons de l’historique
2. **Vérification du seuil** — comparer au seuil configuré (par défaut : 50 000)
3. **Sélection des messages** — repérer les messages anciens à compresser
4. **Résumé** — produire un résumé concis avec un modèle bon marché
5. **Remplacement du contexte** — remplacer les anciens messages par le résumé
6. **Transmission de la requête** — envoyer le contexte compressé au modèle cible

## Configuration

### Activer la compression

```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 50000,
    "target_tokens": 20000,
    "summarizer_model": "claude-3-haiku-20240307",
    "preserve_recent_messages": 5,
    "tokens_per_char": 0.25
  }
}
```

**Options :**

| Option | Valeur par défaut | Description |
|--------|---------|-------------|
| `enabled` | `false` | Active la compression du contexte |
| `threshold_tokens` | `50000` | Déclenche la compression au-delà de cette valeur |
| `target_tokens` | `20000` | Nombre de jetons visé après compression |
| `summarizer_model` | `claude-3-haiku-20240307` | Modèle utilisé pour le résumé |
| `preserve_recent_messages` | `5` | Nombre de messages récents laissés intacts |
| `tokens_per_char` | `0.25` | Ratio d’estimation pour le comptage des jetons |

### Configuration par profil

Activez la compression pour certains profils :

```json
{
  "profiles": {
    "long-context": {
      "providers": ["anthropic"],
      "compression": {
        "enabled": true,
        "threshold_tokens": 100000,
        "target_tokens": 40000
      }
    },
    "short-context": {
      "providers": ["openai"],
      "compression": {
        "enabled": false
      }
    }
  }
}
```

## Estimation des jetons

GoZen estime rapidement les jetons à partir du nombre de caractères :

```
estimated_tokens = character_count * tokens_per_char
```

**Ratio par défaut :** 0,25 jeton par caractère (1 jeton ≈ 4 caractères)

**Précision :** ±10 % pour l’anglais, variable pour d’autres langues

Pour un comptage exact, GoZen utilise la bibliothèque `tiktoken-go` quand elle est disponible.

## Stratégie de compression

### Sélection des messages

1. **Messages système** — toujours préservés
2. **Messages récents** — les N derniers messages sont préservés (par défaut : 5)
3. **Messages plus anciens** — candidats à la compression

### Invite de résumé

```
Summarize the following conversation history concisely while preserving key information, decisions, and context:

[older messages]

Provide a brief summary that captures the essential points.
```

### Résultat

```
Original: 45,000 tokens (30 messages)
After compression: 22,000 tokens (summary + 5 recent messages)
Savings: 23,000 tokens (51%)
```

## Interface web

Les réglages de compression se trouvent sur `http://localhost:19840/settings` :

1. Ouvrez l’onglet « Compression » (marqué d’un badge BÊTA)
2. Activez « Enable Compression »
3. Ajustez le seuil et le nombre de jetons visé
4. Choisissez le modèle de résumé
5. Fixez le nombre de messages récents à préserver
6. Cliquez sur « Save »

### Tableau de bord statistique

Consultez les statistiques de compression :

- **Compressions totales** — nombre de déclenchements
- **Jetons économisés** — total des jetons économisés
- **Économie moyenne** — réduction moyenne par compression
- **Taux de compression** — pourcentage de requêtes ayant déclenché une compression

## Points d’API

### Obtenir les statistiques de compression

```bash
GET /api/v1/compression/stats
```

Réponse :
```json
{
  "enabled": true,
  "total_compressions": 42,
  "tokens_saved": 1250000,
  "average_savings": 29761,
  "compression_rate": 0.15,
  "last_compression": "2026-03-05T10:30:00Z"
}
```

### Mettre à jour les réglages de compression

```bash
PUT /api/v1/compression/settings
Content-Type: application/json

{
  "enabled": true,
  "threshold_tokens": 60000,
  "target_tokens": 25000
}
```

### Réinitialiser les statistiques

```bash
POST /api/v1/compression/stats/reset
```

## Cas d’usage

### Longues sessions de code

**Scénario :** plusieurs heures de code avec Claude Code

**Configuration :**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 80000,
    "target_tokens": 30000,
    "preserve_recent_messages": 10
  }
}
```

**Bénéfice :** garder la continuité de la conversation sans atteindre les limites de contexte

### Traitement par lots

**Scénario :** traitement de nombreux documents avec l’IA

**Configuration :**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 40000,
    "target_tokens": 15000,
    "preserve_recent_messages": 3
  }
}
```

**Bénéfice :** réduire les coûts sur de grands ensembles de documents

### Recherche et analyse

**Scénario :** longues sessions de recherche sur plusieurs sujets

**Configuration :**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 100000,
    "target_tokens": 40000,
    "preserve_recent_messages": 8
  }
}
```

**Bénéfice :** rester concentré sur les sujets récents tout en conservant le contexte antérieur

## Bonnes pratiques

1. **Commencez par les valeurs par défaut** — elles conviennent à la plupart des usages
2. **Surveillez les statistiques** — vérifiez régulièrement le taux et les économies
3. **Ajustez le seuil** — augmentez-le pour les modèles à long contexte (Claude Opus), diminuez-le sinon
4. **Préservez assez de messages** — gardez 5 à 10 messages récents pour la continuité
5. **Utilisez un modèle de résumé bon marché** — Haiku est rapide et économique
6. **Testez avant la production** — vérifiez la qualité sur votre cas d’usage précis

## Limites

1. **Perte de qualité** — le résumé peut perdre des détails fins
2. **Latence accrue** — un appel d’API supplémentaire est ajouté
3. **Arbitrage de coût** — coût du résumé contre jetons économisés
4. **Prise en charge des langues** — meilleure en anglais, variable ailleurs
5. **Fenêtre de contexte** — impossible de dépasser la fenêtre maximale du modèle

## Dépannage

### La compression ne se déclenche pas

1. Vérifiez que `compression.enabled` vaut `true`
2. Vérifiez que le nombre de jetons dépasse le seuil
3. Assurez-vous que la conversation contient assez de messages à compresser
4. Consultez les journaux du daemon à la recherche d’erreurs de compression

### Qualité de résumé médiocre

1. Essayez un autre modèle de résumé (par exemple claude-3-sonnet)
2. Augmentez `preserve_recent_messages` pour garder plus de contexte
3. Ajustez `target_tokens` pour autoriser des résumés plus longs
4. Vérifiez que le modèle de résumé est disponible et fonctionne

### Latence accrue

1. La compression ajoute un appel d’API (le résumé)
2. Prenez un modèle de résumé plus rapide (Haiku est le plus rapide)
3. Augmentez le seuil pour compresser moins souvent
4. Envisagez de la désactiver pour les applications sensibles à la latence

### Coûts inattendus

1. Suivez le coût des résumés dans le tableau de bord d’usage
2. Comparez les économies au coût du résumé
3. Augmentez le seuil pour compresser moins souvent
4. Utilisez le modèle le moins cher disponible pour le résumé

## Impact sur les performances

- **Estimation des jetons** — environ 1 ms par requête (négligeable)
- **Résumé** — 1 à 3 secondes (selon le modèle et le nombre de messages)
- **Surcoût mémoire** — minime (environ 1 Ko par compression)
- **Économies** — typiquement 30 à 50 % de jetons en moins

## Configuration avancée

### Invite de résumé personnalisée

```json
{
  "compression": {
    "enabled": true,
    "custom_prompt": "Create a technical summary of the following conversation, focusing on code changes, decisions, and action items:\n\n{messages}\n\nSummary:"
  }
}
```

### Compression conditionnelle

N’activez la compression que pour certains scénarios :

```json
{
  "profiles": {
    "default": {
      "scenarios": {
        "longContext": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": true,
            "threshold_tokens": 100000
          }
        },
        "default": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": false
          }
        }
      }
    }
  }
}
```

### Compression en plusieurs étapes

Compressez plusieurs fois pour les conversations très longues :

```json
{
  "compression": {
    "enabled": true,
    "stages": [
      {
        "threshold_tokens": 50000,
        "target_tokens": 30000
      },
      {
        "threshold_tokens": 80000,
        "target_tokens": 40000
      }
    ]
  }
}
```

## Évolutions prévues

- Sélection intelligente des messages par similarité sémantique
- Résumé multi-modèles pour comparer la qualité
- Métriques de qualité et retours sur la compression
- Stratégies de compression personnalisées par cas d’usage
- Intégration avec le RAG pour stocker le contexte à l’extérieur
