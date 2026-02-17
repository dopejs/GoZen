---
sidebar_position: 8
title: Referencia de Configuración
---

# Referencia de Configuración

## Ubicación de Archivos

| Archivo | Descripción |
|---------|-------------|
| `~/.zen/zen.json` | Archivo de configuración principal |
| `~/.zen/zend.log` | Registro del daemon |
| `~/.zen/zend.pid` | Archivo PID del daemon |
| `~/.zen/logs.db` | Base de datos de registros de solicitudes (SQLite) |

## Ejemplo de Configuración Completo

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

## Referencia de Campos

| Campo | Descripción |
|-------|-------------|
| `version` | Número de versión del archivo de configuración |
| `default_profile` | Nombre del perfil predeterminado |
| `default_client` | Cliente CLI predeterminado (claude/codex/opencode) |
| `proxy_port` | Puerto del servidor proxy (predeterminado: 19841) |
| `web_port` | Puerto de la interfaz de gestión Web (predeterminado: 19840) |
| `providers` | Colección de configuraciones de proveedores |
| `profiles` | Colección de configuraciones de perfiles |
| `project_bindings` | Configuración de vinculaciones de proyecto |
| `sync` | Ajustes de sincronización de configuración (opcional) |
