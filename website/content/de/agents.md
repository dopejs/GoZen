---
title: Agents
---

# Agents

GoZen kann als Betriebsschicht für Coding-Agents wie Claude Code, Codex und andere CLI-gesteuerte Assistenten dienen. Es hilft, deren Arbeit zu koordinieren, Sitzungen zu beobachten und Schutzmechanismen zur Laufzeit durchzusetzen, ohne den bestehenden Arbeitsablauf zu ändern.

## Was GoZen beisteuert

- **Koordination**: weniger Konflikte, wenn mehrere Agents am selben Projekt arbeiten.
- **Beobachtbarkeit**: Sitzungen, Kosten, Fehler und Aktivität an einer Stelle verfolgen.
- **Schutzmechanismen**: Grenzen für Ausgaben, Anfragerate und heikle Aktionen setzen.
- **Aufgaben-Routing**: unterschiedliche Arbeit an unterschiedliche Anbieter oder Profile schicken.

## Beispielkonfiguration

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

## Typische Abläufe

### Koordination mehrerer Agents

Arbeiten mehrere Agents im selben Repository, verfolgt GoZen die Dateiaktivität, meldet Warnungen und macht Kollisionen leichter vermeidbar.

### Sitzungsüberwachung

Über das Dashboard und die APIs lassen sich aktive Sitzungen, Token-Verbrauch, Fehlerzahlen und Laufzeit einsehen.

### Durchsetzen der Schutzmechanismen

Schutzmechanismen können außer Kontrolle geratene Sitzungen anhalten, riskante Operationen markieren und Wiederholungsschleifen bremsen, bevor sie teuer werden.

## Verwandte Dokumente

- [Agent-Infrastruktur](/docs/agent-infrastructure) beschreibt die neuere Architektur aus Runtime, Observatory, Koordinator und Schutzmechanismen im Detail.
- [Bot-Gateway](/docs/bot) erklärt, wie laufende Sitzungen aus Telegram, Slack, Discord und anderen Chat-Plattformen gesteuert werden.
- [Nutzungserfassung](/docs/usage-tracking) und [Health-Monitoring](/docs/health-monitoring) beschreiben die Messwerte hinter dem Agent-Betrieb.
