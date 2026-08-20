---
sidebar_position: 2
title: Anbieter
---

# Anbieterverwaltung

Ein Anbieter beschreibt die Konfiguration eines API-Endpunkts: Basis-URL, Auth-Token, Modellname und mehr.

## Beispielkonfiguration

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000"
      },
      "codex_env_vars": {
        "CODEX_SOME_VAR": "value"
      },
      "opencode_env_vars": {
        "OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": "64000"
      }
    }
  }
}
```

## Umgebungsvariablen

Jeder Anbieter kann Umgebungsvariablen je CLI festlegen:

### Häufige Umgebungsvariablen von Claude Code

| Variable | Beschreibung |
|----------|--------------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | Maximale Ausgabe-Tokens |
| `MAX_THINKING_TOKENS` | Budget für erweitertes Nachdenken |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | Maximales Kontextfenster |
| `BASH_DEFAULT_TIMEOUT_MS` | Standard-Timeout für Bash |
