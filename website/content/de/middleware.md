---
sidebar_position: 15
title: Middleware-Pipeline (BETA)
---

# Middleware-Pipeline (BETA)

:::warning BETA-Funktion
Die Middleware-Pipeline ist derzeit in der Beta-Phase. Sie ist standardmäßig deaktiviert und muss ausdrücklich konfiguriert werden.
:::

Erweitere GoZen mit einsteckbarer Middleware: Umformung von Anfragen und Antworten, Protokollierung, Ratenbegrenzung und eigene Verarbeitung.

## Funktionen

- **Einsteckbare Architektur** — eigene Logik ergänzen, ohne den Kern zu ändern
- **Ausführung nach Priorität** — die Reihenfolge der Middleware steuern
- **Haken für Anfrage und Antwort** — vor dem Senden und nach dem Empfangen eingreifen
- **Mitgelieferte Middleware** — Kontextinjektion, Protokollierung, Ratenbegrenzung, Kompression
- **Plugin-Loader** — Middleware aus lokalen Dateien oder von entfernten URLs laden
- **Fehlerbehandlung** — saubere Behandlung mit Rückfallverhalten

## Architektur

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

## Konfiguration

### Middleware-Pipeline aktivieren

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

**Optionen:**

| Option | Beschreibung |
|--------|-------------|
| `enabled` | Aktiviert die Middleware-Pipeline |
| `pipeline` | Liste der Middleware-Konfigurationen |
| `name` | Bezeichner der Middleware |
| `priority` | Ausführungsreihenfolge (kleiner = früher) |
| `config` | Middleware-eigene Konfiguration |

## Mitgelieferte Middleware

### 1. Kontextinjektion

Fügt eigenen Kontext in Anfragen ein.

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

**Anwendungsfälle:**
- Systemprompts ergänzen
- Sitzungsmetadaten einfügen
- Nutzerkontext ergänzen

### 2. Anfrageprotokoll

Protokolliert alle Anfragen und Antworten.

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

**Anwendungsfälle:**
- Fehlersuche
- Prüfpfade
- Performance-Beobachtung

### 3. Ratenbegrenzer

Begrenzt die Anfragerate je Anbieter oder global.

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

**Anwendungsfälle:**
- Rate-Limit-Fehler vermeiden
- API-Nutzung steuern
- Vor Missbrauch schützen

### 4. Kompression (BETA)

Komprimiert den Kontext, wenn die Token-Zahl den Schwellenwert überschreitet.

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

Details stehen unter [Kontextkompression](./compression.md).

### 5. Sitzungsgedächtnis (BETA)

Hält das Gesprächsgedächtnis über Sitzungen hinweg.

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

**Anwendungsfälle:**
- Nutzerpräferenzen merken
- Gesprächsverlauf verfolgen
- Kontext über Sitzungen hinweg bewahren

### 6. Orchestrierung (BETA)

Schickt Anfragen an mehrere Anbieter und führt die Antworten zusammen.

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

**Anwendungsfälle:**
- Ausgaben mehrerer Modelle vergleichen
- Redundanz für kritische Anfragen
- Qualitätsgewinn durch Konsens

## Eigene Middleware

### Middleware-Schnittstelle

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

### Beispiel: eigenen Header einfügen

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

### Eigene Middleware laden

#### Lokales Plugin

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

#### Entferntes Plugin

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

## Weboberfläche

Die Middleware-Einstellungen liegen unter `http://localhost:19840/settings`:

1. Zum Reiter „Middleware“ wechseln (mit BETA-Kennzeichnung)
2. „Enable Middleware Pipeline“ einschalten
3. Middleware zur Pipeline hinzufügen oder daraus entfernen
4. Priorität und Konfiguration anpassen
5. Einzelne Middleware ein- oder ausschalten
6. Auf „Save“ klicken

## API-Endpunkte

### Middleware auflisten

```bash
GET /api/v1/middleware
```

Antwort:
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

### Middleware hinzufügen

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

### Middleware aktualisieren

```bash
PUT /api/v1/middleware/{name}
Content-Type: application/json

{
  "enabled": false
}
```

### Middleware entfernen

```bash
DELETE /api/v1/middleware/{name}
```

### Pipeline neu laden

```bash
POST /api/v1/middleware/reload
```

## Anwendungsfälle

### Entwicklungsumgebung

Debug-Protokollierung und Anfrageeinsicht ergänzen:

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

### Produktionsumgebung

Ratenbegrenzung und Überwachung ergänzen:

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

### Vergleich mehrerer Anbieter

Mit Orchestrierung die Ausgaben vergleichen:

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

## Bewährtes Vorgehen

1. **Passende Prioritäten wählen** — kleinere Zahlen laufen zuerst
2. **Middleware fokussiert halten** — jede Middleware erledigt genau eine Sache gut
3. **Fehler sauber behandeln** — ein Fehler darf die Pipeline nicht zerreißen
4. **Gründlich testen** — das Verhalten vor der Produktion prüfen
5. **Performance beobachten** — den Zusatzaufwand messen
6. **Konfiguration dokumentieren** — die Optionen klar beschreiben

## Grenzen

1. **Zusätzliche Latenz** — jede Middleware kostet Zeit
2. **Komplexität** — zu viele Middleware erschweren die Fehlersuche
3. **Plugin-Sicherheit** — entfernte Plugins verlangen Vertrauen und Prüfung
4. **Fehlerausbreitung** — Fehler in einer Middleware können alle Anfragen treffen
5. **Konfigurationsaufwand** — komplexe Pipelines sind schwerer zu pflegen

## Fehlersuche

### Die Middleware läuft nicht

1. Prüfen, ob `middleware.enabled` `true` ist
2. Prüfen, ob die Middleware in der Pipeline aktiviert ist
3. Prüfen, ob die Priorität richtig gesetzt ist
4. Die Daemon-Protokolle auf Middleware-Fehler durchsehen

### Unerwartetes Verhalten

1. Die Ausführungsreihenfolge prüfen (Priorität)
2. Die Konfiguration prüfen
3. Die Middleware isoliert testen
4. Ihre Protokolle durchsehen

### Performance-Probleme

1. Die langsame Middleware finden (Protokolle prüfen)
2. Die Zahl der Middleware verringern
3. Die Implementierung optimieren
4. Nicht zwingend nötige Middleware abschalten

### Plugin lässt sich nicht laden

1. Den Pfad des Plugins prüfen
2. Prüfen, ob es für die richtige Architektur kompiliert ist
3. Die Prüfsumme vergleichen (bei entfernten Plugins)
4. Die Protokolle des Plugins durchsehen

## Sicherheitsaspekte

1. **Plugins prüfen** — nur vertrauenswürdige Plugins laden
2. **Prüfsummen vergleichen** — bei entfernten Plugins immer
3. **Plugins isolieren** — den Betrieb in einer abgeschotteten Umgebung erwägen
4. **Middleware auditieren** — den Code vor dem Ausrollen lesen
5. **Verhalten beobachten** — auf ungewöhnliches Verhalten achten

## Geplante Erweiterungen

- Unterstützung für WebAssembly-Plugins für plattformübergreifende Nutzung
- Ein Marktplatz zum Teilen von Community-Plugins
- Visueller Pipeline-Editor in der Weboberfläche
- Performance-Profiling für Middleware
- Hot-Reload bei Plugin-Updates
- Ein Test-Framework für Middleware
