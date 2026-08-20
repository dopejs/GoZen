---
sidebar_position: 5
title: Vinculaciones de Proyecto
---

# Vinculaciones de Proyecto

Vincula directorios a perfiles y/o CLIs específicos para configuración automática a nivel de proyecto.

## Uso

```bash
cd ~/work/company-project

# Bind profile
zen bind work-profile

# Bind CLI
zen bind --cli codex

# Bind both
zen bind work-profile --cli codex

# Check status
zen status

# Unbind
zen unbind
```

## Prioridad

Argumentos CLI > Vinculaciones de proyecto > Valores predeterminados globales
