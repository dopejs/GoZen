---
sidebar_position: 3
title: Perfiles y Conmutación
---

# Perfiles y Conmutación

Un perfil es una lista ordenada de proveedores para conmutación automática. Cuando el primer proveedor no está disponible, cambia automáticamente al siguiente.

## Ejemplo de Configuración

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

## Uso de Perfiles

```bash
# Use default profile
zen

# Use specified profile
zen -p work

# Interactively select
zen -p
```
