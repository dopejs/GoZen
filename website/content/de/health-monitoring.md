---
sidebar_position: 12
title: Health-Monitoring
---

# Health-Monitoring und Lastverteilung

Beobachte den Zustand der Anbieter in Echtzeit und leite Anfragen automatisch an den besten verfügbaren Anbieter.

## Funktionen

- **Health-Checks in Echtzeit** — regelmäßige Prüfungen mit konfigurierbarem Intervall
- **Erfolgsquote** — der Zustand wird aus der Erfolgsquote der Anfragen berechnet
- **Latenzüberwachung** — durchschnittliche Antwortzeit je Anbieter
- **Mehrere Strategien** — Failover, Round Robin, geringste Latenz, geringste Kosten
- **Automatisches Failover** — Wechsel auf Ersatzanbieter, wenn der primäre ungesund ist
- **Health-Dashboard** — visuelle Statusanzeigen in der Weboberfläche

## Konfiguration

### Health-Monitoring aktivieren

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

**Optionen:**
- `interval` — wie oft der Zustand geprüft wird (Standard: 5 Minuten)
- `timeout` — Zeitlimit der Prüfanfrage (Standard: 10 Sekunden)
- `endpoint` — zu testender API-Endpunkt (Standard: `/v1/messages`)
- `method` — HTTP-Methode des Health-Checks (Standard: `POST`)

### Lastverteilung konfigurieren

```json
{
  "load_balancing": {
    "strategy": "least-latency",
    "health_aware": true,
    "cache_ttl": "30s"
  }
}
```

## Strategien der Lastverteilung

### 1. Failover (Standard)

Nutzt die Anbieter der Reihe nach und wechselt bei Fehlern zum nächsten.

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

**Verhalten:**
1. `anthropic-primary` versuchen
2. Schlägt das fehl, `anthropic-backup` versuchen
3. Schlägt das fehl, `openai` versuchen
4. Scheitern alle, wird ein Fehler zurückgegeben

**Am besten für:** Produktionslasten mit klarer Primär-/Backup-Hierarchie

### 2. Round Robin

Verteilt Anfragen gleichmäßig auf alle gesunden Anbieter.

```json
{
  "load_balancing": {
    "strategy": "round-robin"
  }
}
```

**Verhalten:**
- Anfrage 1 → Anbieter A
- Anfrage 2 → Anbieter B
- Anfrage 3 → Anbieter C
- Anfrage 4 → Anbieter A (der Zyklus wiederholt sich)

**Am besten für:** Last auf mehrere Konten verteilen, um Rate-Limits zu umgehen

### 3. Geringste Latenz

Leitet an den Anbieter mit der niedrigsten durchschnittlichen Latenz.

```json
{
  "load_balancing": {
    "strategy": "least-latency"
  }
}
```

**Verhalten:**
- Erfasst die durchschnittliche Antwortzeit je Anbieter
- Leitet an den schnellsten Anbieter
- Aktualisiert die Messwerte alle 30 Sekunden (über `cache_ttl` konfigurierbar)

**Am besten für:** latenzempfindliche Anwendungen und Interaktion in Echtzeit

### 4. Geringste Kosten

Leitet an den günstigsten Anbieter für das angefragte Modell.

```json
{
  "load_balancing": {
    "strategy": "least-cost"
  }
}
```

**Verhalten:**
- Vergleicht die Preise der Anbieter
- Leitet an die günstigste Option
- Berücksichtigt Kosten für Eingabe- und Ausgabe-Tokens

**Am besten für:** Kostenoptimierung und Batch-Verarbeitung

## Zustandsstufen

Anbieter werden in vier Zustände eingeteilt:

| Status | Erfolgsquote | Verhalten |
|--------|--------------|----------|
| **Gesund** | ≥ 95 % | Normale Priorität |
| **Beeinträchtigt** | 70–95 % | Niedrigere Priorität, weiterhin nutzbar |
| **Ungesund** | < 70 % | Wird übersprungen, außer es gibt keine gesunden Anbieter |
| **Unbekannt** | Keine Daten | Gilt zunächst als gesund |

### Health-bewusstes Routing

Bei `health_aware: true` (Standard):
- Gesunde Anbieter werden bevorzugt
- Beeinträchtigte Anbieter dienen als Rückfallebene
- Ungesunde Anbieter werden übersprungen, außer alle anderen scheitern

## Dashboard der Weboberfläche

Das Health-Dashboard liegt unter `http://localhost:19840/health`:

### Anbieterstatus

- **Statusanzeige** — grün (gesund), gelb (beeinträchtigt), rot (ungesund)
- **Erfolgsquote** — Anteil erfolgreicher Anfragen
- **Durchschnittliche Latenz** — mittlere Antwortzeit in Millisekunden
- **Letzte Prüfung** — Zeitstempel des jüngsten Health-Checks
- **Fehleranzahl** — Zahl der jüngsten Fehlschläge

### Zeitverlauf der Messwerte

- **Latenzdiagramm** — Verlauf der Antwortzeiten
- **Diagramm der Erfolgsquote** — Verlauf des Zustands
- **Anfragevolumen** — Anfragen je Anbieter

## API-Endpunkte

### Anbieterzustand abrufen

```bash
GET /api/v1/health/providers
```

Antwort:
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

### Anbieter-Messwerte abrufen

```bash
GET /api/v1/health/providers/{name}/metrics?period=1h
```

Antwort:
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

### Health-Check manuell auslösen

```bash
POST /api/v1/health/check
Content-Type: application/json

{
  "provider": "anthropic-primary"
}
```

## Webhook-Benachrichtigungen

Werde benachrichtigt, wenn sich der Status eines Anbieters ändert:

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

**Ereignistypen:**
- `provider_down` — der Anbieter wird ungesund
- `provider_up` — der Anbieter ist wieder gesund
- `failover` — die Anfrage wurde auf einen Ersatzanbieter umgeleitet

## Szenariobasiertes Routing

Kombiniere Health-Monitoring mit szenariobasiertem Routing für eine kluge Verteilung der Anfragen:

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

Details stehen unter [Szenario-Routing](./routing.md).

## Bewährtes Vorgehen

1. **Passendes Intervall wählen** — 5 Minuten reichen meist, 1 Minute für kritische Systeme
2. **Health-bewusstes Routing nutzen** — in der Produktion immer aktivieren
3. **Beeinträchtigte Anbieter beobachten** — nachsehen, sobald die Erfolgsquote unter 95 % fällt
4. **Strategien kombinieren** — Failover für Primär/Backup, Round Robin zur Lastverteilung
5. **Webhooks aktivieren** — sofort erfahren, wenn ein Anbieter ausfällt
6. **Dashboard regelmäßig prüfen** — Trends ansehen, um Muster zu erkennen

## Fehlersuche

### Health-Checks schlagen fehl

1. Prüfen, ob die API-Schlüssel der Anbieter gültig sind
2. Die Netzwerkverbindung zu den Anbieter-Endpunkten prüfen
3. Das Zeitlimit erhöhen, wenn Anbieter langsam sind: `"timeout": "30s"`
4. Die Daemon-Protokolle auf konkrete Fehlermeldungen durchsehen

### Falsche Latenzwerte

1. Die Latenz umfasst Netzwerkzeit und Verarbeitungszeit der API
2. Prüfen, ob ein Proxy oder VPN zusätzliche Zeit kostet
3. Messwerte werden standardmäßig 30 Sekunden zwischengespeichert (über `cache_ttl` konfigurierbar)

### Failover greift nicht

1. `health_aware: true` in der Lastverteilungs-Konfiguration prüfen
2. Prüfen, ob im Profil Ersatzanbieter konfiguriert sind
3. Sicherstellen, dass Health-Checks aktiviert sind und laufen
4. Failover-Ereignisse in der Weboberfläche oder den Protokollen prüfen

### Anbieter bleibt im Zustand „ungesund“

1. Health-Check manuell über die API auslösen
2. Prüfen, ob der Anbieter wirklich ausgefallen ist (mit curl testen)
3. Daemon neu starten, um den Zustand zurückzusetzen: `zen daemon restart`
4. Die Fehlerprotokolle nach der Ursache durchsehen

## Auswirkung auf die Performance

- **Health-Checks** — minimaler Aufwand, laufen in einer Hintergrund-Goroutine
- **Zwischenspeicherung der Messwerte** — 30 Sekunden TTL sparen Datenbankabfragen
- **Atomare Operationen** — threadsichere Zähler für nebenläufige Anfragen
- **Ohne Blockade** — Health-Checks blockieren die Anfrageverarbeitung nicht

## Erweiterte Konfiguration

### Eigene Nutzlast für den Health-Check

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

### Health-Einstellungen je Anbieter

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
