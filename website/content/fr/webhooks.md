---
sidebar_position: 13
title: Webhooks
---

# Webhooks

Recevez des notifications en temps réel — alertes de budget, changements d’état des fournisseurs, résumés quotidiens — via Slack, Discord ou un webhook personnalisé.

## Fonctionnalités

- **Plusieurs formats** — Slack, Discord ou JSON générique
- **Filtrage des événements** — abonnez-vous à certains types d’événements
- **En-têtes personnalisés** — ajoutez une authentification ou vos propres en-têtes
- **Envoi asynchrone** — la livraison ne bloque pas les requêtes
- **Mise en forme automatique** — messages enrichis, avec émojis et couleurs
- **Fonction de test** — vérifiez la configuration avant de l’activer

## Configuration

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": [
        "budget_warning",
        "budget_exceeded",
        "provider_down",
        "provider_up",
        "failover",
        "daily_summary"
      ],
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  ]
}
```

## Types d’événements

| Événement | Description | Déclenchement |
|-------|-------------|----------------|
| `budget_warning` | Seuil de budget atteint | Quand la dépense atteint 80 % de la limite |
| `budget_exceeded` | Limite de budget dépassée | Quand la dépense dépasse la limite configurée |
| `provider_down` | Le fournisseur devient défaillant | Quand le taux de succès passe sous 70 % |
| `provider_up` | Le fournisseur se rétablit | Quand un fournisseur défaillant redevient sain |
| `failover` | Requête basculée | Quand une requête passe au fournisseur de secours |
| `daily_summary` | Résumé quotidien d’usage | Une fois par jour, à minuit UTC |

## Formats de webhook

### Slack

Détecté automatiquement quand l’URL contient `slack.com`.

**Exemple de message :**
```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

**Format :**
```json
{
  "text": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)"
      }
    }
  ]
}
```

### Discord

Détecté automatiquement quand l’URL contient `discord.com`.

**Exemple d’embed :**
- **Titre :** budget_warning
- **Description :** ⚠️ Alerte de budget : budget quotidien à 85,0 % (8,50 $ / 10,00 $)
- **Couleur :** ambre (#FBBF24)
- **Horodatage :** 2026-03-05T10:30:00Z

**Format :**
```json
{
  "content": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
  "embeds": [
    {
      "title": "budget_warning",
      "description": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
      "timestamp": "2026-03-05T10:30:00Z",
      "color": 16432932
    }
  ]
}
```

### JSON générique

Utilisé pour toutes les autres URL.

**Format :**
```json
{
  "event": "budget_warning",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "period": "daily",
    "spent": 8.5,
    "limit": 10.0,
    "percent": 85.0,
    "project": ""
  }
}
```

## Structures des données d’événement

### Alerte / dépassement de budget

```json
{
  "event": "budget_warning",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "period": "daily",
    "spent": 8.5,
    "limit": 10.0,
    "percent": 85.0,
    "action": "warn",
    "project": "my-project"
  }
}
```

### Fournisseur défaillant / rétabli

```json
{
  "event": "provider_down",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "provider": "anthropic-primary",
    "status": "unhealthy",
    "error": "connection timeout",
    "latency_ms": 0
  }
}
```

### Bascule

```json
{
  "event": "failover",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "from_provider": "anthropic-primary",
    "to_provider": "anthropic-backup",
    "reason": "rate limit exceeded",
    "session_id": "sess_abc123"
  }
}
```

### Résumé quotidien

```json
{
  "event": "daily_summary",
  "timestamp": "2026-03-05T00:00:00Z",
  "data": {
    "date": "2026-03-04",
    "total_cost": 25.50,
    "total_requests": 150,
    "total_input_tokens": 125000,
    "total_output_tokens": 35000,
    "by_provider": {
      "anthropic": 18.20,
      "openai": 7.30
    }
  }
}
```

## Mise en place par plateforme

### Slack

1. Rendez-vous sur [Slack API](https://api.slack.com/apps)
2. Créez une application ou choisissez-en une existante
3. Activez « Incoming Webhooks »
4. Ajoutez le webhook à l’espace de travail
5. Copiez l’URL du webhook (elle commence par `https://hooks.slack.com/`)

**Configuration :**
```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
      "events": ["budget_warning", "provider_down"]
    }
  ]
}
```

### Discord

1. Ouvrez les paramètres du serveur Discord
2. Allez dans Intégrations → Webhooks
3. Cliquez sur « New Webhook »
4. Choisissez le salon et copiez l’URL du webhook

**Configuration :**
```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://discord.com/api/webhooks/123456789/XXXXXXXXXXXXXXXXXXXX",
      "events": ["budget_exceeded", "failover"]
    }
  ]
}
```

### Webhook personnalisé

Pour vos intégrations, utilisez le format JSON générique :

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://your-server.com/webhook",
      "events": ["budget_warning", "daily_summary"],
      "headers": {
        "Authorization": "Bearer YOUR_SECRET_TOKEN",
        "X-Custom-Header": "value"
      }
    }
  ]
}
```

## Configuration dans l’interface web

Les réglages des webhooks se trouvent sur `http://localhost:19840/settings` :

1. Ouvrez l’onglet « Webhooks »
2. Cliquez sur « Add Webhook »
3. Saisissez l’URL du webhook
4. Choisissez les événements auxquels vous abonner
5. (Facultatif) Ajoutez des en-têtes personnalisés
6. Cliquez sur « Test » pour vérifier la configuration
7. Cliquez sur « Save »

## Points d’API

### Lister les webhooks

```bash
GET /api/v1/webhooks
```

### Ajouter un webhook

```bash
POST /api/v1/webhooks
Content-Type: application/json

{
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "events": ["budget_warning", "provider_down"]
}
```

### Mettre à jour un webhook

```bash
PUT /api/v1/webhooks/{id}
Content-Type: application/json

{
  "enabled": false
}
```

### Supprimer un webhook

```bash
DELETE /api/v1/webhooks/{id}
```

### Tester un webhook

```bash
POST /api/v1/webhooks/{id}/test
```

Envoie un message de test pour vérifier la configuration.

## Exemples de messages

### Alerte de budget (Slack)

```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

### Budget dépassé (Discord)

```
🚫 Budget Exceeded: monthly limit of $200.00 reached (spent: $205.50). Action: block
```

### Fournisseur défaillant (Slack)

```
🔴 Provider Down: anthropic-primary is unhealthy. Error: connection timeout
```

### Fournisseur rétabli (Discord)

```
🟢 Provider Up: anthropic-primary is healthy again (latency: 1250ms)
```

### Bascule (Slack)

```
🔄 Failover: Switched from anthropic-primary to anthropic-backup. Reason: rate limit exceeded
```

### Résumé quotidien (Discord)

```
📊 Daily Summary (2026-03-04): 150 requests, $25.50 total cost, 125000 input / 35000 output tokens
```

## Bonnes pratiques

1. **Séparez les webhooks** — créez des webhooks distincts selon les types d’événements
2. **Testez avant d’activer** — vérifiez toujours la configuration avant d’enregistrer
3. **Sécurisez vos webhooks** — utilisez HTTPS et des en-têtes d’authentification
4. **Surveillez les échecs** — consultez les journaux du daemon si les notifications s’arrêtent
5. **Évitez les données sensibles** — ne mettez ni clés d’API ni jetons dans les URL de webhook
6. **Mettez en place des alertes** — abonnez-vous aux événements critiques comme `budget_exceeded` et `provider_down`

## Dépannage

### Le webhook ne reçoit aucun message

1. Vérifiez que le webhook est activé dans la configuration
2. Vérifiez l’URL (testez avec curl)
3. Vérifiez que les événements sont correctement configurés
4. Cherchez les erreurs de webhook dans les journaux du daemon : `tail -f ~/.zen/zend.log`
5. Testez le webhook via l’API : `POST /api/v1/webhooks/{id}/test`

### Le webhook Slack échoue

1. Vérifiez que l’URL commence par `https://hooks.slack.com/`
2. Vérifiez que le webhook n’a pas été révoqué dans les réglages Slack
3. Assurez-vous que l’espace de travail n’a pas désactivé les webhooks entrants
4. Testez avec curl :
   ```bash
   curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

### Le webhook Discord échoue

1. Vérifiez que l’URL commence par `https://discord.com/api/webhooks/`
2. Vérifiez que le webhook n’a pas été supprimé dans les réglages Discord
3. Assurez-vous que le bot a le droit de publier dans le salon
4. Testez avec curl :
   ```bash
   curl -X POST https://discord.com/api/webhooks/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"content":"test"}'
   ```

### Le webhook personnalisé ne fonctionne pas

1. Vérifiez que le point d’accès est joignable (testez avec curl)
2. Vérifiez les en-têtes d’authentification
3. Assurez-vous que le point d’accès accepte les requêtes POST
4. Vérifiez qu’il renvoie un code d’état 2xx
5. Consultez ses journaux à la recherche d’erreurs

## Sécurité

1. **Protégez les URL de webhook** — traitez-les comme des secrets
2. **Utilisez HTTPS** — toujours, pour les points d’accès de webhook
3. **Validez les signatures** — mettez en place une validation de signature pour vos webhooks
4. **Limitez le débit** — appliquez une limitation de débit sur vos points d’accès
5. **Ne journalisez pas de données sensibles** — évitez d’enregistrer la charge utile complète

## Configuration avancée

### Webhooks conditionnels

Envoyez différents événements vers différents webhooks :

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/CRITICAL/ALERTS",
      "events": ["budget_exceeded", "provider_down"]
    },
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/DAILY/REPORTS",
      "events": ["daily_summary"]
    },
    {
      "enabled": true,
      "url": "https://discord.com/api/webhooks/MONITORING",
      "events": ["failover", "provider_up"]
    }
  ]
}
```

### En-têtes personnalisés pour l’authentification

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://your-server.com/webhook",
      "events": ["budget_warning"],
      "headers": {
        "Authorization": "Bearer YOUR_SECRET_TOKEN",
        "X-API-Key": "your-api-key",
        "X-Webhook-Source": "gozen"
      }
    }
  ]
}
```
