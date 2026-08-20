---
title: Lastverteilung
---

# Lastverteilung

GoZen bietet über das einfache Failover hinaus mehrere Strategien zur Anbieterauswahl. Du wählst eine Strategie je Profil, kombinierst sie mit Health-Checks und steuerst den Verkehr nach Verfügbarkeit, Latenz oder Kosten.

## Verfügbare Strategien

### Failover

Probiert die Anbieter der Reihe nach, bis einer antwortet. Das ist die Standardstrategie und passt gut zu Primär-/Backup-Aufbauten.

```json
{
  "profiles": {
    "default": {
      "providers": ["primary", "backup"],
      "strategy": "failover"
    }
  }
}
```

### Round Robin

Verteilt Anfragen gleichmäßig auf mehrere gleichwertige Anbieter.

```json
{
  "profiles": {
    "balanced": {
      "providers": ["provider-a", "provider-b", "provider-c"],
      "strategy": "round-robin"
    }
  }
}
```

### Geringste Latenz

Bevorzugt den Anbieter mit der niedrigsten jüngsten Antwortzeit.

```json
{
  "profiles": {
    "fast": {
      "providers": ["us-east", "us-west", "eu"],
      "strategy": "least-latency"
    }
  }
}
```

### Geringste Kosten

Bevorzugt den günstigsten Anbieter für das angefragte Modell.

```json
{
  "profiles": {
    "budget": {
      "providers": ["cheap-provider", "premium-provider"],
      "strategy": "least-cost"
    }
  }
}
```

## Health-bewusstes Routing

Alle Strategien lassen sich mit dem Health-Monitoring kombinieren. Ist `health_aware` aktiv, werden ungesunde Anbieter automatisch übersprungen, bis sie sich erholt haben.

```json
{
  "profiles": {
    "production": {
      "providers": ["primary", "secondary", "tertiary"],
      "strategy": "least-latency",
      "health_aware": true
    }
  }
}
```

## Die richtige Strategie wählen

- Nimm `failover`, wenn Zuverlässigkeit an erster Stelle steht.
- Nimm `round-robin`, wenn die Anbieter austauschbar sind.
- Nimm `least-latency` für interaktive oder zeitkritische Lasten.
- Nimm `least-cost`, wenn das Budget wichtiger ist als reine Geschwindigkeit.

## Verwandte Dokumente

- [Profile](/docs/profiles) erklärt, wie Anbietergruppen definiert werden.
- [Routing](/docs/routing) behandelt die szenariobasierte Anbieterauswahl.
- [Health-Monitoring](/docs/health-monitoring) erklärt, wie Health-Checks das Routing beeinflussen.
