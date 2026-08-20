---
sidebar_position: 12
title: Monitoreo de Salud y Balanceo de Carga
---

# Monitoreo de Salud y Balanceo de Carga

Monitorea la salud de los proveedores en tiempo real y enruta automáticamente las solicitudes al mejor proveedor disponible.

## Características

- **Verificaciones de Salud en Tiempo Real** — Monitoreo periódico de salud con intervalo de verificación configurable
- **Seguimiento de Tasa de Éxito** — Calcula la salud del proveedor basándose en la tasa de éxito de solicitudes
- **Monitoreo de Latencia** — Rastrea el tiempo de respuesta promedio por proveedor
- **Múltiples Estrategias** — Failover, round-robin, menor latencia, menor costo
- **Failover Automático** — Cambia a proveedores de respaldo cuando el primario no está saludable
- **Panel de Salud** — Indicadores de estado visuales en la Web UI

## Configuración

### Habilitar Monitoreo de Salud

```json
{
  "health_check": {
    "enabled": true,
    "interval": "5m",
    "timeout": "10s",
    "endpoint": "/v1/messages",
    "method": "POST"
  }
}
```

**Opciones:**
- `interval` — Frecuencia de verificación de salud del proveedor (predeterminado: 5 minutos)
- `timeout` — Timeout de solicitud para verificaciones de salud (predeterminado: 10 segundos)
- `endpoint` — Endpoint de API a probar (predeterminado: `/v1/messages`)
- `method` — Método HTTP para verificaciones de salud (predeterminado: `POST`)

### Configurar Balanceo de Carga

```json
{
  "load_balancing": {
    "strategy": "least-latency",
    "health_aware": true,
    "cache_ttl": "30s"
  }
}
```

## Estrategias de Balanceo de Carga

### 1. Failover (Predeterminado)

Usa proveedores en orden, cambiando al siguiente en caso de fallo.

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-primary", "anthropic-backup", "openai"],
      "load_balancing": {
        "strategy": "failover"
      }
    }
  }
}
```

**Comportamiento:**
1. Intenta `anthropic-primary`
2. Si falla, intenta `anthropic-backup`
3. Si falla, intenta `openai`
4. Si todos fallan, devuelve error

**Mejor para:** Cargas de trabajo de producción con jerarquía clara primario/respaldo

### 2. Round-Robin

Distribuye solicitudes uniformemente entre todos los proveedores saludables.

```json
{
  "load_balancing": {
    "strategy": "round-robin"
  }
}
```

**Comportamiento:**
- Solicitud 1 → Proveedor A
- Solicitud 2 → Proveedor B
- Solicitud 3 → Proveedor C
- Solicitud 4 → Proveedor A (repite el ciclo)

**Mejor para:** Distribuir carga entre múltiples cuentas para evitar límites de tasa

### 3. Menor Latencia

Enruta al proveedor con la latencia promedio más baja.

```json
{
  "load_balancing": {
    "strategy": "least-latency"
  }
}
```

**Comportamiento:**
- Rastrea el tiempo de respuesta promedio de cada proveedor
- Enruta al proveedor más rápido
- Actualiza métricas cada 30 segundos (configurable mediante `cache_ttl`)

**Mejor para:** Aplicaciones sensibles a latencia, interacciones en tiempo real

### 4. Menor Costo

Enruta al proveedor más económico para el modelo solicitado.

```json
{
  "load_balancing": {
    "strategy": "least-cost"
  }
}
```

**Comportamiento:**
- Compara precios entre proveedores
- Enruta a la opción más económica
- Considera costos de tokens de entrada y salida

**Mejor para:** Optimización de costos, procesamiento por lotes

## Estados de Salud

Los proveedores se clasifican en cuatro estados de salud:

| Estado | Tasa de Éxito | Comportamiento |
|------|--------|------|
| **Saludable** | ≥ 95% | Prioridad normal |
| **Degradado** | 70-95% | Prioridad más baja, aún utilizable |
| **No Saludable** | < 70% | Omitido a menos que no haya proveedores saludables |
| **Desconocido** | Sin datos | Tratado como saludable inicialmente |

### Enrutamiento Consciente de Salud

Cuando `health_aware: true` (predeterminado):
- Los proveedores saludables tienen prioridad
- Los proveedores degradados se usan como respaldo
- Los proveedores no saludables se omiten a menos que todos los demás fallen

## Panel de Web UI

Accede al panel de salud en `http://localhost:19840/health`:

### Estado de Proveedores

- **Indicador de Estado** — Verde (saludable), amarillo (degradado), rojo (no saludable)
- **Tasa de Éxito** — Porcentaje de solicitudes exitosas
- **Latencia Promedio** — Tiempo de respuesta promedio en milisegundos
- **Última Verificación** — Timestamp de la verificación de salud más reciente
- **Conteo de Errores** — Número de fallos recientes

### Línea de Tiempo de Métricas

- **Gráfico de Latencia** — Tendencia del tiempo de respuesta a lo largo del tiempo
- **Gráfico de Tasa de Éxito** — Tendencia de salud a lo largo del tiempo
- **Volumen de Solicitudes** — Conteo de solicitudes por proveedor

## Endpoints de API

### Obtener Salud de Proveedores

```bash
GET /api/v1/health/providers
```

Respuesta:
```json
{
  "providers": [
    {
      "name": "anthropic-primary",
      "status": "healthy",
      "success_rate": 98.5,
      "avg_latency_ms": 1250,
      "last_check": "2026-03-05T10:30:00Z",
      "error_count": 2,
      "total_requests": 150
    },
    {
      "name": "openai-backup",
      "status": "degraded",
      "success_rate": 85.0,
      "avg_latency_ms": 2100,
      "last_check": "2026-03-05T10:29:00Z",
      "error_count": 15,
      "total_requests": 100
    }
  ]
}
```

### Obtener Métricas de Proveedor

```bash
GET /api/v1/health/providers/{name}/metrics?period=1h
```

Respuesta:
```json
{
  "provider": "anthropic-primary",
  "period": "1h",
  "metrics": [
    {
      "timestamp": "2026-03-05T10:00:00Z",
      "latency_ms": 1200,
      "success_rate": 99.0,
      "requests": 25
    },
    {
      "timestamp": "2026-03-05T10:05:00Z",
      "latency_ms": 1300,
      "success_rate": 98.0,
      "requests": 28
    }
  ]
}
```

### Activar Verificación de Salud Manual

```bash
POST /api/v1/health/check
Content-Type: application/json

{
  "provider": "anthropic-primary"
}
```

## Notificaciones Webhook

Recibe alertas cuando cambia el estado del proveedor:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": ["provider_down", "provider_up", "failover"]
    }
  ]
}
```

**Tipos de Eventos:**
- `provider_down` — El proveedor se vuelve no saludable
- `provider_up` — El proveedor se recupera a estado saludable
- `failover` — La solicitud hace failover a proveedor de respaldo

## Enrutamiento Basado en Escenarios

Combina el monitoreo de salud con enrutamiento de escenarios para distribución inteligente de solicitudes:

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-primary", "anthropic-backup"],
      "scenarios": {
        "thinking": {
          "providers": ["anthropic-thinking"],
          "load_balancing": {
            "strategy": "least-latency"
          }
        },
        "image": {
          "providers": ["anthropic-vision", "openai-vision"],
          "load_balancing": {
            "strategy": "failover"
          }
        }
      }
    }
  }
}
```

Ver [Enrutamiento de Escenarios](./routing.md).

## Mejores Prácticas

1. **Establecer Intervalos Apropiados** — 5 minutos funciona para la mayoría, usa 1 minuto para sistemas críticos
2. **Usar Enrutamiento Consciente de Salud** — Siempre habilitar para cargas de trabajo de producción
3. **Monitorear Proveedores Degradados** — Investigar cuando la tasa de éxito cae por debajo del 95%
4. **Combinar Estrategias** — Usar failover para primario/respaldo, round-robin para distribución de carga
5. **Habilitar Webhooks** — Recibir notificaciones inmediatas cuando los proveedores caen
6. **Revisar Panel Regularmente** — Observar tendencias de salud para identificar patrones

## Solución de Problemas

### Las Verificaciones de Salud Fallan

1. Verifica que las claves API del proveedor sean válidas
2. Revisa la conectividad de red al endpoint del proveedor
3. Aumenta el timeout si el proveedor es lento: `"timeout": "30s"`
4. Revisa los logs del daemon para mensajes de error específicos

### Métricas de Latencia Incorrectas

1. La latencia incluye tiempo de red + tiempo de procesamiento de API
2. Revisa si proxies o VPN agregan sobrecarga
3. Las métricas se almacenan en caché por 30 segundos por defecto (configurable mediante `cache_ttl`)

### El Failover No Funciona

1. Verifica que `health_aware: true` en la configuración de balanceo de carga
2. Revisa que los proveedores de respaldo estén configurados en el perfil
3. Asegúrate de que las verificaciones de salud estén habilitadas y ejecutándose
4. Busca eventos de failover en la Web UI o logs

### El Proveedor Está Atascado en Estado No Saludable

1. Activa manualmente una verificación de salud mediante API
2. Verifica si el proveedor realmente está caído (prueba con curl)
3. Reinicia el daemon para restablecer el estado de salud: `zen daemon restart`
4. Revisa los logs de errores para encontrar la causa raíz

## Impacto en el Rendimiento

- **Verificaciones de Salud** — Sobrecarga mínima, se ejecutan en goroutines en segundo plano
- **Caché de Métricas** — TTL de 30 segundos reduce consultas a la base de datos
- **Operaciones Atómicas** — Contadores thread-safe para solicitudes concurrentes
- **Sin Bloqueo** — Las verificaciones de salud no bloquean el procesamiento de solicitudes

## Configuración Avanzada

### Payload de Verificación de Salud Personalizado

```json
{
  "health_check": {
    "enabled": true,
    "custom_payload": {
      "model": "claude-3-haiku-20240307",
      "max_tokens": 10,
      "messages": [
        {
          "role": "user",
          "content": "ping"
        }
      ]
    }
  }
}
```

### Configuración de Salud por Proveedor

```json
{
  "providers": {
    "anthropic-primary": {
      "health_check": {
        "interval": "1m",
        "timeout": "5s"
      }
    },
    "openai-backup": {
      "health_check": {
        "interval": "5m",
        "timeout": "10s"
      }
    }
  }
}
```
