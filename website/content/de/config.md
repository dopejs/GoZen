---
sidebar_position: 8
title: Konfigurationsreferenz
---

# Konfigurationsreferenz

## Dateiorte

| Datei | Beschreibung |
|------|-------|
| `~/.zen/zen.json` | Hauptkonfigurationsdatei |
| `~/.zen/zend.log` | Daemon-Protokoll |
| `~/.zen/zend.pid` | PID-Datei des Daemons |
| `~/.zen/logs.db` | Datenbank der Anfrageprotokolle (SQLite) |

## Vollständiges Konfigurationsbeispiel

```json
{
  "version": 7,
  "default_profile": "default",
  "default_client": "claude",
  "proxy_port": 19841,
  "web_port": 19840,
  "providers": {
    "anthropic": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000"
      }
    }
  },
  "profiles": {
    "default": {
      "providers": ["anthropic"]
    }
  },
  "project_bindings": {
    "/path/to/project": {
      "profile": "work",
      "client": "codex"
    }
  }
}
```

## Feldreferenz

| Feld | Beschreibung |
|-------|-------|
| `version` | Versionsnummer der Konfigurationsdatei |
| `default_profile` | Name des Standardprofils |
| `default_client` | Standard-CLI-Client (claude/codex/opencode) |
| `proxy_port` | Port des Proxy-Servers (Standard: 19841) |
| `web_port` | Port der Weboberfläche (Standard: 19840) |
| `providers` | Sammlung der Anbieterkonfigurationen |
| `profiles` | Sammlung der Profilkonfigurationen |
| `project_bindings` | Konfiguration der Projektbindungen |
| `sync` | Einstellungen der Konfigurationssynchronisierung (optional) |
