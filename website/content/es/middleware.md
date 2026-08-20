---
sidebar_position: 15
title: Pipeline de Middleware (BETA)
---

# Pipeline de Middleware (BETA)

:::warning FUNCIÓN BETA
El pipeline de middleware está actualmente en beta. Está deshabilitado por defecto y requiere configuración explícita para habilitarlo.
:::

Extiende GoZen con middleware conectable para transformación de solicitudes/respuestas, registro, límite de tasa y procesamiento personalizado.

## Características

- **Arquitectura Conectable** — Agrega lógica de procesamiento personalizada sin modificar el código central
- **Ejecución Basada en Prioridad** — Controla el orden de ejecución del middleware
- **Hooks de Solicitud/Respuesta** — Procesa solicitudes antes de enviar, respuestas después de recibir
- **Middleware Integrado** — Inyección de contexto, registro, límite de tasa, compresión
- **Cargador de Plugins** — Carga middleware desde archivos locales o URLs remotas
- **Manejo de Errores** — Manejo de errores elegante y comportamiento de respaldo

## Arquitectura

```
Solicitud del Cliente
    ↓
[Middleware 1: Prioridad 100]
    ↓
[Middleware 2: Prioridad 200]
    ↓
[Middleware 3: Prioridad 300]
    ↓
API del Proveedor
    ↓
[Middleware 3: Respuesta]
    ↓
[Middleware 2: Respuesta]
    ↓
[Middleware 1: Respuesta]
    ↓
Respuesta al Cliente
```

## Configuración

### Habilitar Pipeline de Middleware

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "context-injection",
        "enabled": true,
        "priority": 100,
        "config": {}
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info"
        }
      }
    ]
  }
}
```

**Opciones:**

| Opción | Descripción |
|------|------|
| `enabled` | Habilita el pipeline de middleware |
| `pipeline` | Array de configuraciones de middleware |
| `name` | Identificador del middleware |
| `priority` | Orden de ejecución (menor = más temprano) |
| `config` | Configuración específica del middleware |

## Middleware Integrado

### 1. Inyección de Contexto

Inyecta contexto personalizado en las solicitudes.

```json
{
  "name": "context-injection",
  "enabled": true,
  "priority": 100,
  "config": {
    "system_prompt": "Eres un asistente de codificación útil.",
    "metadata": {
      "session_id": "sess_123",
      "user_id": "user_456"
    }
  }
}
```

**Casos de Uso:**
- Agregar prompts del sistema
- Inyectar metadatos de sesión
- Agregar contexto de usuario

### 2. Registrador de Solicitudes

Registra todas las solicitudes y respuestas.

```json
{
  "name": "request-logger",
  "enabled": true,
  "priority": 200,
  "config": {
    "log_level": "info",
    "log_body": false,
    "log_headers": true
  }
}
```

**Casos de Uso:**
- Depuración
- Pista de auditoría
- Monitoreo de rendimiento

### 3. Limitador de Tasa

Limita la tasa de solicitudes por proveedor o globalmente.

```json
{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60,
    "burst": 10,
    "per_provider": true
  }
}
```

**Casos de Uso:**
- Prevenir errores de límite de tasa
- Controlar uso de API
- Prevenir abuso

### 4. Compresión (BETA)

Comprime el contexto cuando el conteo de tokens excede el umbral.

```json
{
  "name": "compression",
  "enabled": true,
  "priority": 400,
  "config": {
    "threshold_tokens": 50000,
    "target_tokens": 20000
  }
}
```

Ver [Compresión de Contexto](./compression.md) para detalles.

### 5. Memoria de Sesión (BETA)

Mantiene memoria de conversación entre sesiones.

```json
{
  "name": "session-memory",
  "enabled": true,
  "priority": 150,
  "config": {
    "max_memories": 100,
    "ttl_hours": 24,
    "storage": "sqlite"
  }
}
```

**Casos de Uso:**
- Recordar preferencias de usuario
- Rastrear historial de conversación
- Mantener contexto entre sesiones

### 6. Orquestación (BETA)

Enruta solicitudes a múltiples proveedores y agrega respuestas.

```json
{
  "name": "orchestration",
  "enabled": true,
  "priority": 500,
  "config": {
    "strategy": "parallel",
    "providers": ["anthropic", "openai"],
    "consensus": "longest"
  }
}
```

**Casos de Uso:**
- Comparar salidas de modelos
- Redundancia para solicitudes críticas
- Mejorar calidad mediante consenso

## Middleware Personalizado

### Interfaz de Middleware

```go
type Middleware interface {
    Name() string
    Priority() int
    ProcessRequest(ctx *RequestContext) error
    ProcessResponse(ctx *ResponseContext) error
}

type RequestContext struct {
    Provider  string
    Model     string
    Messages  []Message
    Metadata  map[string]interface{}
}

type ResponseContext struct {
    Provider  string
    Model     string
    Response  *APIResponse
    Latency   time.Duration
    Metadata  map[string]interface{}
}
```

### Ejemplo: Inyección de Encabezados Personalizados

```go
package main

import (
    "github.com/dopejs/gozen/internal/middleware"
)

type CustomHeaderMiddleware struct {
    headers map[string]string
}

func (m *CustomHeaderMiddleware) Name() string {
    return "custom-headers"
}

func (m *CustomHeaderMiddleware) Priority() int {
    return 250
}

func (m *CustomHeaderMiddleware) ProcessRequest(ctx *middleware.RequestContext) error {
    for k, v := range m.headers {
        ctx.Metadata[k] = v
    }
    return nil
}

func (m *CustomHeaderMiddleware) ProcessResponse(ctx *middleware.ResponseContext) error {
    // No se necesita procesamiento de respuesta
    return nil
}

func init() {
    middleware.Register("custom-headers", func(config map[string]interface{}) middleware.Middleware {
        return &CustomHeaderMiddleware{
            headers: config["headers"].(map[string]string),
        }
    })
}
```

### Cargar Middleware Personalizado

#### Plugin Local

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "local",
        "path": "/path/to/custom-middleware.so",
        "config": {
          "headers": {
            "X-Custom-Header": "value"
          }
        }
      }
    ]
  }
}
```

#### Plugin Remoto

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "remote",
        "url": "https://example.com/middleware/custom-headers.so",
        "checksum": "sha256:abc123...",
        "config": {}
      }
    ]
  }
}
```

## Web UI

Accede a la configuración de middleware en `http://localhost:19840/settings`:

1. Navega a la pestaña "Middleware" (marcada con insignia BETA)
2. Activa "Enable Middleware Pipeline"
3. Agrega/elimina middleware del pipeline
4. Ajusta prioridades y configuración
5. Habilita/deshabilita middleware individual
6. Haz clic en "Save"

## Endpoints de API

### Listar Middleware

```bash
GET /api/v1/middleware
```

Respuesta:
```json
{
  "enabled": true,
  "pipeline": [
    {
      "name": "context-injection",
      "enabled": true,
      "priority": 100,
      "type": "builtin"
    },
    {
      "name": "request-logger",
      "enabled": true,
      "priority": 200,
      "type": "builtin"
    }
  ]
}
```

### Agregar Middleware

```bash
POST /api/v1/middleware
Content-Type: application/json

{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60
  }
}
```

### Actualizar Middleware

```bash
PUT /api/v1/middleware/{name}
Content-Type: application/json

{
  "enabled": false
}
```

### Eliminar Middleware

```bash
DELETE /api/v1/middleware/{name}
```

### Recargar Pipeline

```bash
POST /api/v1/middleware/reload
```

## Casos de Uso

### Entorno de Desarrollo

Agrega registro de depuración e inspección de solicitudes:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 100,
        "config": {
          "log_level": "debug",
          "log_body": true
        }
      }
    ]
  }
}
```

### Entorno de Producción

Agrega límite de tasa y monitoreo:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "rate-limiter",
        "enabled": true,
        "priority": 100,
        "config": {
          "requests_per_minute": 100,
          "burst": 20
        }
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info",
          "log_body": false
        }
      }
    ]
  }
}
```

### Comparación Multi-Proveedor

Usa orquestación para comparar salidas:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "orchestration",
        "enabled": true,
        "priority": 500,
        "config": {
          "strategy": "parallel",
          "providers": ["anthropic", "openai", "google"],
          "consensus": "longest"
        }
      }
    ]
  }
}
```

## Mejores Prácticas

1. **Usar Prioridades Apropiadas** — Los números más bajos se ejecutan primero
2. **Mantener el Middleware Enfocado** — Cada middleware debe hacer una cosa bien
3. **Manejar Errores Elegantemente** — No romper el pipeline por errores
4. **Probar Exhaustivamente** — Validar el comportamiento del middleware antes de producción
5. **Monitorear Rendimiento** — Rastrear la sobrecarga del middleware
6. **Documentar Configuración** — Documentar claramente las opciones de configuración

## Limitaciones

1. **Sobrecarga de Rendimiento** — Cada middleware agrega latencia
2. **Complejidad** — Demasiado middleware dificulta la depuración
3. **Seguridad de Plugins** — Los plugins remotos requieren confianza y verificación
4. **Propagación de Errores** — Los errores de middleware afectan todas las solicitudes
5. **Complejidad de Configuración** — Los pipelines complejos son más difíciles de mantener

## Solución de Problemas

### El Middleware No Se Ejecuta

1. Verifica que `middleware.enabled` sea `true`
2. Revisa que el middleware esté habilitado en el pipeline
3. Valida que la prioridad esté configurada correctamente
4. Revisa los logs del daemon para errores de middleware

### Comportamiento Inesperado

1. Revisa el orden de ejecución del middleware (prioridad)
2. Valida que la configuración sea correcta
3. Prueba el middleware de forma aislada
4. Revisa los logs del middleware

### Problemas de Rendimiento

1. Identifica middleware lento (revisa logs)
2. Reduce el número de middleware
3. Optimiza la implementación del middleware
4. Considera deshabilitar middleware no esencial

### Fallo de Carga de Plugin

1. Verifica que la ruta del plugin sea correcta
2. Revisa que el plugin esté compilado para la arquitectura correcta
3. Valida que el checksum coincida (para plugins remotos)
4. Revisa los logs del plugin para errores

## Consideraciones de Seguridad

1. **Validar Plugins** — Solo carga plugins de confianza
2. **Verificar Checksums** — Siempre verifica checksums de plugins remotos
3. **Sandbox de Plugins** — Considera ejecutar plugins en entornos aislados
4. **Auditar Middleware** — Revisa el código del middleware antes de desplegar
5. **Monitorear Comportamiento** — Observa comportamiento inesperado del middleware

## Mejoras Futuras

- Soporte de plugins WebAssembly para compatibilidad multiplataforma
- Marketplace de middleware para compartir plugins de la comunidad
- Editor visual de pipeline en la Web UI
- Perfilado de rendimiento de middleware
- Recarga en caliente para actualizaciones de plugins
- Framework de pruebas de middleware
