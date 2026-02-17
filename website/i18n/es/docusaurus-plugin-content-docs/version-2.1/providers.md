---
sidebar_position: 2
title: Gestión de Proveedores
---

# Gestión de Proveedores

Un proveedor representa una configuración de endpoint API que incluye URL base, token de autenticación, nombre de modelo y más.

## Ejemplo de Configuración

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

## Variables de Entorno

Cada proveedor puede tener variables de entorno por CLI:

### Variables de Entorno Comunes de Claude Code

| Variable | Descripción |
|----------|-------------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | Tokens de salida máximos |
| `MAX_THINKING_TOKENS` | Presupuesto de pensamiento extendido |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | Ventana de contexto máxima |
| `BASH_DEFAULT_TIMEOUT_MS` | Tiempo de espera predeterminado de Bash |
