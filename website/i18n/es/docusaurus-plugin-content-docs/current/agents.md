---
title: Agentes
---

# Agentes

GoZen puede actuar como una capa operativa para agentes de programación como Claude Code, Codex y otros asistentes basados en CLI. Te ayuda a coordinar el trabajo de los agentes, monitorear sesiones y aplicar controles de seguridad en tiempo de ejecución sin cambiar tu flujo actual.

## Lo que agrega GoZen

- **Coordinación**: Reduce conflictos cuando varios agentes trabajan sobre el mismo proyecto.
- **Observabilidad**: Centraliza sesiones, costos, errores y actividad.
- **Guardrails**: Aplica límites de gasto, tasa de solicitudes y acciones sensibles.
- **Enrutamiento de tareas**: Envía distintos tipos de trabajo a diferentes proveedores o perfiles.

## Ejemplo de configuración

```json
{
  "agent": {
    "enabled": true,
    "coordinator": {
      "enabled": true,
      "lock_timeout_sec": 300,
      "inject_warnings": true
    },
    "observatory": {
      "enabled": true,
      "stuck_threshold": 5,
      "idle_timeout_min": 30
    },
    "guardrails": {
      "enabled": true,
      "session_spending_cap": 5.0,
      "request_rate_limit": 30
    }
  }
}
```

## Flujos de trabajo comunes

### Coordinación multiagente

Cuando varios agentes trabajan en el mismo repositorio, GoZen puede rastrear la actividad de archivos, mostrar advertencias y ayudar a evitar colisiones.

### Monitoreo de sesiones

Usa el panel y las APIs para inspeccionar sesiones activas, uso de tokens, número de errores y tiempo de ejecución.

### Aplicación de seguridad

Los guardrails pueden pausar sesiones fuera de control, señalar operaciones riesgosas y frenar bucles de reintento antes de que se vuelvan costosos.

## Documentos relacionados

- [Infraestructura de Agentes](/docs/agent-infrastructure) cubre con más detalle la nueva arquitectura de runtime, observatorio, coordinador y guardrails.
- [Gateway de Bot](/docs/bot) explica cómo controlar sesiones en ejecución desde Telegram, Slack, Discord y otras plataformas de chat.
- [Seguimiento de Uso](/docs/usage-tracking) y [Monitoreo de Salud](/docs/health-monitoring) cubren las métricas que impulsan la operación de agentes.
