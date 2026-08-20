---
sidebar_position: 15
title: Chaîne de middlewares (BÊTA)
---

# Chaîne de middlewares (BÊTA)

:::warning Fonctionnalité BÊTA
La chaîne de middlewares est en bêta. Elle est désactivée par défaut et doit être activée explicitement.
:::

Étendez GoZen avec des middlewares enfichables : transformation des requêtes et réponses, journalisation, limitation de débit et traitements sur mesure.

## Fonctionnalités

- **Architecture enfichable** — ajoutez votre logique sans modifier le cœur du produit
- **Exécution par priorité** — maîtrisez l’ordre d’exécution des middlewares
- **Points d’ancrage requête/réponse** — agissez avant l’envoi et après la réception
- **Middlewares intégrés** — injection de contexte, journalisation, limitation de débit, compression
- **Chargeur de greffons** — chargez des middlewares depuis un fichier local ou une URL distante
- **Gestion des erreurs** — traitement gracieux, avec comportement de repli

## Architecture

```
Client Request
    ↓
[Middleware 1: Priority 100]
    ↓
[Middleware 2: Priority 200]
    ↓
[Middleware 3: Priority 300]
    ↓
Provider API
    ↓
[Middleware 3: Response]
    ↓
[Middleware 2: Response]
    ↓
[Middleware 1: Response]
    ↓
Client Response
```

## Configuration

### Activer la chaîne de middlewares

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "context-injection",
        "enabled": true,
        "priority": 100,
        "config": {}
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info"
        }
      }
    ]
  }
}
```

**Options :**

| Option | Description |
|--------|-------------|
| `enabled` | Active la chaîne de middlewares |
| `pipeline` | Tableau des configurations de middlewares |
| `name` | Identifiant du middleware |
| `priority` | Ordre d’exécution (plus petit = plus tôt) |
| `config` | Configuration propre au middleware |

## Middlewares intégrés

### 1. Injection de contexte

Injecte du contexte personnalisé dans les requêtes.

```json
{
  "name": "context-injection",
  "enabled": true,
  "priority": 100,
  "config": {
    "system_prompt": "You are a helpful coding assistant.",
    "metadata": {
      "session_id": "sess_123",
      "user_id": "user_456"
    }
  }
}
```

**Cas d’usage :**
- Ajouter des invites système
- Injecter des métadonnées de session
- Ajouter du contexte utilisateur

### 2. Journalisation des requêtes

Journalise toutes les requêtes et réponses.

```json
{
  "name": "request-logger",
  "enabled": true,
  "priority": 200,
  "config": {
    "log_level": "info",
    "log_body": false,
    "log_headers": true
  }
}
```

**Cas d’usage :**
- Débogage
- Pistes d’audit
- Suivi des performances

### 3. Limiteur de débit

Limite le débit des requêtes, par fournisseur ou globalement.

```json
{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60,
    "burst": 10,
    "per_provider": true
  }
}
```

**Cas d’usage :**
- Éviter les erreurs de limite de débit
- Maîtriser la consommation d’API
- Se protéger des abus

### 4. Compression (BÊTA)

Compresse le contexte quand le nombre de jetons dépasse le seuil.

```json
{
  "name": "compression",
  "enabled": true,
  "priority": 400,
  "config": {
    "threshold_tokens": 50000,
    "target_tokens": 20000
  }
}
```

Voir [Compression du contexte](./compression.md) pour les détails.

### 5. Mémoire de session (BÊTA)

Conserve la mémoire de la conversation d’une session à l’autre.

```json
{
  "name": "session-memory",
  "enabled": true,
  "priority": 150,
  "config": {
    "max_memories": 100,
    "ttl_hours": 24,
    "storage": "sqlite"
  }
}
```

**Cas d’usage :**
- Retenir les préférences de l’utilisateur
- Suivre l’historique des échanges
- Maintenir le contexte entre les sessions

### 6. Orchestration (BÊTA)

Envoie les requêtes à plusieurs fournisseurs et agrège les réponses.

```json
{
  "name": "orchestration",
  "enabled": true,
  "priority": 500,
  "config": {
    "strategy": "parallel",
    "providers": ["anthropic", "openai"],
    "consensus": "longest"
  }
}
```

**Cas d’usage :**
- Comparer les sorties de plusieurs modèles
- Redondance pour les requêtes critiques
- Améliorer la qualité par consensus

## Middleware personnalisé

### Interface d’un middleware

```go
type Middleware interface {
    Name() string
    Priority() int
    ProcessRequest(ctx *RequestContext) error
    ProcessResponse(ctx *ResponseContext) error
}

type RequestContext struct {
    Provider  string
    Model     string
    Messages  []Message
    Metadata  map[string]interface{}
}

type ResponseContext struct {
    Provider  string
    Model     string
    Response  *APIResponse
    Latency   time.Duration
    Metadata  map[string]interface{}
}
```

### Exemple : injection d’un en-tête personnalisé

```go
package main

import (
    "github.com/dopejs/gozen/internal/middleware"
)

type CustomHeaderMiddleware struct {
    headers map[string]string
}

func (m *CustomHeaderMiddleware) Name() string {
    return "custom-headers"
}

func (m *CustomHeaderMiddleware) Priority() int {
    return 250
}

func (m *CustomHeaderMiddleware) ProcessRequest(ctx *middleware.RequestContext) error {
    for k, v := range m.headers {
        ctx.Metadata[k] = v
    }
    return nil
}

func (m *CustomHeaderMiddleware) ProcessResponse(ctx *middleware.ResponseContext) error {
    // No response processing needed
    return nil
}

func init() {
    middleware.Register("custom-headers", func(config map[string]interface{}) middleware.Middleware {
        return &CustomHeaderMiddleware{
            headers: config["headers"].(map[string]string),
        }
    })
}
```

### Charger un middleware personnalisé

#### Greffon local

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "local",
        "path": "/path/to/custom-middleware.so",
        "config": {
          "headers": {
            "X-Custom-Header": "value"
          }
        }
      }
    ]
  }
}
```

#### Greffon distant

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "remote",
        "url": "https://example.com/middleware/custom-headers.so",
        "checksum": "sha256:abc123...",
        "config": {}
      }
    ]
  }
}
```

## Interface web

Les réglages des middlewares se trouvent sur `http://localhost:19840/settings` :

1. Ouvrez l’onglet « Middleware » (marqué d’un badge BÊTA)
2. Activez « Enable Middleware Pipeline »
3. Ajoutez ou retirez des middlewares de la chaîne
4. Ajustez la priorité et la configuration
5. Activez ou désactivez chaque middleware
6. Cliquez sur « Save »

## Points d’API

### Lister les middlewares

```bash
GET /api/v1/middleware
```

Réponse :
```json
{
  "enabled": true,
  "pipeline": [
    {
      "name": "context-injection",
      "enabled": true,
      "priority": 100,
      "type": "builtin"
    },
    {
      "name": "request-logger",
      "enabled": true,
      "priority": 200,
      "type": "builtin"
    }
  ]
}
```

### Ajouter un middleware

```bash
POST /api/v1/middleware
Content-Type: application/json

{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60
  }
}
```

### Mettre à jour un middleware

```bash
PUT /api/v1/middleware/{name}
Content-Type: application/json

{
  "enabled": false
}
```

### Retirer un middleware

```bash
DELETE /api/v1/middleware/{name}
```

### Recharger la chaîne

```bash
POST /api/v1/middleware/reload
```

## Cas d’usage

### Environnement de développement

Ajoutez la journalisation de débogage et l’inspection des requêtes :

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 100,
        "config": {
          "log_level": "debug",
          "log_body": true
        }
      }
    ]
  }
}
```

### Environnement de production

Ajoutez la limitation de débit et la supervision :

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "rate-limiter",
        "enabled": true,
        "priority": 100,
        "config": {
          "requests_per_minute": 100,
          "burst": 20
        }
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info",
          "log_body": false
        }
      }
    ]
  }
}
```

### Comparaison entre fournisseurs

Utilisez l’orchestration pour comparer les sorties :

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "orchestration",
        "enabled": true,
        "priority": 500,
        "config": {
          "strategy": "parallel",
          "providers": ["anthropic", "openai", "google"],
          "consensus": "longest"
        }
      }
    ]
  }
}
```

## Bonnes pratiques

1. **Choisissez bien les priorités** — les plus petits nombres s’exécutent en premier
2. **Gardez chaque middleware focalisé** — une seule responsabilité, bien remplie
3. **Gérez les erreurs proprement** — une erreur ne doit pas casser la chaîne
4. **Testez soigneusement** — vérifiez le comportement avant la production
5. **Surveillez les performances** — mesurez le surcoût des middlewares
6. **Documentez la configuration** — décrivez clairement les options

## Limites

1. **Surcoût de performance** — chaque middleware ajoute de la latence
2. **Complexité** — trop de middlewares rendent le débogage difficile
3. **Sécurité des greffons** — les greffons distants demandent confiance et vérification
4. **Propagation des erreurs** — une erreur de middleware peut toucher toutes les requêtes
5. **Complexité de configuration** — les chaînes complexes sont plus dures à maintenir

## Dépannage

### Le middleware ne s’exécute pas

1. Vérifiez que `middleware.enabled` vaut `true`
2. Vérifiez que le middleware est activé dans la chaîne
3. Vérifiez que la priorité est correcte
4. Consultez les journaux du daemon à la recherche d’erreurs de middleware

### Comportement inattendu

1. Vérifiez l’ordre d’exécution (la priorité)
2. Vérifiez la configuration
3. Testez le middleware isolément
4. Passez en revue ses journaux

### Problèmes de performance

1. Identifiez le middleware lent (voir les journaux)
2. Réduisez le nombre de middlewares
3. Optimisez l’implémentation
4. Envisagez de désactiver les middlewares non essentiels

### Échec du chargement d’un greffon

1. Vérifiez le chemin du greffon
2. Vérifiez qu’il est compilé pour la bonne architecture
3. Vérifiez la somme de contrôle (pour les greffons distants)
4. Consultez les journaux du greffon

## Sécurité

1. **Validez les greffons** — ne chargez que des greffons de confiance
2. **Vérifiez les sommes de contrôle** — toujours, pour les greffons distants
3. **Isolez les greffons** — envisagez de les exécuter dans un environnement isolé
4. **Auditez les middlewares** — relisez le code avant tout déploiement
5. **Surveillez le comportement** — guettez tout comportement inattendu

## Évolutions prévues

- Prise en charge de greffons WebAssembly pour la portabilité
- Place de marché pour partager les greffons de la communauté
- Éditeur visuel de chaîne dans l’interface web
- Profilage des performances des middlewares
- Rechargement à chaud lors des mises à jour de greffons
- Cadre de test pour les middlewares
