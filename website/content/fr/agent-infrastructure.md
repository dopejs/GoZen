---
sidebar_position: 16
title: Infrastructure d’agents (BÊTA)
---

# Infrastructure d’agents (BÊTA)

:::warning Fonctionnalité BÊTA
L’infrastructure d’agents est en bêta. Elle est désactivée par défaut et doit être activée explicitement.
:::

Prise en charge intégrée des flux d’agents autonomes : gestion des sessions, coordination des fichiers, supervision en temps réel et garde-fous.

## Fonctionnalités

- **Runtime d’agents** — exécutez des tâches autonomes avec une gestion complète du cycle de vie
- **Observatoire** — supervision en temps réel des sessions et des activités
- **Garde-fous** — contrôles et contraintes de sécurité sur le comportement des agents
- **Coordinateur** — coordination par fichiers pour les flux multi-agents
- **File de tâches** — gérez les tâches avec priorités et dépendances
- **Gestion des sessions** — suivez les sessions d’agents sur plusieurs projets

## Architecture

```
Agent Client (Claude Code, Codex, etc.)
    ↓
Agent Runtime
    ↓
┌─────────────┬──────────────┬─────────────┐
│ Observatory │ Guardrails   │ Coordinator │
│ (Monitor)   │ (Safety)     │ (Sync)      │
└─────────────┴──────────────┴─────────────┘
    ↓
Task Queue → Provider API
```

## Configuration

### Activer l’infrastructure d’agents

```json
{
  "agent": {
    "enabled": true,
    "runtime": {
      "max_concurrent_tasks": 5,
      "task_timeout": "30m",
      "auto_cleanup": true
    },
    "observatory": {
      "enabled": true,
      "update_interval": "5s",
      "history_retention": "7d"
    },
    "guardrails": {
      "enabled": true,
      "max_file_operations": 100,
      "max_api_calls": 1000,
      "allowed_paths": ["/Users/john/projects"],
      "blocked_commands": ["rm -rf", "sudo"]
    },
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    }
  }
}
```

## Composants

### 1. Runtime d’agents

Gère le cycle de vie d’exécution des tâches d’agents.

**Fonctionnalités :**
- Planification et exécution des tâches
- Gestion des tâches concurrentes
- Gestion des délais d’expiration
- Nettoyage automatique
- Reprise sur erreur

**Configuration :**
```json
{
  "runtime": {
    "max_concurrent_tasks": 5,
    "task_timeout": "30m",
    "auto_cleanup": true,
    "retry_failed_tasks": true,
    "max_retries": 3
  }
}
```

**API :**
```bash
# Start agent task
POST /api/v1/agent/tasks
Content-Type: application/json

{
  "name": "code-review",
  "description": "Review pull request #123",
  "priority": 1,
  "config": {
    "model": "claude-opus-4",
    "max_tokens": 100000
  }
}

# Get task status
GET /api/v1/agent/tasks/{task_id}

# Cancel task
DELETE /api/v1/agent/tasks/{task_id}
```

### 2. Observatoire

Supervision en temps réel des activités des agents.

**Fonctionnalités :**
- Suivi des sessions
- Journalisation des activités
- Métriques de performance
- Mises à jour d’état
- Données historiques

**Configuration :**
```json
{
  "observatory": {
    "enabled": true,
    "update_interval": "5s",
    "history_retention": "7d",
    "metrics": {
      "track_tokens": true,
      "track_costs": true,
      "track_latency": true
    }
  }
}
```

**Métriques suivies :**
- Sessions actives
- Tâches en cours
- Consommation de jetons
- Appels d’API
- Opérations sur les fichiers
- Taux d’erreur
- Latence moyenne

**API :**
```bash
# Get all active sessions
GET /api/v1/agent/sessions

# Get session details
GET /api/v1/agent/sessions/{session_id}

# Get session metrics
GET /api/v1/agent/sessions/{session_id}/metrics
```

### 3. Garde-fous

Contrôles et contraintes de sécurité sur le comportement des agents.

**Fonctionnalités :**
- Limites d’opérations
- Restrictions de chemins
- Blocage de commandes
- Quotas de ressources
- Circuits d’approbation

**Configuration :**
```json
{
  "guardrails": {
    "enabled": true,
    "max_file_operations": 100,
    "max_api_calls": 1000,
    "max_tokens_per_session": 1000000,
    "allowed_paths": [
      "/Users/john/projects",
      "/tmp/agent-workspace"
    ],
    "blocked_paths": [
      "/etc",
      "/System",
      "~/.ssh"
    ],
    "blocked_commands": [
      "rm -rf /",
      "sudo",
      "chmod 777"
    ],
    "require_approval": {
      "file_delete": true,
      "system_commands": true,
      "network_requests": false
    }
  }
}
```

**Application :**
- Validation avant exécution
- Supervision en temps réel
- Blocage automatique
- Demandes d’approbation
- Journalisation d’audit

**API :**
```bash
# Get guardrail status
GET /api/v1/agent/guardrails

# Update guardrail rules
PUT /api/v1/agent/guardrails
Content-Type: application/json

{
  "max_file_operations": 200,
  "blocked_commands": ["rm -rf", "sudo", "dd"]
}
```

### 4. Coordinateur

Coordination par fichiers pour les flux multi-agents.

**Fonctionnalités :**
- Verrouillage de fichiers
- Détection des modifications
- Résolution des conflits
- Synchronisation d’état
- Notifications d’événements

**Configuration :**
```json
{
  "coordinator": {
    "enabled": true,
    "lock_timeout": "5m",
    "change_detection": true,
    "conflict_resolution": "last-write-wins",
    "notification_webhook": "https://hooks.slack.com/..."
  }
}
```

**Cas d’usage :**
- Plusieurs agents modifient les mêmes fichiers
- Empêcher les modifications concurrentes
- Détecter les changements de fichiers externes
- Coordonner des flux d’agents

**API :**
```bash
# Acquire file lock
POST /api/v1/agent/locks
Content-Type: application/json

{
  "path": "/path/to/file.go",
  "session_id": "sess_123",
  "timeout": "5m"
}

# Release file lock
DELETE /api/v1/agent/locks/{lock_id}

# Get file change events
GET /api/v1/agent/changes?since=2026-03-05T10:00:00Z
```

### 5. File de tâches

Gère les tâches d’agents avec priorités et dépendances.

**Fonctionnalités :**
- Ordonnancement par priorité
- Dépendances entre tâches
- Gestion de la file
- Suivi d’état
- Logique de réessai

**Configuration :**
```json
{
  "task_queue": {
    "enabled": true,
    "max_queue_size": 100,
    "priority_levels": 5,
    "enable_dependencies": true,
    "retry_policy": {
      "max_retries": 3,
      "backoff": "exponential"
    }
  }
}
```

**API :**
```bash
# Add task to queue
POST /api/v1/agent/queue
Content-Type: application/json

{
  "name": "run-tests",
  "priority": 2,
  "depends_on": ["build-project"],
  "config": {}
}

# Get queue status
GET /api/v1/agent/queue

# Remove task from queue
DELETE /api/v1/agent/queue/{task_id}
```

## Interface web

Le tableau de bord des agents est sur `http://localhost:19840/agent` :

### Onglet Sessions

- **Sessions actives** — les sessions d’agents en cours
- **Détail de session** — avancement des tâches, métriques, journaux
- **Contrôles de session** — mettre en pause, reprendre, annuler

### Onglet Tâches

- **File de tâches** — tâches en attente et en cours
- **Historique** — tâches terminées et échouées
- **Détail d’une tâche** — configuration, journaux, résultats

### Onglet Garde-fous

- **Limites d’opérations** — usage courant par rapport aux limites
- **Opérations bloquées** — tentatives récemment bloquées
- **File d’approbation** — opérations en attente d’approbation

### Onglet Métriques

- **Consommation de jetons** — par session et au total
- **Appels d’API** — nombre et fréquence des requêtes
- **Opérations sur les fichiers** — lectures, écritures, suppressions
- **Performance** — latence et débit

## Intégration avec Claude Code

GoZen détecte automatiquement les sessions Claude Code et leur fournit l’infrastructure d’agents :

```bash
# Start Claude Code with agent support
zen --agent

# Agent features are automatically enabled:
# - Session tracking
# - File coordination
# - Guardrails enforcement
# - Real-time monitoring
```

**Bénéfices :**
- Empêcher les modifications concurrentes de fichiers
- Suivre la consommation de jetons et les coûts
- Appliquer des contraintes de sécurité
- Superviser les activités des agents
- Coordonner les flux multi-agents

## Cas d’usage

### Développement multi-agents

Plusieurs agents travaillent sur la même base de code :

```json
{
  "agent": {
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    },
    "guardrails": {
      "max_file_operations": 200,
      "allowed_paths": ["/Users/john/project"]
    }
  }
}
```

### Tâches de longue durée

Superviser et piloter des tâches d’agents longues :

```json
{
  "agent": {
    "runtime": {
      "task_timeout": "2h",
      "auto_cleanup": false
    },
    "observatory": {
      "update_interval": "10s",
      "history_retention": "30d"
    }
  }
}
```

### Opérations critiques pour la sécurité

Appliquer des contrôles stricts :

```json
{
  "agent": {
    "guardrails": {
      "enabled": true,
      "max_file_operations": 50,
      "blocked_commands": ["rm", "sudo", "chmod"],
      "require_approval": {
        "file_delete": true,
        "system_commands": true,
        "network_requests": true
      }
    }
  }
}
```

## Bonnes pratiques

1. **Activez les garde-fous** — toujours, en production
2. **Fixez des limites adaptées** — configurez-les selon votre usage
3. **Supervisez activement** — consultez régulièrement l’observatoire
4. **Utilisez le verrouillage de fichiers** — activez le coordinateur pour les flux multi-agents
5. **Configurez les approbations** — exigez une approbation pour les opérations destructrices
6. **Relisez les journaux** — auditez régulièrement les activités des agents

## Limites

1. **Surcoût de performance** — la supervision et la coordination ajoutent de la latence
2. **Verrouillage de fichiers** — peut ralentir les scénarios multi-agents
3. **Consommation mémoire** — l’historique des sessions occupe de la mémoire
4. **Complexité** — demande de comprendre les flux d’agents
5. **Statut bêta** — les fonctionnalités peuvent changer

## Dépannage

### La session d’agent n’est pas suivie

1. Vérifiez que `agent.enabled` vaut `true`
2. Vérifiez que l’observatoire est activé
3. Vérifiez que le client d’agent est pris en charge (Claude Code, Codex)
4. Consultez les journaux du daemon

### Problèmes de verrouillage de fichiers

1. Vérifiez que le coordinateur est activé
2. Vérifiez que le délai de verrou est approprié
3. Consultez les verrous actifs : `GET /api/v1/agent/locks`
4. Libérez manuellement les verrous bloqués si nécessaire

### Les garde-fous ne s’appliquent pas

1. Vérifiez qu’ils sont activés
2. Vérifiez la configuration des règles
3. Consultez le journal des opérations bloquées
4. Assurez-vous que le client d’agent respecte les garde-fous

### Consommation mémoire élevée

1. Réduisez la durée de conservation de l’historique
2. Espacez l’intervalle de mise à jour
3. Limitez le nombre de tâches concurrentes
4. Activez le nettoyage automatique

## Sécurité

1. **Restrictions de chemins** — configurez toujours les chemins autorisés et interdits
2. **Blocage de commandes** — bloquez les commandes dangereuses
3. **Circuits d’approbation** — exigez une approbation pour les opérations sensibles
4. **Journalisation d’audit** — activez une journalisation complète
5. **Limites de ressources** — fixez des limites d’opérations adaptées

## Évolutions prévues

- Protocoles de collaboration entre agents
- Stratégies avancées de résolution de conflits
- Apprentissage automatique pour détecter les anomalies
- Intégration avec des outils de supervision externes
- Analyse du comportement des agents
- Génération automatique de politiques de sécurité
