---
sidebar_position: 11
title: Seguimiento de Uso y Control de Presupuesto
---

# Seguimiento de Uso y Control de Presupuesto

Rastrea el uso de tokens y costos a través de proveedores, modelos y proyectos. Establece límites de gasto y aplica acciones automáticamente.

## Características

- **Seguimiento en Tiempo Real** — Monitorea el uso de tokens y costos por solicitud
- **Agregación Multidimensional** — Rastrea por proveedor, modelo, proyecto y período de tiempo
- **Límites de Presupuesto** — Establece límites de gasto diarios, semanales y mensuales
- **Acciones Automáticas** — Advierte, degrada o bloquea solicitudes cuando se exceden los límites
- **Estimación de Costos** — Precios precisos para todos los modelos de AI principales
- **Datos Históricos** — Almacenamiento SQLite con agregación por hora para rendimiento

## Configuración

### Habilitar Seguimiento de Uso

```json
{
  "usage_tracking": {
    "enabled": true,
    "db_path": "~/.zen/usage.db"
  }
}
```

### Configurar Precios de Modelos

```json
{
  "pricing": {
    "models": {
      "claude-opus-4": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet-4": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4o": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    },
    "model_families": {
      "claude-opus": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    }
  }
}
```

**Coincidencia de Modelos**: Primero coincide con nombres de modelo exactos, luego recurre a prefijos de familia de modelos.

### Establecer Límites de Presupuesto

```json
{
  "budget": {
    "daily": {
      "enabled": true,
      "limit": 10.0,
      "action": "warn"
    },
    "weekly": {
      "enabled": true,
      "limit": 50.0,
      "action": "downgrade"
    },
    "monthly": {
      "enabled": true,
      "limit": 200.0,
      "action": "block"
    }
  }
}
```

## Acciones de Presupuesto

| Acción | Comportamiento |
|------|------|
| `warn` | Registra advertencia y envía notificación webhook, pero permite la solicitud |
| `downgrade` | Cambia a un modelo más económico (ej: opus → sonnet → haiku) |
| `block` | Rechaza la solicitud con código de estado 429 |

## Web UI

Accede al panel de uso en `http://localhost:19840/usage`:

- **Resumen** — Costo total, solicitudes y tokens para el período actual
- **Por Proveedor** — Desglose de costos por cada proveedor
- **Por Modelo** — Estadísticas de uso por cada modelo
- **Por Proyecto** — Rastrea costos por proyecto (mediante vinculaciones de proyecto)
- **Línea de Tiempo** — Tendencias de costos por hora/día
- **Estado de Presupuesto** — Indicadores visuales para límites diarios/semanales/mensuales

## Endpoints de API

### Obtener Resumen de Uso

```bash
GET /api/v1/usage/summary?period=daily
```

Respuesta:
```json
{
  "period": "daily",
  "start": "2026-03-05T00:00:00Z",
  "end": "2026-03-05T23:59:59Z",
  "total_cost": 8.45,
  "total_requests": 42,
  "total_input_tokens": 125000,
  "total_output_tokens": 35000,
  "by_provider": {
    "anthropic": 6.20,
    "openai": 2.25
  },
  "by_model": {
    "claude-sonnet-4": 5.10,
    "claude-opus-4": 1.10,
    "gpt-4o": 2.25
  }
}
```

### Obtener Estado de Presupuesto

```bash
GET /api/v1/budget/status
```

Respuesta:
```json
{
  "daily": {
    "enabled": true,
    "limit": 10.0,
    "spent": 8.45,
    "percent": 84.5,
    "action": "warn",
    "exceeded": false
  },
  "weekly": {
    "enabled": true,
    "limit": 50.0,
    "spent": 32.10,
    "percent": 64.2,
    "action": "downgrade",
    "exceeded": false
  },
  "monthly": {
    "enabled": true,
    "limit": 200.0,
    "spent": 145.80,
    "percent": 72.9,
    "action": "block",
    "exceeded": false
  }
}
```

### Actualizar Límites de Presupuesto

```bash
PUT /api/v1/budget/limits
Content-Type: application/json

{
  "daily": {
    "enabled": true,
    "limit": 15.0,
    "action": "warn"
  }
}
```

## Seguimiento a Nivel de Proyecto

Rastrea costos por proyecto usando vinculaciones de directorio:

```bash
# Vincula el directorio actual a un perfil
zen bind work-profile

# Todas las solicitudes desde este directorio se etiquetan con la ruta del proyecto
# Ver costos bajo "By Project" en la Web UI
```

## Notificaciones Webhook

Recibe alertas cuando se exceden los presupuestos:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": ["budget_warning", "budget_exceeded"]
    }
  ]
}
```

Ver [Webhooks](./webhooks.md) para configuración completa.

## Mejores Prácticas

1. **Comenzar con Advertencias** — Usa la acción `warn` inicialmente para entender los patrones de uso
2. **Establecer Límites Realistas** — Basa los límites en datos de uso histórico
3. **Usar Degradación en Desarrollo** — Cambia automáticamente a modelos más económicos al probar
4. **Reservar Bloqueo para Producción** — Usa la acción `block` solo para límites de gasto estrictos
5. **Monitorear Diariamente** — Revisa regularmente el panel de uso para evitar sorpresas
6. **Habilitar Webhooks** — Obtén alertas en tiempo real cuando te acerques a los límites

## Solución de Problemas

### El Uso No Se Rastrea

1. Verifica que `usage_tracking.enabled` sea `true` en la configuración
2. Revisa que la ruta de la base de datos sea escribible: `~/.zen/usage.db`
3. Reinicia el daemon: `zen daemon restart`

### Costos Incorrectos

1. Verifica que los precios de los modelos en la configuración coincidan con las tarifas actuales
2. Revisa la coincidencia de nombres de modelo (coincidencia exacta vs. prefijo de familia)
3. Actualiza la configuración de precios si los proveedores cambian las tarifas

### El Presupuesto No Se Aplica

1. Revisa que la configuración de presupuesto esté habilitada
2. Verifica que la acción esté establecida (`warn`, `downgrade` o `block`)
3. Revisa los logs del daemon para errores del verificador de presupuesto

## Rendimiento

- **Agregación por Hora** — Los datos sin procesar se agregan por hora para reducir la carga de consultas
- **Consultas Indexadas** — La base de datos indexa proveedor, modelo, proyecto, timestamp
- **Almacenamiento Eficiente** — ~1KB por solicitud, ~30MB para 30,000 solicitudes
- **Panel Rápido** — Consultas en menos de un segundo para patrones de uso típicos
