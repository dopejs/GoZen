---
sidebar_position: 14
title: Compresión de Contexto (BETA)
---

# Compresión de Contexto (BETA)

:::warning FUNCIÓN BETA
La compresión de contexto está actualmente en beta. Está deshabilitada por defecto y requiere configuración explícita para habilitarla.
:::

Comprime automáticamente el contexto de conversación cuando el conteo de tokens excede un umbral, reduciendo costos mientras mantiene la calidad de la conversación.

## Características

- **Compresión Automática** — Se activa cuando el conteo de tokens excede el umbral
- **Resumen Inteligente** — Usa un modelo económico (claude-3-haiku) para resumir mensajes antiguos
- **Preservar Mensajes Recientes** — Mantiene los mensajes más recientes intactos para continuidad del contexto
- **Estimación de Tokens** — Cálculo preciso de tokens antes de llamadas API
- **Seguimiento de Estadísticas** — Monitorea la efectividad de la compresión
- **Operación Transparente** — Funciona sin problemas con todos los clientes de AI

## Cómo Funciona

1. **Estimación de Tokens** — Calcula el conteo de tokens en el historial de conversación
2. **Verificación de Umbral** — Compara con el umbral configurado (predeterminado: 50,000)
3. **Selección de Mensajes** — Identifica mensajes antiguos para compresión
4. **Generación de Resumen** — Usa un modelo económico para crear un resumen conciso
5. **Reemplazo de Contexto** — Reemplaza mensajes antiguos con el resumen
6. **Reenvío de Solicitud** — Envía el contexto comprimido al modelo objetivo

## Configuración

### Habilitar Compresión

```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 50000,
    "target_tokens": 20000,
    "summarizer_model": "claude-3-haiku-20240307",
    "preserve_recent_messages": 5,
    "tokens_per_char": 0.25
  }
}
```

**Opciones:**

| Opción | Predeterminado | Descripción |
|------|--------|------|
| `enabled` | `false` | Habilita la compresión de contexto |
| `threshold_tokens` | `50000` | Activa la compresión cuando el contexto excede esto |
| `target_tokens` | `20000` | Conteo de tokens objetivo después de la compresión |
| `summarizer_model` | `claude-3-haiku-20240307` | Modelo usado para resumir |
| `preserve_recent_messages` | `5` | Número de mensajes recientes a mantener intactos |
| `tokens_per_char` | `0.25` | Ratio de estimación para conteo de tokens |

### Configuración por Perfil

Habilita la compresión para perfiles específicos:

```json
{
  "profiles": {
    "long-context": {
      "providers": ["anthropic"],
      "compression": {
        "enabled": true,
        "threshold_tokens": 100000,
        "target_tokens": 40000
      }
    },
    "short-context": {
      "providers": ["openai"],
      "compression": {
        "enabled": false
      }
    }
  }
}
```

## Estimación de Tokens

GoZen usa estimación basada en caracteres para conteo rápido de tokens:

```
estimated_tokens = character_count * tokens_per_char
```

**Ratio Predeterminado:** 0.25 tokens por carácter (1 token ≈ 4 caracteres)

**Precisión:** ±10% para texto en inglés, puede variar para otros idiomas

Para conteo preciso de tokens, GoZen usa la biblioteca `tiktoken-go` cuando está disponible.

## Estrategia de Compresión

### Selección de Mensajes

1. **Mensajes del Sistema** — Siempre preservados
2. **Mensajes Recientes** — Preserva los últimos N mensajes (predeterminado: 5)
3. **Mensajes Antiguos** — Candidatos para compresión

### Prompt de Resumen

```
Resume concisamente el siguiente historial de conversación mientras preservas información clave, decisiones y contexto:

[Mensajes antiguos]

Proporciona un resumen breve que capture los puntos esenciales.
```

### Resultado

```
Original: 45,000 tokens (30 mensajes)
Comprimido: 22,000 tokens (resumen + 5 mensajes recientes)
Ahorro: 23,000 tokens (51%)
```

## Web UI

Accede a la configuración de compresión en `http://localhost:19840/settings`:

1. Navega a la pestaña "Compression" (marcada con insignia BETA)
2. Activa "Enable Compression"
3. Ajusta los umbrales y conteos de tokens objetivo
4. Selecciona el modelo de resumen
5. Establece el número de mensajes recientes a preservar
6. Haz clic en "Save"

### Panel de Estadísticas

Ver estadísticas de compresión:

- **Total de Compresiones** — Número de veces que se activó la compresión
- **Tokens Ahorrados** — Total de tokens ahorrados en todas las compresiones
- **Ahorro Promedio** — Reducción promedio de tokens por compresión
- **Tasa de Compresión** — Porcentaje de solicitudes que activaron compresión

## Endpoints de API

### Obtener Estadísticas de Compresión

```bash
GET /api/v1/compression/stats
```

Respuesta:
```json
{
  "enabled": true,
  "total_compressions": 42,
  "tokens_saved": 1250000,
  "average_savings": 29761,
  "compression_rate": 0.15,
  "last_compression": "2026-03-05T10:30:00Z"
}
```

### Actualizar Configuración de Compresión

```bash
PUT /api/v1/compression/settings
Content-Type: application/json

{
  "enabled": true,
  "threshold_tokens": 60000,
  "target_tokens": 25000
}
```

### Reiniciar Estadísticas

```bash
POST /api/v1/compression/stats/reset
```

## Casos de Uso

### Sesiones de Codificación Largas

**Escenario:** Sesiones de codificación de múltiples horas con Claude Code

**Configuración:**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 80000,
    "target_tokens": 30000,
    "preserve_recent_messages": 10
  }
}
```

**Beneficio:** Mantiene la continuidad de la conversación sin alcanzar límites de contexto

### Procesamiento por Lotes

**Escenario:** Procesamiento de múltiples documentos con AI

**Configuración:**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 40000,
    "target_tokens": 15000,
    "preserve_recent_messages": 3
  }
}
```

**Beneficio:** Reduce costos al procesar grandes conjuntos de documentos

### Investigación y Análisis

**Escenario:** Sesiones de investigación largas que abarcan múltiples temas

**Configuración:**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 100000,
    "target_tokens": 40000,
    "preserve_recent_messages": 8
  }
}
```

**Beneficio:** Mantiene la conversación enfocada en temas recientes mientras preserva contexto temprano

## Mejores Prácticas

1. **Comenzar con Predeterminados** — La configuración predeterminada funciona para la mayoría de los casos de uso
2. **Monitorear Estadísticas** — Revisa regularmente la tasa de compresión y ahorros
3. **Ajustar Umbrales** — Aumentar para modelos de contexto largo (Claude Opus), disminuir para contexto corto
4. **Preservar Suficientes Mensajes** — Mantén 5-10 mensajes recientes para continuidad del contexto
5. **Usar Resumidor Económico** — Haiku es rápido y rentable para resúmenes
6. **Probar Antes de Producción** — Valida la calidad de compresión con tu caso de uso específico

## Limitaciones

1. **Pérdida de Calidad** — Los resúmenes pueden perder detalles sutiles
2. **Latencia Aumentada** — Agrega sobrecarga de llamada API de resumen
3. **Compensación de Costos** — Costo de resumen vs. ahorro de tokens
4. **Soporte de Idiomas** — Funciona mejor para inglés, puede variar para otros idiomas
5. **Ventana de Contexto** — No puede exceder la ventana de contexto máxima del modelo

## Solución de Problemas

### La Compresión No Se Activa

1. Verifica que `compression.enabled` sea `true`
2. Revisa si el conteo de tokens excede el umbral
3. Asegúrate de que la conversación tenga suficientes mensajes para comprimir
4. Revisa los logs del daemon para errores de compresión

### Calidad de Resumen Deficiente

1. Prueba un modelo de resumen diferente (ej: claude-3-sonnet)
2. Aumenta `preserve_recent_messages` para retener más contexto
3. Ajusta `target_tokens` para permitir resúmenes más largos
4. Verifica que el modelo de resumen esté disponible y funcionando

### Latencia Aumentada

1. La compresión agrega una llamada API extra (resumen)
2. Usa un modelo de resumen más rápido (haiku es el más rápido)
3. Aumenta el umbral para reducir la frecuencia de compresión
4. Considera deshabilitar la compresión para aplicaciones sensibles a latencia

### Costos Inesperados

1. Monitorea el costo de resumen en el panel de uso
2. Compara ahorros vs. costo de resumen
3. Ajusta el umbral para reducir la frecuencia de compresión
4. Usa el modelo más económico disponible para resumen

## Impacto en el Rendimiento

- **Estimación de Tokens** — ~1ms por solicitud (despreciable)
- **Generación de Resumen** — 1-3 segundos (depende del modelo y número de mensajes)
- **Sobrecarga de Memoria** — Mínima (~1KB por compresión)
- **Ahorro de Costos** — Típicamente reduce 30-50% de tokens

## Configuración Avanzada

### Prompt de Resumen Personalizado

```json
{
  "compression": {
    "enabled": true,
    "custom_prompt": "Crea un resumen técnico de la siguiente conversación, enfocándote en cambios de código, decisiones y elementos de acción:\n\n{messages}\n\nResumen:"
  }
}
```

### Compresión Condicional

Habilita la compresión solo para escenarios específicos:

```json
{
  "profiles": {
    "default": {
      "scenarios": {
        "longContext": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": true,
            "threshold_tokens": 100000
          }
        },
        "default": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": false
          }
        }
      }
    }
  }
}
```

### Compresión Multi-Etapa

Múltiples compresiones para conversaciones muy largas:

```json
{
  "compression": {
    "enabled": true,
    "stages": [
      {
        "threshold_tokens": 50000,
        "target_tokens": 30000
      },
      {
        "threshold_tokens": 80000,
        "target_tokens": 40000
      }
    ]
  }
}
```

## Mejoras Futuras

- Coincidencia de similitud semántica para selección inteligente de mensajes
- Resumen multi-modelo para comparación de calidad
- Métricas de calidad de compresión y retroalimentación
- Estrategias de compresión personalizadas por caso de uso
- Integración con RAG para almacenamiento de contexto externo
