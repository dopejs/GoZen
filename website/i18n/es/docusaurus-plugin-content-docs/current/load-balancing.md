---
title: Balanceo de Carga
---

# Balanceo de Carga

GoZen admite varias estrategias de selección de proveedores además del failover básico. Puedes elegir una estrategia por perfil, combinarla con verificaciones de salud y dirigir tráfico según disponibilidad, latencia o costo.

## Estrategias disponibles

### Failover

Prueba los proveedores en orden hasta que uno funcione. Es la estrategia predeterminada y encaja bien en configuraciones primario/respaldo.

```json
{
  "profiles": {
    "default": {
      "providers": ["primary", "backup"],
      "strategy": "failover"
    }
  }
}
```

### Round robin

Distribuye las solicitudes de manera uniforme entre varios proveedores equivalentes.

```json
{
  "profiles": {
    "balanced": {
      "providers": ["provider-a", "provider-b", "provider-c"],
      "strategy": "round-robin"
    }
  }
}
```

### Menor latencia

Prefiere el proveedor con el tiempo de respuesta reciente más bajo.

```json
{
  "profiles": {
    "fast": {
      "providers": ["us-east", "us-west", "eu"],
      "strategy": "least-latency"
    }
  }
}
```

### Menor costo

Prefiere el proveedor más barato para el modelo solicitado.

```json
{
  "profiles": {
    "budget": {
      "providers": ["cheap-provider", "premium-provider"],
      "strategy": "least-cost"
    }
  }
}
```

## Enrutamiento consciente de salud

Todas las estrategias pueden trabajar junto con el monitoreo de salud. Cuando `health_aware` está activado, los proveedores no saludables se omiten automáticamente hasta que se recuperen.

```json
{
  "profiles": {
    "production": {
      "providers": ["primary", "secondary", "tertiary"],
      "strategy": "least-latency",
      "health_aware": true
    }
  }
}
```

## Cómo elegir una estrategia

- Usa `failover` cuando priorices confiabilidad.
- Usa `round-robin` cuando los proveedores sean intercambiables.
- Usa `least-latency` para cargas interactivas o sensibles al tiempo.
- Usa `least-cost` cuando el presupuesto importe más que la velocidad.

## Documentos relacionados

- [Perfiles](/docs/profiles) explica cómo se definen los grupos de proveedores.
- [Enrutamiento](/docs/routing) cubre la selección de proveedores basada en escenarios.
- [Monitoreo de Salud](/docs/health-monitoring) explica cómo las verificaciones de salud afectan al enrutamiento.
