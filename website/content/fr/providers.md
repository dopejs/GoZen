---
sidebar_position: 2
title: Fournisseurs
---

# Gestion des fournisseurs

Un fournisseur représente la configuration d’un point d’accès API : URL de base, jeton d’authentification, nom du modèle, etc.

## Exemple de configuration

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

## Variables d’environnement

Chaque fournisseur peut définir des variables d’environnement par CLI :

### Variables d’environnement courantes de Claude Code

| Variable | Description |
|----------|-------------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | Nombre maximal de jetons en sortie |
| `MAX_THINKING_TOKENS` | Budget de raisonnement étendu |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | Fenêtre de contexte maximale |
| `BASH_DEFAULT_TIMEOUT_MS` | Délai d’expiration par défaut de Bash |
