---
title: Agents
---

# Agents

GoZen peut servir de couche d’exploitation pour des agents de code comme Claude Code, Codex et d’autres assistants pilotés en ligne de commande. Il vous aide à coordonner leur travail, à surveiller les sessions et à appliquer des garde-fous à l’exécution, sans changer votre façon de travailler.

## Ce que GoZen apporte

- **Coordination** : réduit les conflits quand plusieurs agents travaillent sur le même projet.
- **Observabilité** : suivez sessions, coûts, erreurs et activité au même endroit.
- **Garde-fous** : posez des limites de dépense, de débit de requêtes et d’actions sensibles.
- **Routage des tâches** : envoyez des travaux différents vers des fournisseurs ou des profils différents.

## Exemple de configuration

```json
{
  "agent": {
    "enabled": true,
    "coordinator": {
      "enabled": true,
      "lock_timeout_sec": 300,
      "inject_warnings": true
    },
    "observatory": {
      "enabled": true,
      "stuck_threshold": 5,
      "idle_timeout_min": 30
    },
    "guardrails": {
      "enabled": true,
      "session_spending_cap": 5.0,
      "request_rate_limit": 30
    }
  }
}
```

## Usages courants

### Coordination de plusieurs agents

Quand plusieurs agents travaillent dans le même dépôt, GoZen suit l’activité sur les fichiers, remonte des avertissements et rend les collisions plus faciles à éviter.

### Surveillance des sessions

Utilisez le tableau de bord et les API pour inspecter les sessions actives, la consommation de jetons, le nombre d’erreurs et la durée d’exécution.

### Application des garde-fous

Les garde-fous peuvent suspendre une session qui s’emballe, signaler les opérations risquées et ralentir les boucles de réessai avant qu’elles ne coûtent cher.

## Documents liés

- [Infrastructure d’agents](/docs/agent-infrastructure) détaille l’architecture plus récente : runtime, observatoire, coordinateur et garde-fous.
- [Passerelle de bots](/docs/bot) explique comment piloter les sessions en cours depuis Telegram, Slack, Discord et d’autres messageries.
- [Suivi de l’usage](/docs/usage-tracking) et [Supervision de la santé](/docs/health-monitoring) présentent les métriques qui alimentent l’exploitation des agents.
