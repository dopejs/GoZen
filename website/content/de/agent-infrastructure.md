---
sidebar_position: 16
title: Agent-Infrastruktur (BETA)
---

# Agent-Infrastruktur (BETA)

:::warning BETA-Funktion
Die Agent-Infrastruktur ist derzeit in der Beta-Phase. Sie ist standardmäßig deaktiviert und muss ausdrücklich konfiguriert werden.
:::

Eingebaute Unterstützung für autonome Agent-Abläufe: Sitzungsverwaltung, Dateikoordination, Überwachung in Echtzeit und Schutzmechanismen.

## Funktionen

- **Agent-Runtime** — autonome Aufgaben mit vollständiger Lebenszyklusverwaltung ausführen
- **Observatory** — Sitzungen und Aktivitäten in Echtzeit beobachten
- **Schutzmechanismen** — Sicherheitskontrollen und Grenzen für das Verhalten der Agents
- **Koordinator** — dateibasierte Koordination für Abläufe mit mehreren Agents
- **Aufgabenwarteschlange** — Aufgaben mit Priorität und Abhängigkeiten verwalten
- **Sitzungsverwaltung** — Agent-Sitzungen über mehrere Projekte hinweg verfolgen

## Architektur

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

## Konfiguration

### Agent-Infrastruktur aktivieren

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

## Komponenten

### 1. Agent-Runtime

Verwaltet den Ausführungszyklus der Agent-Aufgaben.

**Funktionen:**
- Planung und Ausführung von Aufgaben
- Verwaltung nebenläufiger Aufgaben
- Umgang mit Zeitüberschreitungen
- Automatische Aufräumarbeiten
- Fehlerbehebung

**Konfiguration:**
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

**API:**
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

### 2. Observatory

Beobachtung der Agent-Aktivitäten in Echtzeit.

**Funktionen:**
- Sitzungsverfolgung
- Aktivitätsprotokoll
- Performance-Messwerte
- Statusaktualisierungen
- Historische Daten

**Konfiguration:**
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

**Beobachtete Messwerte:**
- Aktive Sitzungen
- Laufende Aufgaben
- Token-Verbrauch
- API-Aufrufe
- Dateioperationen
- Fehlerquote
- Durchschnittliche Latenz

**API:**
```bash
# Get all active sessions
GET /api/v1/agent/sessions

# Get session details
GET /api/v1/agent/sessions/{session_id}

# Get session metrics
GET /api/v1/agent/sessions/{session_id}/metrics
```

### 3. Schutzmechanismen

Sicherheitskontrollen und Grenzen für das Verhalten der Agents.

**Funktionen:**
- Grenzen für Operationen
- Pfadbeschränkungen
- Blockieren von Befehlen
- Ressourcenkontingente
- Freigabeprozesse

**Konfiguration:**
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

**Durchsetzung:**
- Prüfung vor der Ausführung
- Beobachtung in Echtzeit
- Automatisches Blockieren
- Freigabeabfragen
- Audit-Protokollierung

**API:**
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

### 4. Koordinator

Dateibasierte Koordination für Abläufe mit mehreren Agents.

**Funktionen:**
- Dateisperren
- Änderungserkennung
- Konfliktlösung
- Zustandssynchronisierung
- Ereignisbenachrichtigungen

**Konfiguration:**
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

**Anwendungsfälle:**
- Mehrere Agents bearbeiten dieselben Dateien
- Gleichzeitige Änderungen verhindern
- Externe Dateiänderungen erkennen
- Agent-Abläufe koordinieren

**API:**
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

### 5. Aufgabenwarteschlange

Verwaltet Agent-Aufgaben mit Priorität und Abhängigkeiten.

**Funktionen:**
- Planung nach Priorität
- Abhängigkeiten zwischen Aufgaben
- Verwaltung der Warteschlange
- Statusverfolgung
- Wiederholungslogik

**Konfiguration:**
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

**API:**
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

## Weboberfläche

Das Agent-Dashboard liegt unter `http://localhost:19840/agent`:

### Reiter „Sessions“

- **Aktive Sitzungen** — derzeit laufende Agent-Sitzungen
- **Sitzungsdetails** — Aufgabenfortschritt, Messwerte, Protokolle
- **Sitzungssteuerung** — anhalten, fortsetzen, abbrechen

### Reiter „Tasks“

- **Aufgabenwarteschlange** — wartende und laufende Aufgaben
- **Aufgabenverlauf** — abgeschlossene und fehlgeschlagene Aufgaben
- **Aufgabendetails** — Konfiguration, Protokolle, Ergebnisse

### Reiter „Guardrails“

- **Grenzen für Operationen** — aktuelle Nutzung im Verhältnis zu den Grenzen
- **Blockierte Operationen** — kürzlich blockierte Versuche
- **Freigabewarteschlange** — Operationen, die auf Freigabe warten

### Reiter „Metrics“

- **Token-Verbrauch** — je Sitzung und insgesamt
- **API-Aufrufe** — Anzahl und Rate der Anfragen
- **Dateioperationen** — Lese-, Schreib- und Löschvorgänge
- **Performance** — Latenz und Durchsatz

## Zusammenspiel mit Claude Code

GoZen erkennt Claude-Code-Sitzungen automatisch und stellt ihnen die Agent-Infrastruktur bereit:

```bash
# Start Claude Code with agent support
zen --agent

# Agent features are automatically enabled:
# - Session tracking
# - File coordination
# - Guardrails enforcement
# - Real-time monitoring
```

**Nutzen:**
- Gleichzeitige Dateiänderungen verhindern
- Token-Verbrauch und Kosten verfolgen
- Sicherheitsgrenzen durchsetzen
- Agent-Aktivitäten beobachten
- Abläufe mit mehreren Agents koordinieren

## Anwendungsfälle

### Entwicklung mit mehreren Agents

Mehrere Agents arbeiten an derselben Codebasis:

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

### Lang laufende Aufgaben

Lang laufende Agent-Aufgaben beobachten und steuern:

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

### Sicherheitskritische Operationen

Strenge Sicherheitskontrollen durchsetzen:

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

## Bewährtes Vorgehen

1. **Schutzmechanismen aktivieren** — in der Produktion immer
2. **Passende Grenzen setzen** — an den Anwendungsfall anpassen
3. **Aktiv beobachten** — regelmäßig ins Observatory-Dashboard sehen
4. **Dateisperren nutzen** — den Koordinator für Multi-Agent-Abläufe einschalten
5. **Freigaben einrichten** — für zerstörerische Operationen eine Freigabe verlangen
6. **Protokolle prüfen** — die Aktivitäten der Agents regelmäßig auditieren

## Grenzen

1. **Zusätzlicher Aufwand** — Beobachtung und Koordination kosten Latenz
2. **Dateisperren** — können Multi-Agent-Szenarien verzögern
3. **Speicherverbrauch** — der Sitzungsverlauf belegt Speicher
4. **Komplexität** — setzt Verständnis für Agent-Abläufe voraus
5. **Beta-Status** — Funktionen können sich noch ändern

## Fehlersuche

### Die Agent-Sitzung wird nicht erfasst

1. Prüfen, ob `agent.enabled` `true` ist
2. Prüfen, ob das Observatory aktiviert ist
3. Prüfen, ob der Agent-Client unterstützt wird (Claude Code, Codex)
4. Die Daemon-Protokolle auf Fehler durchsehen

### Probleme mit Dateisperren

1. Prüfen, ob der Koordinator aktiviert ist
2. Prüfen, ob das Sperr-Zeitlimit passend ist
3. Die aktiven Sperren ansehen: `GET /api/v1/agent/locks`
4. Hängende Sperren bei Bedarf manuell lösen

### Die Schutzmechanismen greifen nicht

1. Prüfen, ob sie aktiviert sind
2. Die Regelkonfiguration prüfen
3. Das Protokoll blockierter Operationen ansehen
4. Sicherstellen, dass der Agent-Client die Schutzmechanismen beachtet

### Hoher Speicherverbrauch

1. Die Aufbewahrungsdauer des Verlaufs verkürzen
2. Das Aktualisierungsintervall vergrößern
3. Die Zahl nebenläufiger Aufgaben begrenzen
4. Das automatische Aufräumen einschalten

## Sicherheitsaspekte

1. **Pfadbeschränkungen** — erlaubte und gesperrte Pfade immer konfigurieren
2. **Befehle blockieren** — gefährliche Befehle sperren
3. **Freigabeprozesse** — für heikle Operationen eine Freigabe verlangen
4. **Audit-Protokollierung** — umfassende Protokollierung aktivieren
5. **Ressourcengrenzen** — passende Grenzen für Operationen setzen

## Geplante Erweiterungen

- Protokolle für die Zusammenarbeit mehrerer Agents
- Fortgeschrittene Strategien zur Konfliktlösung
- Maschinelles Lernen zur Anomalieerkennung
- Anbindung an externe Überwachungswerkzeuge
- Analysen zum Verhalten der Agents
- Automatisch erzeugte Sicherheitsrichtlinien
