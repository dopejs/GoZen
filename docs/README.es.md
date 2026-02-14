# GoZen

[English](../README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md)

> **Go Zen** — entra en un estado de flujo zen para programar. **Goes Env** — cambio de entorno fluido.

Conmutador de entornos multi-CLI para Claude Code, Codex y OpenCode con conmutación automática por fallos en el proxy de API.

## Características

- **Soporte multi-CLI** — Compatible con Claude Code, Codex y OpenCode, configurable por proyecto
- **Gestión multi-configuración** — Gestiona todas las configuraciones de API en `~/.zen/zen.json`
- **Conmutación por fallos del proxy** — Proxy HTTP integrado que cambia automáticamente a proveedores de respaldo cuando el principal no está disponible
- **Enrutamiento por escenarios** — Enrutamiento inteligente basado en características de la solicitud (thinking, image, longContext, etc.)
- **Vinculación de proyectos** — Vincula directorios a perfiles y CLIs específicos para configuración automática por proyecto
- **Variables de entorno** — Configura variables de entorno específicas por CLI a nivel de proveedor
- **Interfaz TUI** — Interfaz de terminal interactiva con modos Dashboard y legado
- **Interfaz web de gestión** — Gestión visual desde el navegador para proveedores, perfiles y vinculaciones de proyectos
- **Verificación de actualizaciones** — Verificación automática no bloqueante de nuevas versiones al iniciar (caché de 24h)
- **Autoactualización** — Actualización con un solo comando vía `zen upgrade` con coincidencia de versiones semver
- **Autocompletado de Shell** — Compatible con zsh / bash / fish

## Instalación

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

Desinstalar:

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh -s -- --uninstall
```

## Inicio rápido

```sh
# Abrir la interfaz TUI y crear el primer proveedor
zen config

# Iniciar (usando el perfil predeterminado)
zen

# Usar un perfil específico
zen -p work

# Usar un CLI específico
zen --cli codex
```

## Referencia de comandos

| Comando | Descripción |
|---------|-------------|
| `zen` | Iniciar CLI (usando vinculación de proyecto o configuración predeterminada) |
| `zen -p <profile>` | Iniciar con un perfil específico |
| `zen -p` | Seleccionar perfil interactivamente |
| `zen --cli <cli>` | Usar un CLI específico (claude/codex/opencode) |
| `zen use <provider>` | Usar directamente un proveedor específico (sin proxy) |
| `zen pick` | Seleccionar interactivamente un proveedor para iniciar |
| `zen list` | Listar todos los proveedores y perfiles |
| `zen config` | Abrir la interfaz TUI de configuración |
| `zen config --legacy` | Usar la interfaz TUI legada |
| `zen bind <profile>` | Vincular el directorio actual a un perfil |
| `zen bind --cli <cli>` | Vincular el directorio actual a un CLI específico |
| `zen unbind` | Eliminar la vinculación del directorio actual |
| `zen status` | Mostrar el estado de vinculación del directorio actual |
| `zen web start` | Iniciar la interfaz web de gestión |
| `zen web open` | Abrir la interfaz web en el navegador |
| `zen web stop` | Detener el servidor web |
| `zen web restart` | Reiniciar el servidor web |
| `zen upgrade` | Actualizar a la última versión |
| `zen version` | Mostrar versión |

## Soporte multi-CLI

zen es compatible con tres CLIs de asistentes de programación con IA:

| CLI | Descripción | Formato de API |
|-----|-------------|----------------|
| `claude` | Claude Code (predeterminado) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

### Establecer CLI predeterminado

```sh
# Vía TUI
zen config  # Settings → Default CLI

# Vía Web UI
zen web open  # Página de Settings
```

### CLI por proyecto

```sh
cd ~/work/project
zen bind --cli codex  # Usar Codex para este directorio
```

### Usar otro CLI temporalmente

```sh
zen --cli opencode  # Usar OpenCode para esta sesión
```

## Gestión de perfiles

Un perfil es una lista ordenada de proveedores utilizada para conmutación por fallos.

### Ejemplo de configuración

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-main", "anthropic-backup"]
    },
    "work": {
      "providers": ["company-api"],
      "routing": {
        "think": {"providers": [{"name": "thinking-api"}]}
      }
    }
  }
}
```

### Uso de perfiles

```sh
# Usar perfil predeterminado
zen

# Usar un perfil específico
zen -p work

# Selección interactiva
zen -p
```

## Vinculación de proyectos

Vincula directorios a perfiles y/o CLIs específicos para configuración automática por proyecto.

```sh
cd ~/work/company-project

# Vincular perfil
zen bind work-profile

# Vincular CLI
zen bind --cli codex

# Vincular ambos
zen bind work-profile --cli codex

# Ver estado
zen status

# Eliminar vinculación
zen unbind
```

**Prioridad**: Argumentos de línea de comandos > Vinculación de proyecto > Predeterminado global

## Interfaz TUI de configuración

```sh
zen config
```

v1.5 introduce una nueva interfaz Dashboard:

- **Panel izquierdo**: Proveedores, Perfiles, Vinculaciones de proyectos
- **Panel derecho**: Detalles del elemento seleccionado
- **Atajos de teclado**:
  - `a` - Añadir nuevo elemento
  - `e` - Editar elemento seleccionado
  - `d` - Eliminar elemento seleccionado
  - `Tab` - Cambiar foco
  - `q` - Volver / Salir

Usa `--legacy` para cambiar a la interfaz legada.

## Interfaz web de gestión

```sh
# Iniciar (se ejecuta en segundo plano, puerto 19840)
zen web start

# Abrir en el navegador
zen web open

# Detener
zen web stop

# Reiniciar
zen web restart
```

Funcionalidades de la interfaz web:
- Gestión de proveedores y perfiles
- Gestión de vinculaciones de proyectos
- Configuración global (CLI predeterminado, perfil predeterminado, puerto)
- Visor de registros de solicitudes
- Autocompletado del campo de modelo

## Variables de entorno

Cada proveedor puede tener variables de entorno específicas por CLI:

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
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

### Variables de entorno comunes de Claude Code

| Variable | Descripción |
|----------|-------------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | Tokens de salida máximos |
| `MAX_THINKING_TOKENS` | Presupuesto de pensamiento extendido |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | Ventana de contexto máxima |
| `BASH_DEFAULT_TIMEOUT_MS` | Tiempo de espera predeterminado de Bash |

## Enrutamiento por escenarios

Enruta automáticamente las solicitudes a diferentes proveedores según las características de la solicitud:

| Escenario | Condición de activación |
|-----------|------------------------|
| `think` | Modo thinking activado |
| `image` | Contiene contenido de imagen |
| `longContext` | El contenido supera el umbral |
| `webSearch` | Usa la herramienta web_search |
| `background` | Usa el modelo Haiku |

**Mecanismo de fallback**: Si todos los proveedores en la configuración del escenario fallan, se recurre automáticamente a los proveedores predeterminados del perfil.

Ejemplo de configuración:

```json
{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}
```

## Archivos de configuración

| Archivo | Descripción |
|---------|-------------|
| `~/.zen/zen.json` | Archivo de configuración principal |
| `~/.zen/proxy.log` | Registro del proxy |
| `~/.zen/web.log` | Registro del servidor web |

### Ejemplo de configuración completa

```json
{
  "version": 6,
  "default_profile": "default",
  "default_cli": "claude",
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
      "cli": "codex"
    }
  }
}
```

## Actualización

```sh
# Última versión
zen upgrade

# Versión específica
zen upgrade 2.1
zen upgrade 2.1.0
```

## Migración desde versiones anteriores

GoZen migra automáticamente las configuraciones de versiones anteriores:
- `~/.opencc/opencc.json` → `~/.zen/zen.json` (desde OpenCC v1.x)
- `~/.cc_envs/` → `~/.zen/zen.json` (desde formato legado)

## Desarrollo

```sh
# Compilar
go build -o zen .

# Probar
go test ./...
```

Publicación: Empuja un tag y GitHub Actions compilará automáticamente.

```sh
git tag v2.0.0
git push origin v2.0.0
```

## License

MIT
