---
sidebar_position: 13
title: Webhooks
---

# Webhooks

Erhalte Benachrichtigungen in Echtzeit — Budgetwarnungen, Statusänderungen von Anbietern, Tageszusammenfassungen — über Slack, Discord oder einen eigenen Webhook.

## Funktionen

- **Mehrere Formate** — Slack, Discord oder generisches JSON
- **Ereignisfilter** — nur bestimmte Ereignistypen abonnieren
- **Eigene Header** — Authentifizierung oder eigene Header ergänzen
- **Asynchroner Versand** — die Zustellung blockiert nichts
- **Automatische Formatierung** — ansprechende Nachrichten mit Emojis und Farben
- **Testfunktion** — die Konfiguration vor dem Aktivieren prüfen

## Konfiguration

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

## Ereignistypen

| Ereignis | Beschreibung | Auslöser |
|-------|-------------|----------------|
| `budget_warning` | Budgetschwelle erreicht | Wenn die Ausgaben 80 % der Grenze erreichen |
| `budget_exceeded` | Budgetgrenze überschritten | Wenn die Ausgaben die konfigurierte Grenze übersteigen |
| `provider_down` | Anbieter wird ungesund | Wenn die Erfolgsquote unter 70 % fällt |
| `provider_up` | Anbieter erholt sich | Wenn ein ungesunder Anbieter wieder gesund wird |
| `failover` | Anfrage umgeleitet | Wenn eine Anfrage auf den Ersatzanbieter wechselt |
| `daily_summary` | Tägliche Nutzungsübersicht | Einmal täglich um Mitternacht UTC |

## Webhook-Formate

### Slack

Wird automatisch erkannt, wenn die URL `slack.com` enthält.

**Beispielnachricht:**
```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

**Format:**
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

Wird automatisch erkannt, wenn die URL `discord.com` enthält.

**Beispiel-Embed:**
- **Titel:** budget_warning
- **Beschreibung:** ⚠️ Budgetwarnung: Tagesbudget bei 85,0 % (8,50 $ / 10,00 $)
- **Farbe:** Bernstein (#FBBF24)
- **Zeitstempel:** 2026-03-05T10:30:00Z

**Format:**
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

### Generisches JSON

Wird für alle übrigen URLs verwendet.

**Format:**
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

## Datenstrukturen der Ereignisse

### Budgetwarnung / -überschreitung

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

### Anbieter ausgefallen / wieder da

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

### Failover

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

### Tageszusammenfassung

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

## Einrichtung je Plattform

### Slack

1. Öffne die [Slack API](https://api.slack.com/apps)
2. Lege eine neue App an oder wähle eine bestehende
3. Aktiviere „Incoming Webhooks“
4. Füge den Webhook dem Workspace hinzu
5. Kopiere die Webhook-URL (sie beginnt mit `https://hooks.slack.com/`)

**Konfiguration:**
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

1. Öffne die Servereinstellungen in Discord
2. Gehe zu Integrationen → Webhooks
3. Klicke auf „New Webhook“
4. Wähle den Kanal und kopiere die Webhook-URL

**Konfiguration:**
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

### Eigener Webhook

Für eigene Integrationen nutze das generische JSON-Format:

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

## Konfiguration in der Weboberfläche

Die Webhook-Einstellungen liegen unter `http://localhost:19840/settings`:

1. Wechsle zum Reiter „Webhooks“
2. Klicke auf „Add Webhook“
3. Gib die Webhook-URL ein
4. Wähle die Ereignisse aus, die du abonnieren willst
5. (Optional) Ergänze eigene Header
6. Klicke auf „Test“, um die Konfiguration zu prüfen
7. Klicke auf „Save“

## API-Endpunkte

### Webhooks auflisten

```bash
GET /api/v1/webhooks
```

### Webhook hinzufügen

```bash
POST /api/v1/webhooks
Content-Type: application/json

{
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "events": ["budget_warning", "provider_down"]
}
```

### Webhook aktualisieren

```bash
PUT /api/v1/webhooks/{id}
Content-Type: application/json

{
  "enabled": false
}
```

### Webhook löschen

```bash
DELETE /api/v1/webhooks/{id}
```

### Webhook testen

```bash
POST /api/v1/webhooks/{id}/test
```

Sendet eine Testnachricht, um die Konfiguration zu prüfen.

## Beispielnachrichten

### Budgetwarnung (Slack)

```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

### Budget überschritten (Discord)

```
🚫 Budget Exceeded: monthly limit of $200.00 reached (spent: $205.50). Action: block
```

### Anbieter ausgefallen (Slack)

```
🔴 Provider Down: anthropic-primary is unhealthy. Error: connection timeout
```

### Anbieter wieder da (Discord)

```
🟢 Provider Up: anthropic-primary is healthy again (latency: 1250ms)
```

### Failover (Slack)

```
🔄 Failover: Switched from anthropic-primary to anthropic-backup. Reason: rate limit exceeded
```

### Tageszusammenfassung (Discord)

```
📊 Daily Summary (2026-03-04): 150 requests, $25.50 total cost, 125000 input / 35000 output tokens
```

## Bewährtes Vorgehen

1. **Getrennte Webhooks nutzen** — für unterschiedliche Ereignistypen eigene Webhooks anlegen
2. **Vor dem Aktivieren testen** — die Konfiguration immer vor dem Speichern prüfen
3. **Eigene Webhooks absichern** — HTTPS und Authentifizierungs-Header verwenden
4. **Fehlschläge beobachten** — die Daemon-Protokolle prüfen, wenn keine Benachrichtigungen mehr kommen
5. **Keine sensiblen Daten** — keine API-Schlüssel oder Tokens in Webhook-URLs
6. **Alarme einrichten** — kritische Ereignisse wie `budget_exceeded` und `provider_down` abonnieren

## Fehlersuche

### Der Webhook empfängt keine Nachrichten

1. Prüfen, ob der Webhook in der Konfiguration aktiviert ist
2. Die URL prüfen (mit curl testen)
3. Prüfen, ob die Ereignisse richtig konfiguriert sind
4. Die Daemon-Protokolle auf Webhook-Fehler prüfen: `tail -f ~/.zen/zend.log`
5. Den Webhook über die API testen: `POST /api/v1/webhooks/{id}/test`

### Der Slack-Webhook schlägt fehl

1. Prüfen, ob die URL mit `https://hooks.slack.com/` beginnt
2. Prüfen, ob der Webhook in den Slack-Einstellungen widerrufen wurde
3. Sicherstellen, dass der Workspace eingehende Webhooks nicht deaktiviert hat
4. Mit curl testen:
   ```bash
   curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

### Der Discord-Webhook schlägt fehl

1. Prüfen, ob die URL mit `https://discord.com/api/webhooks/` beginnt
2. Prüfen, ob der Webhook in den Discord-Einstellungen gelöscht wurde
3. Sicherstellen, dass der Bot im Kanal schreiben darf
4. Mit curl testen:
   ```bash
   curl -X POST https://discord.com/api/webhooks/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"content":"test"}'
   ```

### Der eigene Webhook funktioniert nicht

1. Prüfen, ob der Endpunkt erreichbar ist (mit curl testen)
2. Die Authentifizierungs-Header prüfen
3. Sicherstellen, dass der Endpunkt POST-Anfragen annimmt
4. Prüfen, ob er einen 2xx-Statuscode zurückgibt
5. Die Protokolle des Endpunkts auf Fehler durchsehen

## Sicherheitsaspekte

1. **Webhook-URLs schützen** — sie wie Geheimnisse behandeln
2. **HTTPS verwenden** — für Webhook-Endpunkte immer
3. **Signaturen prüfen** — für eigene Webhooks eine Signaturprüfung einbauen
4. **Rate Limiting** — die Webhook-Endpunkte begrenzen
5. **Keine sensiblen Daten protokollieren** — vollständige Nutzlasten nicht mitschreiben

## Erweiterte Konfiguration

### Bedingte Webhooks

Verschiedene Ereignisse an verschiedene Webhooks senden:

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

### Eigene Header zur Authentifizierung

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
