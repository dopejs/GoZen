---
sidebar_position: 4
title: Enrutamiento por Escenario
---

# Enrutamiento por Escenario

Enruta automáticamente las solicitudes a diferentes proveedores según las características de la solicitud.

## Escenarios Soportados

| Escenario | Descripción |
|-----------|-------------|
| `think` | Modo de pensamiento habilitado |
| `image` | Contiene contenido de imagen |
| `longContext` | El contenido excede el umbral |
| `webSearch` | Usa la herramienta web_search |
| `background` | Usa el modelo Haiku |

## Mecanismo de Respaldo

Si todos los proveedores de un escenario fallan, se recurre automáticamente a los proveedores predeterminados del perfil.

## Ejemplo de Configuración

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
