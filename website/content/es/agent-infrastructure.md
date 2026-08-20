---
sidebar_position: 16
title: Infraestructura de Agentes (BETA)
---

# Infraestructura de Agentes (BETA)

:::warning FUNCIÓN BETA
La infraestructura de agentes está actualmente en beta. Está deshabilitada por defecto y requiere configuración explícita para habilitarla.
:::

Soporte integrado para flujos de trabajo de agentes autónomos, incluyendo gestión de sesiones, coordinación de archivos, monitoreo en tiempo real y controles de seguridad.

## Características

- **Runtime de Agentes** — Ejecuta tareas de agentes autónomos con gestión completa del ciclo de vida
- **Observatorio** — Monitoreo en tiempo real de sesiones y actividades de agentes
- **Barandillas** — Controles de seguridad y restricciones para el comportamiento de agentes
- **Coordinador** — Coordinación de flujos de trabajo multi-agente basada en archivos
- **Cola de Tareas** — Gestiona tareas de agentes con prioridades y dependencias
- **Gestión de Sesiones** — Rastrea sesiones de agentes a través de múltiples proyectos

## Arquitectura

```
Cliente de Agente (Claude Code, Codex, etc.)
    ↓
Runtime de Agente
    ↓
┌─────────────┬──────────────┬─────────────┐
│ Observatorio│ Barandillas  │ Coordinador │
│ (Monitoreo) │ (Seguridad)  │ (Sincroniz.)│
└─────────────┴──────────────┴─────────────┘
    ↓
Cola de Tareas → API de Proveedores
```

## Configuración

### Habilitar Infraestructura de Agentes

```json
{
  "agent": {
    "enabled": true,
    "runtime": {
      "max_concurrent_tasks": 5,
      "task_timeout": "30m",
      "auto_cleanup": true
    },
    "observatory": {
      "enabled": true,
      "update_interval": "5s",
      "history_retention": "7d"
    },
    "guardrails": {
      "enabled": true,
      "max_file_operations": 100,
      "max_api_calls": 1000,
      "allowed_paths": ["/Users/john/projects"],
      "blocked_commands": ["rm -rf", "sudo"]
    },
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    }
  }
}
```

## Componentes

### 1. Runtime de Agente

Gestiona el ciclo de vida de ejecución de tareas de agentes.

**Características:**
- Programación y ejecución de tareas
- Gestión de tareas concurrentes
- Manejo de timeouts
- Limpieza automática
- Recuperación de errores

**Configuración:**
```json
{
  "runtime": {
    "max_concurrent_tasks": 5,
    "task_timeout": "30m",
    "auto_cleanup": true,
    "retry_failed_tasks": true,
    "max_retries": 3
  }
}
```

**API:**
```bash
# Iniciar tarea de agente
POST /api/v1/agent/tasks
Content-Type: application/json

{
  "name": "code-review",
  "description": "Review pull request #123",
  "priority": 1,
  "config": {
    "model": "claude-opus-4",
    "max_tokens": 100000
  }
}

# Obtener estado de tarea
GET /api/v1/agent/tasks/{task_id}

# Cancelar tarea
DELETE /api/v1/agent/tasks/{task_id}
```

### 2. Observatorio

Monitoreo en tiempo real de actividad de agentes.

**Características:**
- Seguimiento de sesiones
- Registro de actividad
- Métricas de rendimiento
- Actualizaciones de estado
- Datos históricos

**Configuración:**
```json
{
  "observatory": {
    "enabled": true,
    "update_interval": "5s",
    "history_retention": "7d",
    "metrics": {
      "track_tokens": true,
      "track_costs": true,
      "track_latency": true
    }
  }
}
```

**Métricas de Monitoreo:**
- Sesiones activas
- Tareas en progreso
- Uso de tokens
- Llamadas API
- Operaciones de archivos
- Tasa de errores
- Latencia promedio

**API:**
```bash
# Obtener todas las sesiones activas
GET /api/v1/agent/sessions

# Obtener detalles de sesión
GET /api/v1/agent/sessions/{session_id}

# Obtener métricas de sesión
GET /api/v1/agent/sessions/{session_id}/metrics
```

### 3. Barandillas

Controles de seguridad y restricciones para el comportamiento de agentes.

**Características:**
- Límites de operaciones
- Restricciones de rutas
- Bloqueo de comandos
- Cuotas de recursos
- Flujos de trabajo de aprobación

**Configuración:**
```json
{
  "guardrails": {
    "enabled": true,
    "max_file_operations": 100,
    "max_api_calls": 1000,
    "max_tokens_per_session": 1000000,
    "allowed_paths": [
      "/Users/john/projects",
      "/tmp/agent-workspace"
    ],
    "blocked_paths": [
      "/etc",
      "/System",
      "~/.ssh"
    ],
    "blocked_commands": [
      "rm -rf /",
      "sudo",
      "chmod 777"
    ],
    "require_approval": {
      "file_delete": true,
      "system_commands": true,
      "network_requests": false
    }
  }
}
```

**Mecanismo de Aplicación:**
- Validación previa a la ejecución
- Monitoreo en tiempo real
- Bloqueo automático
- Prompts de aprobación
- Registro de auditoría

**API:**
```bash
# Obtener estado de barandillas
GET /api/v1/agent/guardrails

# Actualizar reglas de barandillas
PUT /api/v1/agent/guardrails
Content-Type: application/json

{
  "max_file_operations": 200,
  "blocked_commands": ["rm -rf", "sudo", "dd"]
}
```

### 4. Coordinador

Coordinación de flujos de trabajo multi-agente basada en archivos.

**Características:**
- Bloqueo de archivos
- Detección de cambios
- Resolución de conflictos
- Sincronización de estado
- Notificaciones de eventos

**Configuración:**
```json
{
  "coordinator": {
    "enabled": true,
    "lock_timeout": "5m",
    "change_detection": true,
    "conflict_resolution": "last-write-wins",
    "notification_webhook": "https://hooks.slack.com/..."
  }
}
```

**Casos de Uso:**
- Múltiples agentes editando los mismos archivos
- Prevenir modificaciones concurrentes
- Detectar cambios externos de archivos
- Coordinar flujos de trabajo de agentes

**API:**
```bash
# Adquirir bloqueo de archivo
POST /api/v1/agent/locks
Content-Type: application/json

{
  "path": "/path/to/file.go",
  "session_id": "sess_123",
  "timeout": "5m"
}

# Liberar bloqueo de archivo
DELETE /api/v1/agent/locks/{lock_id}

# Obtener eventos de cambio de archivo
GET /api/v1/agent/changes?since=2026-03-05T10:00:00Z
```

### 5. Cola de Tareas

Gestiona tareas de agentes con prioridades y dependencias.

**Características:**
- Programación por prioridad
- Dependencias de tareas
- Gestión de cola
- Seguimiento de estado
- Lógica de reintentos

**Configuración:**
```json
{
  "task_queue": {
    "enabled": true,
    "max_queue_size": 100,
    "priority_levels": 5,
    "enable_dependencies": true,
    "retry_policy": {
      "max_retries": 3,
      "backoff": "exponential"
    }
  }
}
```

**API:**
```bash
# Agregar tarea a la cola
POST /api/v1/agent/queue
Content-Type: application/json

{
  "name": "run-tests",
  "priority": 2,
  "depends_on": ["build-project"],
  "config": {}
}

# Obtener estado de la cola
GET /api/v1/agent/queue

# Eliminar tarea de la cola
DELETE /api/v1/agent/queue/{task_id}
```

## Web UI

Accede al panel de agentes en `http://localhost:19840/agent`

### Pestaña de Sesiones

- **Sesiones Activas** — Sesiones de agentes actualmente en ejecución
- **Detalles de Sesión** — Progreso de tareas, métricas, logs
- **Control de Sesión** — Pausar, reanudar, cancelar

### Pestaña de Tareas

- **Cola de Tareas** — Tareas pendientes y en progreso
- **Historial de Tareas** — Tareas completadas y fallidas
- **Detalles de Tarea** — Configuración, logs, resultados

### Pestaña de Barandillas

- **Límites de Operaciones** — Uso actual vs. límites
- **Operaciones Bloqueadas** — Intentos bloqueados recientes
- **Cola de Aprobación** — Operaciones esperando aprobación

### Pestaña de Métricas

- **Uso de Tokens** — Por sesión y total
- **Llamadas API** — Conteo de solicitudes y tasa
- **Operaciones de Archivos** — Conteos de lectura/escritura/eliminación
- **Rendimiento** — Latencia y throughput

## Integración con Claude Code

GoZen detecta automáticamente sesiones de Claude Code y proporciona infraestructura de agentes:

```bash
# Iniciar Claude Code con soporte de agentes
zen --agent

# Las funciones de agente se habilitan automáticamente:
# - Seguimiento de sesiones
# - Coordinación de archivos
# - Aplicación de barandillas
# - Monitoreo en tiempo real
```

**Beneficios:**
- Prevenir modificaciones concurrentes de archivos
- Rastrear uso de tokens y costos
- Aplicar restricciones de seguridad
- Monitorear actividad de agentes
- Coordinar flujos de trabajo multi-agente

## Casos de Uso

### Desarrollo Multi-Agente

Múltiples agentes trabajando en la misma base de código:

```json
{
  "agent": {
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    },
    "guardrails": {
      "max_file_operations": 200,
      "allowed_paths": ["/Users/john/project"]
    }
  }
}
```

### Tareas de Larga Duración

Monitorear y controlar tareas de agentes de larga duración:

```json
{
  "agent": {
    "runtime": {
      "task_timeout": "2h",
      "auto_cleanup": false
    },
    "observatory": {
      "update_interval": "10s",
      "history_retention": "30d"
    }
  }
}
```

### Operaciones Críticas de Seguridad

Aplicar controles de seguridad estrictos:

```json
{
  "agent": {
    "guardrails": {
      "enabled": true,
      "max_file_operations": 50,
      "blocked_commands": ["rm", "sudo", "chmod"],
      "require_approval": {
        "file_delete": true,
        "system_commands": true,
        "network_requests": true
      }
    }
  }
}
```

## Mejores Prácticas

1. **Habilitar Barandillas** — Siempre usar barandillas en producción
2. **Establecer Límites Apropiados** — Configurar límites basados en el caso de uso
3. **Monitorear Activamente** — Revisar regularmente el panel del observatorio
4. **Usar Bloqueo de Archivos** — Habilitar coordinador para flujos de trabajo multi-agente
5. **Configurar Aprobaciones** — Requerir aprobación para operaciones destructivas
6. **Revisar Logs** — Auditar regularmente la actividad de agentes

## Limitaciones

1. **Sobrecarga de Rendimiento** — El monitoreo y coordinación agregan latencia
2. **Bloqueo de Archivos** — Puede causar retrasos en escenarios multi-agente
3. **Uso de Memoria** — El historial de sesiones consume memoria
4. **Complejidad** — Requiere comprensión de flujos de trabajo de agentes
5. **Estado Beta** — Las funciones pueden cambiar en versiones futuras

## Solución de Problemas

### Las Sesiones de Agentes No Se Rastrean

1. Verifica que `agent.enabled` sea `true`
2. Revisa que el observatorio esté habilitado
3. Asegúrate de que el cliente de agente sea compatible (Claude Code, Codex)
4. Revisa los logs del daemon para errores

### Problemas de Bloqueo de Archivos

1. Revisa que el coordinador esté habilitado
2. Valida que el timeout de bloqueo sea apropiado
3. Revisa los bloqueos activos: `GET /api/v1/agent/locks`
4. Libera manualmente bloqueos atascados si es necesario

### Las Barandillas No Se Aplican

1. Verifica que las barandillas estén habilitadas
2. Revisa que la configuración de reglas sea correcta
3. Revisa los logs de operaciones bloqueadas
4. Asegúrate de que el cliente de agente respete las barandillas

### Alto Uso de Memoria

1. Reduce el período de retención del historial
2. Disminuye el intervalo de actualización
3. Limita el número máximo de tareas concurrentes
4. Habilita la limpieza automática

## Consideraciones de Seguridad

1. **Restricciones de Rutas** — Siempre configura rutas permitidas/bloqueadas
2. **Bloqueo de Comandos** — Bloquea comandos peligrosos
3. **Flujos de Trabajo de Aprobación** — Requiere aprobación para operaciones sensibles
4. **Registro de Auditoría** — Habilita registro completo
5. **Límites de Recursos** — Establece límites apropiados de operaciones

## Mejoras Futuras

- Protocolos de colaboración multi-agente
- Estrategias avanzadas de resolución de conflictos
- Machine learning para detección de anomalías
- Integración con herramientas de monitoreo externas
- Análisis de comportamiento de agentes
- Generación automática de políticas de seguridad
