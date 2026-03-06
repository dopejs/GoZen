---
sidebar_position: 13
title: Webhooks
---

# Webhooks

Recibe notificaciones en tiempo real de alertas de presupuesto, cambios de estado de proveedores y resúmenes diarios a través de Slack, Discord o webhooks personalizados.

## Características

- **Múltiples Formatos** — Slack, Discord o JSON genérico
- **Filtrado de Eventos** — Suscríbete a tipos de eventos específicos
- **Encabezados Personalizados** — Agrega autenticación o encabezados personalizados
- **Entrega Asíncrona** — Entrega de webhooks sin bloqueo
- **Formato Automático** — Mensajes enriquecidos con emojis y colores
- **Función de Prueba** — Valida la configuración del webhook antes de habilitar

## Configuración

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": [
        "budget_warning",
        "budget_exceeded",
        "provider_down",
        "provider_up",
        "failover",
        "daily_summary"
      ],
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  ]
}
```

## Tipos de Eventos

| Evento | Descripción | Cuándo se Activa |
|------|------|----------|
| `budget_warning` | Umbral de presupuesto alcanzado | Cuando el gasto alcanza el 80% del límite |
| `budget_exceeded` | Límite de presupuesto excedido | Cuando el gasto supera el límite configurado |
| `provider_down` | El proveedor se vuelve no saludable | Cuando la tasa de éxito cae por debajo del 70% |
| `provider_up` | El proveedor se recupera | Cuando un proveedor no saludable vuelve a estar saludable |
| `failover` | Failover de solicitud | Cuando una solicitud cambia a proveedor de respaldo |
| `daily_summary` | Resumen de uso diario | Una vez al día a medianoche UTC |

## Formatos de Webhook

### Slack

Detectado automáticamente cuando la URL contiene `slack.com`.

**Mensaje de Ejemplo:**
```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

**Formato:**
```json
{
  "text": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)"
      }
    }
  ]
}
```

### Discord

Detectado automáticamente cuando la URL contiene `discord.com`.

**Embed de Ejemplo:**
- **Título:** budget_warning
- **Descripción:** ⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
- **Color:** Ámbar (#FBBF24)
- **Timestamp:** 2026-03-05T10:30:00Z

**Formato:**
```json
{
  "content": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
  "embeds": [
    {
      "title": "budget_warning",
      "description": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
      "timestamp": "2026-03-05T10:30:00Z",
      "color": 16432932
    }
  ]
}
```

### JSON Genérico

Usado para todas las demás URLs.

**Formato:**
```json
{
  "event": "budget_warning",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "period": "daily",
    "spent": 8.5,
    "limit": 10.0,
    "percent": 85.0,
    "project": ""
  }
}
```

## Estructuras de Datos de Eventos

### Advertencia / Exceso de Presupuesto

```json
{
  "event": "budget_warning",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "period": "daily",
    "spent": 8.5,
    "limit": 10.0,
    "percent": 85.0,
    "action": "warn",
    "project": "my-project"
  }
}
```

### Proveedor Caído / Recuperado

```json
{
  "event": "provider_down",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "provider": "anthropic-primary",
    "status": "unhealthy",
    "error": "connection timeout",
    "latency_ms": 0
  }
}
```

### Failover

```json
{
  "event": "failover",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "from_provider": "anthropic-primary",
    "to_provider": "anthropic-backup",
    "reason": "rate limit exceeded",
    "session_id": "sess_abc123"
  }
}
```

### Resumen Diario

```json
{
  "event": "daily_summary",
  "timestamp": "2026-03-05T00:00:00Z",
  "data": {
    "date": "2026-03-04",
    "total_cost": 25.50,
    "total_requests": 150,
    "total_input_tokens": 125000,
    "total_output_tokens": 35000,
    "by_provider": {
      "anthropic": 18.20,
      "openai": 7.30
    }
  }
}
```

## Configuración de Plataformas

### Slack

1. Visita [Slack API](https://api.slack.com/apps)
2. Crea una nueva aplicación o selecciona una existente
3. Habilita "Incoming Webhooks"
4. Agrega el webhook al workspace
5. Copia la URL del webhook (comienza con `https://hooks.slack.com/`)

**Configuración:**
```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
      "events": ["budget_warning", "provider_down"]
    }
  ]
}
```

### Discord

1. Abre la configuración del servidor de Discord
2. Ve a Integrations → Webhooks
3. Haz clic en "New Webhook"
4. Selecciona el canal y copia la URL del webhook

**Configuración:**
```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://discord.com/api/webhooks/123456789/XXXXXXXXXXXXXXXXXXXX",
      "events": ["budget_exceeded", "failover"]
    }
  ]
}
```

### Webhook Personalizado

Para integraciones personalizadas, usa el formato JSON genérico:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://your-server.com/webhook",
      "events": ["budget_warning", "daily_summary"],
      "headers": {
        "Authorization": "Bearer YOUR_SECRET_TOKEN",
        "X-Custom-Header": "value"
      }
    }
  ]
}
```

## Configuración de Web UI

Accede a la configuración de webhooks en `http://localhost:19840/settings`:

1. Navega a la pestaña "Webhooks"
2. Haz clic en "Add Webhook"
3. Ingresa la URL del webhook
4. Selecciona los eventos a los que suscribirse
5. (Opcional) Agrega encabezados personalizados
6. Haz clic en "Test" para validar la configuración
7. Haz clic en "Save"

## Endpoints de API

### Listar Webhooks

```bash
GET /api/v1/webhooks
```

### Agregar Webhook

```bash
POST /api/v1/webhooks
Content-Type: application/json

{
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "events": ["budget_warning", "provider_down"]
}
```

### Actualizar Webhook

```bash
PUT /api/v1/webhooks/{id}
Content-Type: application/json

{
  "enabled": false
}
```

### Eliminar Webhook

```bash
DELETE /api/v1/webhooks/{id}
```

### Probar Webhook

```bash
POST /api/v1/webhooks/{id}/test
```

Envía un mensaje de prueba para validar la configuración.

## Ejemplos de Mensajes

### Advertencia de Presupuesto (Slack)

```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

### Presupuesto Excedido (Discord)

```
🚫 Budget Exceeded: monthly limit of $200.00 reached (spent: $205.50). Action: block
```

### Proveedor Caído (Slack)

```
🔴 Provider Down: anthropic-primary is unhealthy. Error: connection timeout
```

### Proveedor Recuperado (Discord)

```
🟢 Provider Up: anthropic-primary is healthy again (latency: 1250ms)
```

### Failover (Slack)

```
🔄 Failover: Switched from anthropic-primary to anthropic-backup. Reason: rate limit exceeded
```

### Resumen Diario (Discord)

```
📊 Daily Summary (2026-03-04): 150 requests, $25.50 total cost, 125000 input / 35000 output tokens
```

## Mejores Prácticas

1. **Usar Webhooks Separados** — Crea diferentes webhooks para diferentes tipos de eventos
2. **Probar Antes de Habilitar** — Siempre prueba la configuración del webhook antes de guardar
3. **Asegurar Webhooks Personalizados** — Usa HTTPS y encabezados de autenticación
4. **Monitorear Fallos de Webhook** — Revisa los logs del daemon si las notificaciones se detienen
5. **Evitar Datos Sensibles** — No incluyas claves API o tokens en URLs de webhook
6. **Configurar Alertas** — Suscríbete a eventos críticos como `budget_exceeded` y `provider_down`

## Solución de Problemas

### El Webhook No Recibe Mensajes

1. Verifica que el webhook esté habilitado en la configuración
2. Revisa que la URL sea correcta (prueba con curl)
3. Valida que la configuración de eventos sea correcta
4. Revisa los logs del daemon para errores de webhook: `tail -f ~/.zen/zend.log`
5. Prueba el webhook mediante API: `POST /api/v1/webhooks/{id}/test`

### Fallo de Webhook de Slack

1. Verifica que la URL del webhook comience con `https://hooks.slack.com/`
2. Revisa que el webhook no haya sido revocado en la configuración de Slack
3. Asegúrate de que el workspace no haya deshabilitado los webhooks entrantes
4. Prueba con curl:
   ```bash
   curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

### Fallo de Webhook de Discord

1. Verifica que la URL del webhook comience con `https://discord.com/api/webhooks/`
2. Revisa que el webhook no haya sido eliminado en la configuración de Discord
3. Asegúrate de que el bot tenga permisos para publicar en el canal
4. Prueba con curl:
   ```bash
   curl -X POST https://discord.com/api/webhooks/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"content":"test"}'
   ```

### Webhook Personalizado No Funciona

1. Verifica que el endpoint sea accesible (prueba con curl)
2. Revisa que los encabezados de autenticación sean correctos
3. Asegúrate de que el endpoint acepte solicitudes POST
4. Valida que el endpoint devuelva un código de estado 2xx
5. Revisa los logs del endpoint para errores

## Consideraciones de Seguridad

1. **Proteger URLs de Webhook** — Trata las URLs de webhook como secretos
2. **Usar HTTPS** — Siempre usa HTTPS para endpoints de webhook
3. **Verificar Firmas** — Implementa verificación de firmas para webhooks personalizados
4. **Límite de Tasa** — Implementa límite de tasa en endpoints de webhook
5. **No Registrar Datos Sensibles** — Evita registrar payloads completos de webhook

## Configuración Avanzada

### Webhooks Condicionales

Envía diferentes eventos a diferentes webhooks:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/CRITICAL/ALERTS",
      "events": ["budget_exceeded", "provider_down"]
    },
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/DAILY/REPORTS",
      "events": ["daily_summary"]
    },
    {
      "enabled": true,
      "url": "https://discord.com/api/webhooks/MONITORING",
      "events": ["failover", "provider_up"]
    }
  ]
}
```

### Encabezados Personalizados para Autenticación

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://your-server.com/webhook",
      "events": ["budget_warning"],
      "headers": {
        "Authorization": "Bearer YOUR_SECRET_TOKEN",
        "X-API-Key": "your-api-key",
        "X-Webhook-Source": "gozen"
      }
    }
  ]
}
```
