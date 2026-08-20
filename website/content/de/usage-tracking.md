---
sidebar_position: 11
title: Nutzungserfassung und Budget
---

# Nutzungserfassung und Budgetkontrolle

Erfasse Token-Verbrauch und Kosten über Anbieter, Modelle und Projekte hinweg. Setze Ausgabengrenzen, die automatisch durchgesetzt werden.

## Funktionen

- **Erfassung in Echtzeit** — Token-Verbrauch und Kosten je Anfrage beobachten
- **Mehrdimensionale Aggregation** — nach Anbieter, Modell, Projekt und Zeitraum
- **Budgetgrenzen** — tägliche, wöchentliche und monatliche Ausgabenobergrenzen
- **Automatische Aktionen** — bei Überschreitung warnen, herabstufen oder blockieren
- **Kostenschätzung** — genaue Preise für alle großen KI-Modelle
- **Historische Daten** — SQLite-Speicher mit stündlicher Aggregation für die Performance

## Konfiguration

### Nutzungserfassung aktivieren

```json
{
  "usage_tracking": {
    "enabled": true,
    "db_path": "~/.zen/usage.db"
  }
}
```

### Modellpreise konfigurieren

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

**Modellabgleich**: Zuerst werden exakte Modellnamen geprüft, danach die Präfixe der Modellfamilie.

### Budgetgrenzen festlegen

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

## Budget-Aktionen

| Aktion | Verhalten |
|--------|-----------|
| `warn` | Schreibt eine Warnung ins Protokoll und sendet eine Webhook-Benachrichtigung, lässt die Anfrage aber zu |
| `downgrade` | Wechselt auf ein günstigeres Modell (z. B. opus → sonnet → haiku) |
| `block` | Weist die Anfrage mit Statuscode 429 ab |

## Weboberfläche

Das Nutzungs-Dashboard liegt unter `http://localhost:19840/usage`:

- **Überblick** — Gesamtkosten, Anfragen und Tokens der laufenden Periode
- **Nach Anbieter** — Kostenaufschlüsselung je Anbieter
- **Nach Modell** — Nutzungsstatistik je Modell
- **Nach Projekt** — Kosten je Projekt (über Projektbindungen)
- **Zeitverlauf** — stündliche und tägliche Kostentrends
- **Budgetstatus** — visuelle Anzeige der Tages-, Wochen- und Monatsgrenzen

## API-Endpunkte

### Nutzungsübersicht abrufen

```bash
GET /api/v1/usage/summary?period=daily
```

Antwort:
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

### Budgetstatus abrufen

```bash
GET /api/v1/budget/status
```

Antwort:
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

### Budgetgrenzen aktualisieren

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

## Erfassung je Projekt

Kosten je Projekt über Verzeichnisbindungen erfassen:

```bash
# Bind current directory to a profile
zen bind work-profile

# All requests from this directory are tagged with the project path
# View costs in Web UI under "By Project"
```

## Webhook-Benachrichtigungen

Werde benachrichtigt, wenn ein Budget überschritten wird:

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

Die vollständige Konfiguration steht unter [Webhooks](./webhooks.md).

## Bewährtes Vorgehen

1. **Mit Warnungen beginnen** — anfangs `warn` nutzen, um die Nutzungsmuster zu verstehen
2. **Realistische Grenzen setzen** — an historischen Verbrauchsdaten orientieren
3. **Herabstufung in der Entwicklung** — beim Testen automatisch auf günstigere Modelle wechseln
4. **Blockieren der Produktion vorbehalten** — `block` nur für harte Ausgabengrenzen
5. **Täglich prüfen** — regelmäßig ins Dashboard schauen, um Überraschungen zu vermeiden
6. **Webhooks aktivieren** — in Echtzeit gewarnt werden, wenn Grenzen näher rücken

## Fehlersuche

### Nutzung wird nicht erfasst

1. Prüfen, ob `usage_tracking.enabled` in der Konfiguration `true` ist
2. Prüfen, ob der Datenbankpfad beschreibbar ist: `~/.zen/usage.db`
3. Daemon neu starten: `zen daemon restart`

### Falsche Kosten

1. Prüfen, ob die Modellpreise in der Konfiguration den aktuellen Tarifen entsprechen
2. Den Modellnamensabgleich prüfen (exakt oder über das Familienpräfix)
3. Die Preiskonfiguration aktualisieren, wenn Anbieter ihre Tarife ändern

### Budget wird nicht durchgesetzt

1. Prüfen, ob die Budgetkonfiguration aktiviert ist
2. Prüfen, ob eine Aktion gesetzt ist (`warn`, `downgrade` oder `block`)
3. Die Daemon-Protokolle auf Fehler des Budget-Checkers durchsehen

## Performance

- **Stündliche Aggregation** — Rohdaten werden stündlich verdichtet, das entlastet Abfragen
- **Indizierte Abfragen** — Datenbankindizes auf Anbieter, Modell, Projekt und Zeitstempel
- **Sparsamer Speicher** — rund 1 KB je Anfrage, rund 30 MB je 30.000 Anfragen
- **Schnelles Dashboard** — Abfragezeiten unter einer Sekunde bei typischer Nutzung
