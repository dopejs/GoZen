---
sidebar_position: 8
title: Référence de configuration
---

# Référence de configuration

## Emplacements des fichiers

| Fichier | Description |
|------|-------|
| `~/.zen/zen.json` | Fichier de configuration principal |
| `~/.zen/zend.log` | Journal du daemon |
| `~/.zen/zend.pid` | Fichier PID du daemon |
| `~/.zen/logs.db` | Base de données des journaux de requêtes (SQLite) |

## Exemple de configuration complète

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

## Référence des champs

| Champ | Description |
|-------|-------|
| `version` | Numéro de version du fichier de configuration |
| `default_profile` | Nom du profil par défaut |
| `default_client` | Client CLI par défaut (claude/codex/opencode) |
| `proxy_port` | Port du serveur proxy (par défaut : 19841) |
| `web_port` | Port de l’interface web de gestion (par défaut : 19840) |
| `providers` | Ensemble des configurations de fournisseurs |
| `profiles` | Ensemble des configurations de profils |
| `project_bindings` | Configuration des liaisons de projet |
| `sync` | Réglages de synchronisation de la configuration (facultatif) |
