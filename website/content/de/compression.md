---
sidebar_position: 14
title: Kontextkompression (BETA)
---

# Kontextkompression (BETA)

:::warning BETA-Funktion
Die Kontextkompression ist derzeit in der Beta-Phase. Sie ist standardmäßig deaktiviert und muss ausdrücklich konfiguriert werden.
:::

Komprimiert den Gesprächskontext automatisch, sobald die Token-Zahl den Schwellenwert überschreitet. Das senkt die Kosten und erhält zugleich die Gesprächsqualität.

## Funktionen

- **Automatische Kompression** — wird beim Überschreiten des Token-Schwellenwerts ausgelöst
- **Kluge Zusammenfassung** — ein günstiges Modell (claude-3-haiku) fasst ältere Nachrichten zusammen
- **Jüngste Nachrichten bleiben erhalten** — die letzten Nachrichten bleiben unangetastet
- **Token-Schätzung** — genaues Zählen vor den API-Aufrufen
- **Statistik** — die Wirksamkeit der Kompression im Blick behalten
- **Transparenter Betrieb** — funktioniert mit allen KI-Clients

## Funktionsweise

1. **Token-Schätzung** — Tokens im Gesprächsverlauf zählen
2. **Schwellenwertprüfung** — mit dem konfigurierten Wert vergleichen (Standard: 50.000)
3. **Nachrichtenauswahl** — ältere Nachrichten für die Kompression bestimmen
4. **Zusammenfassung** — mit einem günstigen Modell eine knappe Zusammenfassung erzeugen
5. **Kontextersetzung** — alte Nachrichten durch die Zusammenfassung ersetzen
6. **Weiterleitung** — den komprimierten Kontext an das Zielmodell senden

## Konfiguration

### Kompression aktivieren

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

**Optionen:**

| Option | Standard | Beschreibung |
|--------|---------|-------------|
| `enabled` | `false` | Aktiviert die Kontextkompression |
| `threshold_tokens` | `50000` | Löst die Kompression aus, sobald der Kontext das überschreitet |
| `target_tokens` | `20000` | Angestrebte Token-Zahl nach der Kompression |
| `summarizer_model` | `claude-3-haiku-20240307` | Modell für die Zusammenfassung |
| `preserve_recent_messages` | `5` | Anzahl der jüngsten Nachrichten, die unangetastet bleiben |
| `tokens_per_char` | `0.25` | Schätzverhältnis für das Token-Zählen |

### Konfiguration je Profil

Kompression für bestimmte Profile aktivieren:

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

## Token-Schätzung

GoZen schätzt Tokens schnell anhand der Zeichenzahl:

```
estimated_tokens = character_count * tokens_per_char
```

**Standardverhältnis:** 0,25 Tokens je Zeichen (1 Token ≈ 4 Zeichen)

**Genauigkeit:** ±10 % für englischen Text, bei anderen Sprachen abweichend

Für exaktes Zählen nutzt GoZen die Bibliothek `tiktoken-go`, sofern verfügbar.

## Kompressionsstrategie

### Nachrichtenauswahl

1. **Systemnachrichten** — bleiben immer erhalten
2. **Jüngste Nachrichten** — die letzten N Nachrichten bleiben erhalten (Standard: 5)
3. **Ältere Nachrichten** — Kandidaten für die Kompression

### Prompt für die Zusammenfassung

```
Summarize the following conversation history concisely while preserving key information, decisions, and context:

[older messages]

Provide a brief summary that captures the essential points.
```

### Ergebnis

```
Original: 45,000 tokens (30 messages)
After compression: 22,000 tokens (summary + 5 recent messages)
Savings: 23,000 tokens (51%)
```

## Weboberfläche

Die Kompressionseinstellungen liegen unter `http://localhost:19840/settings`:

1. Zum Reiter „Compression“ wechseln (mit BETA-Kennzeichnung)
2. „Enable Compression“ einschalten
3. Schwellenwert und Ziel-Token-Zahl anpassen
4. Modell für die Zusammenfassung wählen
5. Anzahl der zu erhaltenden jüngsten Nachrichten festlegen
6. Auf „Save“ klicken

### Statistik-Dashboard

Kompressionsstatistiken ansehen:

- **Kompressionen gesamt** — wie oft die Kompression ausgelöst wurde
- **Eingesparte Tokens** — insgesamt eingesparte Tokens
- **Durchschnittliche Ersparnis** — mittlere Token-Reduktion je Kompression
- **Kompressionsrate** — Anteil der Anfragen, die eine Kompression ausgelöst haben

## API-Endpunkte

### Kompressionsstatistik abrufen

```bash
GET /api/v1/compression/stats
```

Antwort:
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

### Kompressionseinstellungen aktualisieren

```bash
PUT /api/v1/compression/settings
Content-Type: application/json

{
  "enabled": true,
  "threshold_tokens": 60000,
  "target_tokens": 25000
}
```

### Statistik zurücksetzen

```bash
POST /api/v1/compression/stats/reset
```

## Anwendungsfälle

### Lange Programmiersitzungen

**Szenario:** mehrstündige Sitzung mit Claude Code

**Konfiguration:**
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

**Nutzen:** Gesprächskontinuität, ohne an die Kontextgrenze zu stoßen

### Stapelverarbeitung

**Szenario:** viele Dokumente mit KI verarbeiten

**Konfiguration:**
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

**Nutzen:** Kosten bei großen Dokumentmengen senken

### Recherche und Analyse

**Szenario:** lange Recherchesitzungen zu mehreren Themen

**Konfiguration:**
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

**Nutzen:** auf die aktuellen Themen fokussiert bleiben und früheren Kontext bewahren

## Bewährtes Vorgehen

1. **Mit den Standardwerten beginnen** — sie passen für die meisten Fälle
2. **Statistik beobachten** — Kompressionsrate und Ersparnis regelmäßig prüfen
3. **Schwellenwert anpassen** — höher für Modelle mit langem Kontext (Claude Opus), sonst niedriger
4. **Genug Nachrichten erhalten** — 5 bis 10 jüngste Nachrichten für die Kontinuität behalten
5. **Günstiges Modell zum Zusammenfassen** — Haiku ist schnell und preiswert
6. **Vor der Produktion testen** — die Qualität am eigenen Anwendungsfall prüfen

## Grenzen

1. **Qualitätsverlust** — die Zusammenfassung kann feine Details verlieren
2. **Höhere Latenz** — ein zusätzlicher API-Aufruf kommt hinzu
3. **Kostenabwägung** — Kosten der Zusammenfassung gegen eingesparte Tokens
4. **Sprachunterstützung** — am besten für Englisch, bei anderen Sprachen abweichend
5. **Kontextfenster** — das maximale Kontextfenster des Modells lässt sich nicht überschreiten

## Fehlersuche

### Die Kompression wird nicht ausgelöst

1. Prüfen, ob `compression.enabled` `true` ist
2. Prüfen, ob die Token-Zahl den Schwellenwert überschreitet
3. Sicherstellen, dass genügend Nachrichten zum Komprimieren vorhanden sind
4. Die Daemon-Protokolle auf Kompressionsfehler durchsehen

### Schlechte Qualität der Zusammenfassung

1. Ein anderes Modell ausprobieren (z. B. claude-3-sonnet)
2. `preserve_recent_messages` erhöhen, um mehr Kontext zu behalten
3. `target_tokens` anheben, um längere Zusammenfassungen zuzulassen
4. Prüfen, ob das Modell für die Zusammenfassung verfügbar ist und funktioniert

### Höhere Latenz

1. Die Kompression fügt einen zusätzlichen API-Aufruf hinzu (die Zusammenfassung)
2. Ein schnelleres Modell nutzen (Haiku ist am schnellsten)
3. Den Schwellenwert erhöhen, um seltener zu komprimieren
4. Für latenzempfindliche Anwendungen die Kompression erwägen abzuschalten

### Unerwartete Kosten

1. Die Kosten der Zusammenfassung im Nutzungs-Dashboard beobachten
2. Ersparnis und Kosten der Zusammenfassung vergleichen
3. Den Schwellenwert erhöhen, um seltener zu komprimieren
4. Das günstigste verfügbare Modell zum Zusammenfassen verwenden

## Auswirkung auf die Performance

- **Token-Schätzung** — etwa 1 ms je Anfrage (vernachlässigbar)
- **Zusammenfassung** — 1 bis 3 Sekunden (je nach Modell und Nachrichtenzahl)
- **Speicherbedarf** — minimal (rund 1 KB je Kompression)
- **Kostenersparnis** — typischerweise 30 bis 50 % weniger Tokens

## Erweiterte Konfiguration

### Eigener Prompt für die Zusammenfassung

```json
{
  "compression": {
    "enabled": true,
    "custom_prompt": "Create a technical summary of the following conversation, focusing on code changes, decisions, and action items:\n\n{messages}\n\nSummary:"
  }
}
```

### Bedingte Kompression

Kompression nur für bestimmte Szenarien aktivieren:

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

### Mehrstufige Kompression

Sehr lange Gespräche mehrfach komprimieren:

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

## Geplante Erweiterungen

- Auswahl der Nachrichten über semantische Ähnlichkeit
- Zusammenfassung mit mehreren Modellen zum Qualitätsvergleich
- Qualitätsmetriken und Rückmeldungen zur Kompression
- Eigene Kompressionsstrategien je Anwendungsfall
- Anbindung an RAG, um Kontext extern abzulegen
