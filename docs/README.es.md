# GoZen

<p align="center">
  <img src="https://raw.githubusercontent.com/dopejs/GoZen/main/assets/gozen.svg" alt="GoZen Logo" width="120">
</p>

[English](../README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md)

> **Go Zen** — entra en un estado de flujo zen para programar. **Goes Env** — cambio de entorno fluido.

Conmutador de entornos multi-CLI para Claude Code, Codex y OpenCode con conmutación automática por fallos en el proxy de API.

## Características

- **Soporte multi-CLI** — Compatible con Claude Code, Codex y OpenCode, configurable por proyecto
- **Gestión multi-configuración** — Gestiona todas las configuraciones de API en `~/.zen/zen.json`
- **Daemon unificado** — Un único proceso `zend` aloja tanto el servidor proxy como la interfaz web
- **Conmutación por fallos del proxy** — Proxy HTTP integrado que cambia automáticamente a proveedores de respaldo cuando el principal no está disponible
- **Enrutamiento por escenarios** — Enrutamiento inteligente basado en características de la solicitud (thinking, image, longContext, etc.)
- **Vinculación de proyectos** — Vincula directorios a perfiles y CLIs específicos para configuración automática por proyecto
- **Variables de entorno** — Configura variables de entorno específicas por CLI a nivel de proveedor
- **Interfaz web de gestión** — Gestión visual desde el navegador con acceso protegido por contraseña
- **Seguridad web** — Contraseña de acceso autogenerada, autenticación basada en sesiones, cifrado RSA para transporte de tokens
- **Sincronización de configuración** — Sincroniza proveedores, perfiles y ajustes entre dispositivos vía WebDAV, S3, GitHub Gist o GitHub Repo con cifrado AES-256-GCM
- **Verificación de actualizaciones** — Verificación automática no bloqueante de nuevas versiones al iniciar (caché de 24h)
- **Autoactualización** — Actualización con un solo comando vía `zen upgrade` con coincidencia de versiones semver (compatible con versiones preliminares)
- **Autocompletado de Shell** — Compatible con zsh / bash / fish

## Instalación

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

Desinstalar:

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh
```

## Inicio rápido

```sh
# Añadir el primer proveedor
zen config add provider

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
| `zen -y` / `zen --yes` | Aprobar automáticamente permisos CLI (claude `--permission-mode acceptEdits`, codex `-a never`) |
| `zen use <provider>` | Usar directamente un proveedor específico (sin proxy) |
| `zen pick` | Seleccionar interactivamente un proveedor para iniciar |
| `zen list` | Listar todos los proveedores y perfiles |
| `zen config` | Mostrar subcomandos de configuración |
| `zen config add provider` | Añadir un nuevo proveedor |
| `zen config add profile` | Añadir un nuevo perfil |
| `zen config default-client` | Establecer el cliente CLI predeterminado |
| `zen config default-profile` | Establecer el perfil predeterminado |
| `zen config reset-password` | Restablecer la contraseña de acceso a la interfaz web |
| `zen config sync` | Obtener configuración del backend de sincronización remoto |
| `zen daemon start` | Iniciar el daemon zend |
| `zen daemon stop` | Detener el daemon |
| `zen daemon restart` | Reiniciar el daemon |
| `zen daemon status` | Mostrar estado del daemon |
| `zen daemon enable` | Instalar el daemon como servicio del sistema |
| `zen daemon disable` | Desinstalar el servicio del sistema |
| `zen bind <profile>` | Vincular el directorio actual a un perfil |
| `zen bind --cli <cli>` | Vincular el directorio actual a un CLI específico |
| `zen unbind` | Eliminar la vinculación del directorio actual |
| `zen status` | Mostrar el estado de vinculación del directorio actual |
| `zen web` | Abrir la interfaz web de gestión en el navegador |
| `zen upgrade` | Actualizar a la última versión |
| `zen version` | Mostrar versión |

## Arquitectura del daemon

En v2.1, GoZen utiliza un proceso daemon unificado (`zend`) que aloja tanto el proxy HTTP como la interfaz web:

- **Servidor proxy** ejecutándose en el puerto `19841` (configurable vía `proxy_port`)
- **Interfaz web** ejecutándose en el puerto `19840` (configurable vía `web_port`)
- El daemon se inicia automáticamente al ejecutar `zen` o `zen web`
- Los cambios de configuración se recargan en caliente mediante monitoreo de archivos
- El auto-push de sincronización (con antirrebote de 2s) y el auto-pull son gestionados por el daemon

```sh
# Gestión manual del daemon
zen daemon start          # Iniciar el daemon
zen daemon stop           # Detener el daemon
zen daemon restart        # Reiniciar el daemon
zen daemon status         # Ver estado del daemon

# Servicio del sistema (inicio automático al arrancar)
zen daemon enable         # Instalar como servicio del sistema
zen daemon disable        # Eliminar servicio del sistema
```

## Soporte multi-CLI

zen es compatible con tres CLIs de asistentes de programación con IA:

| CLI | Descripción | Formato de API |
|-----|-------------|----------------|
| `claude` | Claude Code (predeterminado) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

### Establecer CLI predeterminado

```sh
zen config default-client

# Vía Web UI
zen web  # Página de Settings
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

## Interfaz web de gestión

```sh
# Abrir en el navegador (inicia el daemon automáticamente si es necesario)
zen web
```

Funcionalidades de la interfaz web:
- Gestión de proveedores y perfiles
- Gestión de vinculaciones de proyectos
- Configuración global (cliente predeterminado, perfil predeterminado, puertos)
- Configuración de sincronización
- Visor de registros de solicitudes con actualización automática
- Autocompletado del campo de modelo

### Seguridad de la interfaz web

Al iniciar el daemon por primera vez, se genera automáticamente una contraseña de acceso. Las solicitudes no locales (fuera de 127.0.0.1/::1) requieren inicio de sesión.

- **Autenticación basada en sesiones** con expiración configurable
- **Protección contra fuerza bruta** con retroceso exponencial
- **Cifrado RSA** para transporte de tokens sensibles (las claves API se cifran en el navegador antes de enviar)
- El acceso local (127.0.0.1) no requiere autenticación

```sh
# Restablecer la contraseña de la interfaz web
zen config reset-password

# Cambiar contraseña vía Web UI
zen web  # Settings → Change Password
```

## Sincronización de configuración

Sincroniza proveedores, perfiles, perfil predeterminado y cliente predeterminado entre dispositivos. Los tokens de autenticación se cifran con AES-256-GCM (derivación de clave PBKDF2-SHA256) antes de subir.

Backends compatibles:
- **WebDAV** — Cualquier servidor WebDAV (ej. Nextcloud, ownCloud)
- **S3** — AWS S3 o almacenamiento compatible con S3 (ej. MinIO, Cloudflare R2)
- **GitHub Gist** — Gist privado (requiere PAT con alcance `gist`)
- **GitHub Repo** — Archivo en repositorio vía Contents API (requiere PAT con alcance `repo`)

### Configuración vía Web UI

```sh
zen web  # Settings → Config Sync
```

### Pull manual vía CLI

```sh
zen config sync
```

### Resolución de conflictos

- Fusión por marca de tiempo por entidad: la modificación más reciente gana
- Las entidades eliminadas usan lápidas (expiran después de 30 días)
- Escalares (perfil/cliente predeterminado): la marca de tiempo más reciente gana

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
| `~/.zen/zend.log` | Registro del daemon |
| `~/.zen/zend.pid` | Archivo PID del daemon |
| `~/.zen/logs.db` | Base de datos de registros de solicitudes (SQLite) |

### Ejemplo de configuración completa

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

## Actualización

```sh
# Última versión
zen upgrade

# Versión específica
zen upgrade 2.1
zen upgrade 2.1.0

# Versión preliminar
zen upgrade 2.1.0-alpha.1
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
git tag v2.1.0
git push origin v2.1.0
```

## License

MIT
